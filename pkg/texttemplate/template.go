// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package texttemplate

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/k14s/ytt/pkg/template"
)

type Template struct {
	name string
}

func NewTemplate(name string) *Template {
	return &Template{name: name}
}

func (e *Template) CompileInline(rootNode *NodeRoot,
	instructions *template.InstructionSet, nodes *template.Nodes) ([]template.Line, error) {

	return e.compile(rootNode, instructions, nodes)
}

func (e *Template) Compile(rootNode *NodeRoot) (*template.CompiledTemplate, error) {
	instructions := template.NewInstructionSet()
	nodes := template.NewNodes()

	code, err := e.compile(rootNode, instructions, nodes)
	if err != nil {
		return nil, err
	}

	return template.NewCompiledTemplate(
		e.name, code, instructions, nodes,
		template.EvaluationCtxDialects{
			EvaluationCtxDialectName: EvaluationCtx{},
		},
	), nil
}

func (e *Template) compile(rootNode *NodeRoot,
	instructions *template.InstructionSet, nodes *template.Nodes) ([]template.Line, error) {

	code := []template.Line{}
	rootNodeTag := nodes.AddRootNode(&NodeRoot{}) // fresh copy to avoid leaking out NodeCode

	code = append(code, []template.Line{
		{Instruction: instructions.NewSetCtxType(EvaluationCtxDialectName)},
		{Instruction: instructions.NewStartCtx(EvaluationCtxDialectName)},
		{Instruction: instructions.NewStartNode(rootNodeTag)},
	}...)

	var trimSpaceRight bool

	for i, node := range rootNode.Items {
		switch typedNode := node.(type) {
		case *NodeText:
			if trimSpaceRight {
				typedNode.Content = strings.TrimLeftFunc(typedNode.Content, unicode.IsSpace)
				trimSpaceRight = false
			}

			nodeTag := nodes.AddNode(typedNode, rootNodeTag)

			code = append(code, template.Line{
				Instruction: instructions.NewStartNode(nodeTag),
				SourceLine:  template.NewSourceLine(typedNode.Position, typedNode.Content),
			})

			code = append(code, template.Line{
				Instruction: instructions.NewSetNode(nodeTag),
				SourceLine:  template.NewSourceLine(typedNode.Position, typedNode.Content),
			})

		case *NodeCode:
			meta := NodeCodeMeta{typedNode}
			trimSpaceRight = meta.ShouldTrimSpaceRight()

			if meta.ShoudTrimSpaceLeft() && i != 0 {
				if typedLastNode, ok := rootNode.Items[i-1].(*NodeText); ok {
					typedLastNode.Content = strings.TrimRightFunc(
						typedLastNode.Content, unicode.IsSpace)
				}
			}

			if meta.ShouldPrint() {
				nodeTag := nodes.AddNode(&NodeText{}, rootNodeTag)

				code = append(code, template.Line{
					Instruction: instructions.NewStartNode(nodeTag),
					SourceLine:  template.NewSourceLine(typedNode.Position, typedNode.Content),
				})

				code = append(code, template.Line{
					Instruction: instructions.NewSetNodeValue(nodeTag, meta.Code()),
					SourceLine:  template.NewSourceLine(typedNode.Position, typedNode.Content),
				})
			} else {
				code = append(code, template.NewCodeFromBytesAtPosition(
					[]byte(meta.Code()), typedNode.Position, instructions)...)
			}

		default:
			panic(fmt.Sprintf("unknown string template node %T", typedNode))
		}
	}

	code = append(code, template.Line{
		Instruction: instructions.NewEndCtx(),
	})

	return code, nil
}
