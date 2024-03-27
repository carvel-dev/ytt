// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"
	"strings"
	"unicode"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/yamlmeta"
)

var (
	nodeSpecificKeywords = map[string]string{
		"if/end":   "if",
		"else/end": "else",
		"elif/end": "elif",
		"for/end":  "for",
		"def/end":  "def",
	}
)

// Metas are the collection of ytt YAML templating values parsed from the comments attached to a yamlmeta.Node
type Metas struct {
	Block       []*yamlmeta.Comment // meant to execute some code
	Values      []*yamlmeta.Comment // meant to return interpolated value
	Annotations []CommentAndAnnotation
	needsEnds   int
}

type CommentAndAnnotation struct {
	Comment    *yamlmeta.Comment
	Annotation *template.Annotation
}

// NewTemplateAnnotationFromYAMLComment parses "comment" into template.Annotation.
// nodePos is the position of the node to which "comment" is attached.
//
// When a comment contains a shorthand annotation (i.e. `#@ `):
//   - if the comment is above its node, it's a template.AnnotationCode annotation: it's _merely_ contributing raw
//     Starlark code.
//   - if the comment is on the same line as its node, it's a template.AnnotationValue annotation: it's meant to set the
//     value of the annotated node.
func NewTemplateAnnotationFromYAMLComment(comment *yamlmeta.Comment, nodePos *filepos.Position, opts template.MetaOpts) (template.Annotation, error) {
	ann, err := template.NewAnnotationFromComment(comment.Data, comment.Position, opts)
	if err != nil {
		return template.Annotation{}, fmt.Errorf(
			"Failed to parse line %s: '#%s': %s", comment.Position.AsIntString(), comment.Data, err)
	}

	if len(ann.Name) == 0 {
		// Default code and value annotations to make templates less verbose
		ann.Name = template.AnnotationCode

		if nodePos.IsKnown() {
			if comment.Position.LineNum() == nodePos.LineNum() {
				ann.Name = template.AnnotationValue
			}
		}
	}

	return ann, nil
}

// extractMetas parses "metas" (i.e. code, values, and/or annotations) from node comments.
// Annotations that are consumable, such as AnnotationValue, AnnotationCode, and AnnotationComment,
// are processed and added for later consumption. Other annotations remain attached to the node.
//
// Returns the extracted metas and a copy of "node" with code and value type comments removed.
func extractMetas(node yamlmeta.Node, opts template.MetaOpts) (Metas, yamlmeta.Node, error) {
	metas := Metas{}

	nonTemplateComments := []*yamlmeta.Comment{}
	for _, comment := range node.GetComments() {
		ann, err := NewTemplateAnnotationFromYAMLComment(comment, node.GetPosition(), opts)
		if err != nil {
			return metas, nil, err
		}

		switch ann.Name {
		case template.AnnotationValue:
			if len(node.GetValues()) > 0 && node.GetValues()[0] != nil {
				return metas, nil, fmt.Errorf(
					"Expected YAML node at %s to have either computed or YAML value, but found both",
					comment.Position.AsString())
			}

			metas.Values = append(metas.Values, &yamlmeta.Comment{
				Position: comment.Position,
				Data:     ann.Content,
			})

		case template.AnnotationCode:
			if metas.needsEnds > 0 {
				return metas, nil, fmt.Errorf(
					"Unexpected code at %s after use of '*/end', expected YAML node",
					comment.Position.AsString())
			}

			code := ann.Content
			spacePrefix := metas.spacePrefix(code)

			for keyword, replacementKeyword := range nodeSpecificKeywords {
				if strings.HasPrefix(code, spacePrefix+keyword) {
					metas.needsEnds++
					code = strings.Replace(code, spacePrefix+keyword, spacePrefix+replacementKeyword, 1)
				}
			}

			metas.Block = append(metas.Block, &yamlmeta.Comment{
				Position: comment.Position,
				Data:     code,
			})

		case template.AnnotationComment:
			// ytt comments are not considered "meta": no nothing.
			// They _are_ considered part of the template's code, so these yamlmeta.Comments are "digested."

		default:
			metas.Annotations = append(metas.Annotations, CommentAndAnnotation{comment, &ann})
			nonTemplateComments = append(nonTemplateComments, comment)
		}
	}
	digestedNode := node.DeepCopyAsNode()
	digestedNode.SetComments(nonTemplateComments)

	return metas, digestedNode, nil
}

func (m Metas) NeedsEnd() bool { return m.needsEnds != 0 }

func (m Metas) spacePrefix(str string) string {
	for i, r := range str {
		if !unicode.IsSpace(r) {
			return str[:i]
		}
	}
	return ""
}
