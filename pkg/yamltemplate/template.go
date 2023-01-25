// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"bytes"
	"fmt"

	"github.com/carvel-dev/ytt/pkg/filepos"
	"github.com/carvel-dev/ytt/pkg/template"
	"github.com/carvel-dev/ytt/pkg/texttemplate"
	"github.com/carvel-dev/ytt/pkg/yamlmeta"
)

const (
	AnnotationTextTemplatedStrings template.AnnotationName = "yaml/text-templated-strings"
)

type Template struct {
	name         string
	opts         TemplateOpts
	docSet       *yamlmeta.DocumentSet
	nodes        *template.Nodes
	instructions *template.InstructionSet

	// memoized source lines
	srcLinesByLine map[int]string
}

type TemplateOpts struct {
	IgnoreUnknownComments   bool
	ImplicitMapKeyOverrides bool
}

func HasTemplating(node yamlmeta.Node) bool {
	return hasTemplating(node)
}

func hasTemplating(val interface{}) bool {
	node, ok := val.(yamlmeta.Node)
	if !ok {
		return false
	}

	metaOpts := template.MetaOpts{IgnoreUnknown: true}
	for _, comment := range node.GetComments() {
		ann, err := NewTemplateAnnotationFromYAMLComment(comment, node.GetPosition(), metaOpts)
		if err != nil {
			return false
		}
		if ann.Name != template.AnnotationComment {
			return true
		}
	}

	for _, childVal := range node.GetValues() {
		if hasTemplating(childVal) {
			return true
		}
	}
	return false
}

func NewTemplate(name string, opts TemplateOpts) *Template {
	return &Template{name: name, opts: opts, instructions: template.NewInstructionSet()}
}

// Compile converts a DocumentSet into code and metadata, returned via CompiledTemplate.
func (e *Template) Compile(docSet *yamlmeta.DocumentSet) (*template.CompiledTemplate, error) {
	e.docSet = docSet
	e.nodes = template.NewNodes()

	code, err := e.build(docSet, nil, template.NodeTagRoot, buildOpts{})
	if err != nil {
		return nil, err
	}

	code = append([]template.Line{
		e.resetCtxType(),
		{Instruction: e.instructions.NewStartCtx(EvaluationCtxDialectName)},
	}, code...)

	code = append(code, template.Line{
		Instruction: e.instructions.NewEndCtxNone(), // TODO ideally we would return array of docset
	})

	return template.NewCompiledTemplate(e.name, code, e.instructions, e.nodes, template.EvaluationCtxDialects{
		EvaluationCtxDialectName: EvaluationCtx{
			implicitMapKeyOverrides: e.opts.ImplicitMapKeyOverrides,
		},
		texttemplate.EvaluationCtxDialectName: texttemplate.EvaluationCtx{},
	}), nil
}

type buildOpts struct {
	TextTemplatedStrings bool
}

// build translates a document into a starlark program. This code is later evaluated to rebuild the output line by line, where annotated expressions are evaluated.
func (e *Template) build(nodeOrScalar interface{}, parentNode yamlmeta.Node, parentTag template.NodeTag, opts buildOpts) ([]template.Line, error) {
	node, ok := nodeOrScalar.(yamlmeta.Node)
	if !ok {
		if s, ok := nodeOrScalar.(string); ok && opts.TextTemplatedStrings {
			return e.buildString(s, parentNode, parentTag, e.instructions.NewSetNodeValue)
		}

		return []template.Line{{
			Instruction: e.instructions.NewSetNode(parentTag).WithDebug(e.debugComment(parentNode)),
			SourceLine:  e.newSourceLine(parentNode.GetPosition()),
		}}, nil
	}

	metas, nodeForEval, err := extractMetas(node, template.MetaOpts{IgnoreUnknown: e.opts.IgnoreUnknownComments})
	if err != nil {
		return nil, err
	}

	nodeTag := e.nodes.AddNode(nodeForEval, parentTag)

	if e.allowsTextTemplatedStrings(metas) {
		opts.TextTemplatedStrings = true
	}

	code := []template.Line{}

	for _, blk := range metas.Block {
		code = append(code, template.Line{
			Instruction: e.instructions.NewCode(blk.Data),
			SourceLine:  e.newSourceLine(blk.Position),
		})
	}

	for _, ann := range metas.Annotations {
		e.nodes.AddAnnotation(nodeTag, *ann.Annotation)
		code = append(code, template.Line{
			Instruction: e.instructions.NewStartNodeAnnotation(nodeTag, *ann.Annotation).WithDebug(e.debugComment(nodeForEval)),
			SourceLine:  e.newSourceLine(ann.Comment.Position),
		})
	}

	if mapItem, ok := nodeForEval.(*yamlmeta.MapItem); ok {
		if keyStr, ok := mapItem.Key.(string); ok && opts.TextTemplatedStrings {
			templateLines, err := e.buildString(keyStr, nodeForEval, nodeTag, e.instructions.NewSetMapItemKey)
			if err != nil {
				return nil, err
			}
			code = append(code, templateLines...)
		}
	}

	code = append(code, template.Line{
		Instruction: e.instructions.NewStartNode(nodeTag).WithDebug(e.debugComment(nodeForEval)),
		SourceLine:  e.newSourceLine(nodeForEval.GetPosition()),
	})

	if len(metas.Values) > 0 {
		for _, val := range metas.Values {
			code = append(code, template.Line{
				Instruction: e.instructions.NewSetNodeValue(nodeTag, val.Data).WithDebug(e.debugComment(nodeForEval)),
				SourceLine:  e.newSourceLine(val.Position),
			})
		}
	} else {
		for _, childVal := range nodeForEval.GetValues() {
			childCode, err := e.build(childVal, nodeForEval, nodeTag, opts)
			if err != nil {
				return nil, err
			}
			code = append(code, childCode...)
		}
	}

	if metas.NeedsEnd() {
		code = append(code, template.Line{
			// TODO should we set position to start node?
			Instruction: e.instructions.NewCode("end"),
		})
	}

	return code, nil
}

