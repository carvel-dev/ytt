package template

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type EvaluationCtxDialectName string
type EvaluationCtxDialects map[EvaluationCtxDialectName]EvaluationCtxDialect

type CompiledTemplate struct {
	name         string
	code         []TemplateLine
	instructions *InstructionSet
	nodes        *Nodes
	evalDialects EvaluationCtxDialects
	rootCtx      *EvaluationCtx
	ctxs         []*EvaluationCtx
	annotations  []GlobalAnnotation
}

func NewCompiledTemplate(name string, code []TemplateLine,
	instructions *InstructionSet, nodes *Nodes,
	evalDialects EvaluationCtxDialects) *CompiledTemplate {

	// TODO package globals
	resolve.AllowFloat = true
	resolve.AllowSet = true
	resolve.AllowLambda = true
	resolve.AllowNestedDef = true
	resolve.AllowBitwise = true
	resolve.AllowRecursion = true
	resolve.AllowGlobalReassign = true

	return &CompiledTemplate{
		name:         name,
		code:         code,
		instructions: instructions,
		nodes:        nodes,
		evalDialects: evalDialects,
	}
}

func (e *CompiledTemplate) Code() []TemplateLine { return e.code }

func (e *CompiledTemplate) CodeAtLine(pos *filepos.Position) *TemplateLine {
	for i, line := range e.code {
		if i+1 == pos.Line() {
			return &line
		}
	}
	return nil
}

func (e *CompiledTemplate) CodeAsString() string {
	result := []string{}
	for _, line := range e.code {
		result = append(result, line.Instruction.AsString())
	}
	// Do not add any unnecessary newlines to match code lines
	return strings.Join(result, "\n")
}

func (e *CompiledTemplate) DebugCodeAsString() string {
	result := []string{"src:  tmpl: code"}
	for i, line := range e.code {
		result = append(result, fmt.Sprintf("%s: %4d: %s",
			line.Position().As4DigitString(), i+1, line.Instruction.AsString()))
	}
	// Do not add any unnecessary newlines to match code lines
	return strings.Join(result, "\n")
}

func (e *CompiledTemplate) Eval(thread *starlark.Thread, loader CompiledTemplateLoader) (
	starlark.StringDict, interface{}, []GlobalAnnotation, error) {

	globals := make(starlark.StringDict)

	if e.nodes != nil {
		instructionBindings := map[string]tplcore.StarlarkFunc{
			// TODO ProgramAST should get rid of set ctx type calls
			e.instructions.SetCtxType.Name:            e.tplSetCtxType,
			e.instructions.StartCtx.Name:              e.tplStartCtx,
			e.instructions.EndCtx.Name:                e.tplEndCtx,
			e.instructions.StartNodeAnnotation.Name:   e.tplStartNodeAnnotation,
			e.instructions.CollectNodeAnnotation.Name: e.tplCollectNodeAnnotation,
			e.instructions.StartNode.Name:             e.tplStartNode,
			e.instructions.SetNode.Name:               e.tplSetNode,
		}

		for name, f := range instructionBindings {
			globals[name] = starlark.NewBuiltin(name, tplcore.ErrWrapper(f))
		}
	}

	updatedGlobals, val, err := e.eval(thread, globals)
	if err != nil {
		return nil, nil, nil, NewCompiledTemplateMultiError(err, loader)
	}

	return updatedGlobals, val, e.annotations, nil
}

func (e *CompiledTemplate) eval(
	thread *starlark.Thread, globals starlark.StringDict) (
	gs starlark.StringDict, resultVal interface{}, resultErr error) {

	// Catch any panics to give a better contextual information
	defer func() {
		if err := recover(); err != nil {
			if typedErr, ok := err.(error); ok {
				resultErr = typedErr
			} else {
				resultErr = fmt.Errorf("(p) %s", err)
			}
		}
	}()

	f, err := syntax.Parse(e.name, e.CodeAsString(), 0)
	if err != nil {
		return nil, nil, err
	}

	NewProgramAST(f, e.instructions).InsertTplCtxs()

	prog, err := starlark.FileProgram(f, globals.Has)
	if err != nil {
		return nil, nil, err
	}

	// clear before execution
	e.rootCtx = nil
	e.ctxs = nil

	updatedGlobals, err := prog.Init(thread, globals)
	if err != nil {
		return nil, nil, err
	}

	updatedGlobals.Freeze()

	if len(e.ctxs) > 0 {
		panic("expected all ctxs to end")
	}

	// Plain starlark programs do not have any ctxs
	if e.rootCtx != nil {
		resultVal = e.rootCtx.RootNode()
	}

	return updatedGlobals, resultVal, nil
}

func (e *CompiledTemplate) newCtx(ctxType EvaluationCtxDialectName) *EvaluationCtx {
	return &EvaluationCtx{
		nodes:     e.nodes,
		ancestors: e.nodes.Ancestors(),
		dialect:   e.evalDialects[ctxType],

		pendingAnnotations: map[NodeTag]NodeAnnotations{},
	}
}

func (e *CompiledTemplate) tplSetCtxType(
	thread *starlark.Thread, _ *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return starlark.None, nil
}

func (e *CompiledTemplate) tplStartCtx(
	thread *starlark.Thread, _ *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	ctxType, err := tplcore.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	e.ctxs = append(e.ctxs, e.newCtx(EvaluationCtxDialectName(ctxType)))

	if len(e.ctxs) == 1 && e.rootCtx == nil {
		e.rootCtx = e.ctxs[0]
	}

	return starlark.None, nil
}

func (e *CompiledTemplate) tplEndCtx(
	thread *starlark.Thread, _ *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if len(e.ctxs) == 0 {
		panic("unexpected ctx end")
	}

	var returnVal starlark.Value
	switch args.Len() {
	case 0:
		returnVal = e.ctxs[len(e.ctxs)-1].RootNodeAsStarlarkValue()
	case 1:
		returnVal = args.Index(0)
	default:
		return starlark.None, fmt.Errorf("expected zero or one argument")
	}

	e.ctxs = e.ctxs[:len(e.ctxs)-1]
	return returnVal, nil
}

func (e *CompiledTemplate) tplSetNode(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return e.ctxs[len(e.ctxs)-1].TplSetNode(thread, f, args, kwargs)
}

func (e *CompiledTemplate) tplStartNodeAnnotation(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return e.ctxs[len(e.ctxs)-1].TplStartNodeAnnotation(e, thread, f, args, kwargs)
}

func (e *CompiledTemplate) tplCollectNodeAnnotation(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return e.ctxs[len(e.ctxs)-1].TplCollectNodeAnnotation(thread, f, args, kwargs)
}

func (e *CompiledTemplate) tplStartNode(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return e.ctxs[len(e.ctxs)-1].TplStartNode(thread, f, args, kwargs)
}

func (e *CompiledTemplate) TplReplaceNode(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return e.ctxs[len(e.ctxs)-1].TplReplace(thread, f, args, kwargs)
}
