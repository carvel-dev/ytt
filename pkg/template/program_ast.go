// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"

	"github.com/k14s/starlark-go/syntax"
)

type ProgramAST struct {
	f                 *syntax.File
	defNestingUsesTpl []bool
	ctxType           *syntax.Literal
	instructions      *InstructionSet
}

func NewProgramAST(f *syntax.File, instructions *InstructionSet) *ProgramAST {
	return &ProgramAST{f: f, instructions: instructions}
}

func (r *ProgramAST) InsertTplCtxs() {
	r.stmts(r.f.Stmts)
}

func (r *ProgramAST) stmts(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		r.stmt(stmt)
	}
}

func (r *ProgramAST) stmt(stmt syntax.Stmt) {
	switch stmt := stmt.(type) {
	case *syntax.ExprStmt:
		r.expr(stmt.X)

	case *syntax.BranchStmt:
		// do nothing

	case *syntax.IfStmt:
		r.expr(stmt.Cond)
		r.stmts(stmt.True)
		r.stmts(stmt.False)

	case *syntax.AssignStmt:
		r.expr(stmt.RHS)

	case *syntax.DefStmt:
		r.function(stmt.Def, stmt.Name.Name, stmt)

	case *syntax.ForStmt:
		r.expr(stmt.X)
		r.stmts(stmt.Body)

	case *syntax.WhileStmt:
		r.expr(stmt.Cond)
		r.stmts(stmt.Body)

	case *syntax.ReturnStmt:
		if len(r.defNestingUsesTpl) > 0 {
			if r.defNestingUsesTpl[len(r.defNestingUsesTpl)-1] {
				args := []syntax.Expr{}
				if stmt.Result != nil {
					args = []syntax.Expr{stmt.Result}
				}
				stmt.Result = &syntax.CallExpr{
					Fn:   &syntax.Ident{Name: r.instructions.EndCtx.Name},
					Args: args,
				}
			}
		} else {
			if stmt.Result != nil {
				r.expr(stmt.Result)
			}
		}

	case *syntax.LoadStmt:
		// do nothing

	default:
		panic(fmt.Sprintf("unexpected stmt %T", stmt))
	}
}

func (r *ProgramAST) expr(e syntax.Expr) {
	switch e := e.(type) {
	case *syntax.Ident:
		if e.Name == r.instructions.StartNode.Name || e.Name == r.instructions.SetNode.Name {
			if len(r.defNestingUsesTpl) > 0 {
				r.defNestingUsesTpl[len(r.defNestingUsesTpl)-1] = true
			}
		}

	case *syntax.Literal:
		// do nothing

	case *syntax.ListExpr:
		for _, x := range e.List {
			r.expr(x)
		}

	case *syntax.CondExpr:
		r.expr(e.Cond)
		r.expr(e.True)
		r.expr(e.False)

	case *syntax.IndexExpr:
		r.expr(e.X)
		r.expr(e.Y)

	case *syntax.DictEntry:
		r.expr(e.Key)
		r.expr(e.Value)

	case *syntax.SliceExpr:
		r.expr(e.X)
		if e.Lo != nil {
			r.expr(e.Lo)
		}
		if e.Hi != nil {
			r.expr(e.Hi)
		}
		if e.Step != nil {
			r.expr(e.Step)
		}

	case *syntax.Comprehension:
		// The 'in' operand of the first clause (always a ForClause)
		// is resolved in the outer block; consider: [x for x in x].
		clause := e.Clauses[0].(*syntax.ForClause)
		r.expr(clause.X)

		for _, clause := range e.Clauses[1:] {
			switch clause := clause.(type) {
			case *syntax.IfClause:
				r.expr(clause.Cond)
			case *syntax.ForClause:
				r.expr(clause.X)
			}
		}
		r.expr(e.Body)

	case *syntax.TupleExpr:
		for _, x := range e.List {
			r.expr(x)
		}

	case *syntax.DictExpr:
		for _, entry := range e.List {
			entry := entry.(*syntax.DictEntry)
			r.expr(entry.Key)
			r.expr(entry.Value)
		}

	case *syntax.UnaryExpr:
		r.expr(e.X)

	case *syntax.BinaryExpr:
		r.expr(e.X)
		r.expr(e.Y)

	case *syntax.DotExpr:
		r.expr(e.X)

	case *syntax.CallExpr:
		if ident, ok := e.Fn.(*syntax.Ident); ok {
			if ident.Name == r.instructions.SetCtxType.Name {
				r.ctxType = e.Args[0].(*syntax.Literal)
			}
		}
		r.expr(e.Fn)
		for _, arg := range e.Args {
			if unop, ok := arg.(*syntax.UnaryExpr); ok && unop.Op == syntax.STARSTAR {
				r.expr(arg)
			} else if ok && unop.Op == syntax.STAR {
				r.expr(arg)
			} else if binop, ok := arg.(*syntax.BinaryExpr); ok && binop.Op == syntax.EQ {
				r.expr(binop.Y)
			} else {
				r.expr(arg)
			}
		}

	case *syntax.LambdaExpr:
		r.expr(e.Body)

	case *syntax.ParenExpr:
		r.expr(e.X)

	default:
		panic(fmt.Sprintf("unexpected expr %T", e))
	}
}

func (r *ProgramAST) function(pos syntax.Position, name string, function *syntax.DefStmt) {
	r.defNestingUsesTpl = append(r.defNestingUsesTpl, false)

	for _, param := range function.Params {
		if binary, ok := param.(*syntax.BinaryExpr); ok {
			r.expr(binary.Y)
		}
	}

	r.stmts(function.Body)

	if r.defNestingUsesTpl[len(r.defNestingUsesTpl)-1] {
		r.addTplCtxToFunction(function)
	}

	r.defNestingUsesTpl = r.defNestingUsesTpl[:len(r.defNestingUsesTpl)-1]
}

func (r *ProgramAST) addTplCtxToFunction(function *syntax.DefStmt) {
	if r.ctxType == nil {
		panic("expected r.ctxType to be set")
	}

	startStmt := &syntax.ExprStmt{
		X: &syntax.CallExpr{
			Fn:   &syntax.Ident{Name: r.instructions.StartCtx.Name},
			Args: []syntax.Expr{r.ctxType},
		},
	}

	endStmt := &syntax.ReturnStmt{
		Result: &syntax.CallExpr{
			Fn: &syntax.Ident{Name: r.instructions.EndCtx.Name},
		},
	}

	function.Body = append(append([]syntax.Stmt{startStmt}, function.Body...), endStmt)
}