func (e *Template) allowsTextTemplatedStrings(metas Metas) bool {
	// TODO potentially use template.NewAnnotations(node).Has(AnnotationTextTemplatedStrings)
	// however if node was not processed by the template, it wont have any annotations set
	for _, metaAndAnn := range metas.Annotations {
		if metaAndAnn.Annotation.Name == AnnotationTextTemplatedStrings {
			return true
		}
	}
	return false
}

func (e *Template) buildString(val string, node yamlmeta.Node, nodeTag template.NodeTag,
	instruction func(template.NodeTag, string) template.Instruction) ([]template.Line, error) {

	// TODO line numbers for inlined template are somewhat correct
	// (does not handle pipe-multi-line string format - off by 1)
	textRoot, err := texttemplate.NewParser().ParseWithPosition([]byte(val), e.name, node.GetPosition())
	if err != nil {
		return nil, err
	}

	code, err := texttemplate.NewTemplate(e.name).CompileInline(textRoot, e.instructions, e.nodes)
	if err != nil {
		return nil, err
	}

	lastInstruction := code[len(code)-1].Instruction
	if lastInstruction.Op() != e.instructions.EndCtx {
		return nil, fmt.Errorf("Expected last instruction to be endctx, but was %#v", lastInstruction.Op())
	}

	code[len(code)-1] = template.Line{
		Instruction: instruction(nodeTag, lastInstruction.AsString()).WithDebug(e.debugComment(node)),
		SourceLine:  e.newSourceLine(node.GetPosition()),
	}

	code = append(code, e.resetCtxType())

	code = e.wrapCodeWithSourceLines(code)

	return code, nil
}

func (e *Template) resetCtxType() template.Line {
	return template.Line{
		Instruction: e.instructions.NewSetCtxType(EvaluationCtxDialectName),
	}
}

func (e *Template) debugComment(node yamlmeta.Node) string {
	var details string

	switch typedNode := node.(type) {
	case *yamlmeta.MapItem:
		details = fmt.Sprintf(" key=%s", typedNode.Key)
	case *yamlmeta.ArrayItem:
		details = " idx=?"
	}

	return fmt.Sprintf("%T%s", node, details) // TODO, node.GetRef())
}

func (e *Template) newSourceLine(pos *filepos.Position) *template.SourceLine {
	if pos.IsKnown() {
		if content, ok := e.sourceCodeLines()[pos.LineNum()]; ok {
			return template.NewSourceLine(pos, content)
		}
	}
	return nil
}

func (e *Template) sourceCodeLines() map[int]string {
	if e.srcLinesByLine != nil {
		return e.srcLinesByLine
	}

	e.srcLinesByLine = map[int]string{}

	if sourceCode, present := e.docSet.AsSourceBytes(); present {
		for i, line := range bytes.Split(sourceCode, []byte("\n")) {
			e.srcLinesByLine[filepos.NewPosition(i+1).LineNum()] = string(line)
		}
	}

	return e.srcLinesByLine
}

func (e *Template) wrapCodeWithSourceLines(code []template.Line) []template.Line {
	var wrappedCode []template.Line
	for _, line := range code {
		if line.SourceLine != nil {
			newSrcLine := e.newSourceLine(line.SourceLine.Position)
			if newSrcLine == nil {
				panic("Expected to find associated source line")
			}
			newSrcLine.Selection = line.SourceLine
			line.SourceLine = newSrcLine
		}
		wrappedCode = append(wrappedCode, line)
	}
	return wrappedCode
}
