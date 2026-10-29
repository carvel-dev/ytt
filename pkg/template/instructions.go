// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type InstructionSet struct {
	SetCtxType            InstructionOp
	StartCtx              InstructionOp
	EndCtx                InstructionOp
	StartNodeAnnotation   InstructionOp
	CollectNodeAnnotation InstructionOp
	StartNode             InstructionOp
	SetNode               InstructionOp
	SetMapItemKey         InstructionOp
	ReplaceNode           InstructionOp
}

// globalInsSetID names each InstructionSet apart from every other one, so that
// the Starlark identifiers a template compiles down to cannot collide with
// those of a template it loads.
//
// This is to protect concurrent usage when consumed as a Go module. In the
// worse case scenario this could be a race that ends up with two templates
// having the same ID.
var globalInsSetID atomic.Int64

func NewInstructionSet() *InstructionSet {
	uniqueID := globalInsSetID.Add(1)
	return &InstructionSet{
		SetCtxType:            InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_ctx_type", uniqueID)},
		StartCtx:              InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_ctx", uniqueID)},
		EndCtx:                InstructionOp{fmt.Sprintf("__ytt_tpl%d_end_ctx", uniqueID)},
		StartNodeAnnotation:   InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_node_annotation", uniqueID)},
		CollectNodeAnnotation: InstructionOp{fmt.Sprintf("__ytt_tpl%d_collect_node_annotation", uniqueID)},
		StartNode:             InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_node", uniqueID)},
		SetNode:               InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_node", uniqueID)},
		SetMapItemKey:         InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_map_item_key", uniqueID)},
		ReplaceNode:           InstructionOp{fmt.Sprintf("__ytt_tpl%d_replace_node", uniqueID)},
	}
}

func (is *InstructionSet) NewSetCtxType(dialect EvaluationCtxDialectName) Instruction {
	return is.SetCtxType.WithArgs(`"` + string(dialect) + `"`)
}

func (is *InstructionSet) NewStartCtx(dialect EvaluationCtxDialectName) Instruction {
	return is.StartCtx.WithArgs(`"` + string(dialect) + `"`)
}

func (is *InstructionSet) NewEndCtx() Instruction {
	return is.EndCtx.WithArgs()
}

func (is *InstructionSet) NewEndCtxNone() Instruction {
	return is.EndCtx.WithArgs("None")
}

func (is *InstructionSet) NewStartNodeAnnotation(nodeTag NodeTag, ann Annotation) Instruction {
	collectedArgs := is.CollectNodeAnnotation.WithArgs(
		withoutComments(ann.Content)).AsString()

	return is.StartNodeAnnotation.WithArgs(nodeTag.AsString(), `"`+string(ann.Name)+`"`, collectedArgs)
}

func (is *InstructionSet) NewStartNode(nodeTag NodeTag) Instruction {
	return is.StartNode.WithArgs(nodeTag.AsString())
}

func (is *InstructionSet) NewSetNode(nodeTag NodeTag) Instruction {
	return is.SetNode.WithArgs(nodeTag.AsString())
}

func (is *InstructionSet) NewSetNodeValue(nodeTag NodeTag, code string) Instruction {
	return is.SetNode.WithArgs(nodeTag.AsString(), parenthesized(code))
}

func (is *InstructionSet) NewSetMapItemKey(nodeTag NodeTag, code string) Instruction {
	return is.SetMapItemKey.WithArgs(nodeTag.AsString(), parenthesized(code))
}

// tripleQuoteLen is the length of a Starlark triple-quote delimiter.
const tripleQuoteLen = 3

// parenthesized wraps code in parentheses, dropping its comments first so
// that a trailing comment such as `#@ value #! note` does not also comment
// out the closing parenthesis.
func parenthesized(code string) string {
	return "(" + withoutComments(code) + ")"
}

// withoutComments removes Starlark comments (an unquoted '#' up to the end
// of its line) from code and keeps string literals as they are.
func withoutComments(code string) string {
	var out strings.Builder
	for i := 0; i < len(code); i++ {
		switch c := code[i]; {
		case c == '\'' || c == '"':
			i = copyStringLiteral(&out, code, i)
		case c == '#':
			// Starlark ends a comment at a line feed or a carriage return.
			end := strings.IndexAny(code[i:], "\r\n")
			if end < 0 {
				return out.String()
			}
			i += end - 1
		default:
			out.WriteByte(c)
		}
	}
	return out.String()
}

// copyStringLiteral writes the string literal that opens at code[start] to
// out and returns the index of its last byte. An unterminated literal runs
// to the end of code.
func copyStringLiteral(out *strings.Builder, code string, start int) int {
	quote := code[start : start+1]
	if triple := strings.Repeat(quote, tripleQuoteLen); strings.HasPrefix(
		code[start:], triple) {
		quote = triple
	}
	for i := start + len(quote); i < len(code); i++ {
		if code[i] == '\\' {
			i++ // the escaped byte cannot close the literal
			continue
		}
		if strings.HasPrefix(code[i:], quote) {
			end := i + len(quote)
			out.WriteString(code[start:end])
			return end - 1
		}
	}
	out.WriteString(code[start:])
	return len(code) - 1
}

func (is *InstructionSet) NewCode(code string) Instruction {
	return Instruction{code: code}
}

type InstructionOp struct {
	Name string
}

func (op InstructionOp) WithArgs(args ...string) Instruction {
	return Instruction{op: op, code: fmt.Sprintf("%s(%s)", op.Name, strings.Join(args, ", "))}
}

type Instruction struct {
	op   InstructionOp
	code string
}

func (i Instruction) Op() InstructionOp { return i.op }
func (i Instruction) AsString() string  { return i.code }

func (i Instruction) WithDebug(info string) Instruction {
	return Instruction{op: i.op, code: i.code + " # " + info}
}
