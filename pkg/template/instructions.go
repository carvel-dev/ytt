package template

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/structmeta"
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

var (
	globalInsSetId = 1
)

func NewInstructionSet() *InstructionSet {
	globalInsSetId += 1
	uniqueId := globalInsSetId
	return &InstructionSet{
		SetCtxType:            InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_ctx_type", uniqueId)},
		StartCtx:              InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_ctx", uniqueId)},
		EndCtx:                InstructionOp{fmt.Sprintf("__ytt_tpl%d_end_ctx", uniqueId)},
		StartNodeAnnotation:   InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_node_annotation", uniqueId)},
		CollectNodeAnnotation: InstructionOp{fmt.Sprintf("__ytt_tpl%d_collect_node_annotation", uniqueId)},
		StartNode:             InstructionOp{fmt.Sprintf("__ytt_tpl%d_start_node", uniqueId)},
		SetNode:               InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_node", uniqueId)},
		SetMapItemKey:         InstructionOp{fmt.Sprintf("__ytt_tpl%d_set_map_item_key", uniqueId)},
		ReplaceNode:           InstructionOp{fmt.Sprintf("__ytt_tpl%d_replace_node", uniqueId)},
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

func (is *InstructionSet) NewStartNodeAnnotation(nodeTag NodeTag, ann structmeta.Annotation) Instruction {
	collectedArgs := is.CollectNodeAnnotation.WithArgs(ann.Content).AsString()
	return is.StartNodeAnnotation.WithArgs(nodeTag.AsString(), `"`+string(ann.Name)+`"`, collectedArgs)
}

func (is *InstructionSet) NewStartNode(nodeTag NodeTag) Instruction {
	return is.StartNode.WithArgs(nodeTag.AsString())
}

func (is *InstructionSet) NewSetNode(nodeTag NodeTag) Instruction {
	return is.SetNode.WithArgs(nodeTag.AsString())
}

func (is *InstructionSet) NewSetNodeValue(nodeTag NodeTag, code string) Instruction {
	return is.SetNode.WithArgs(nodeTag.AsString(), "("+code+")")
}

func (is *InstructionSet) NewSetMapItemKey(nodeTag NodeTag, code string) Instruction {
	return is.SetMapItemKey.WithArgs(nodeTag.AsString(), "("+code+")")
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
