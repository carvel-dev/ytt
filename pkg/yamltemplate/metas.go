// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
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

type MetasOpts struct {
	IgnoreUnknown bool
}

// NewTemplateAnnotationFromYAMLComment parses "comment" into template.Annotation.
//
// nodePos is the position of the node to which "comment" is attached (this is important to differentiate between
// @template/code and @template/value).
func NewTemplateAnnotationFromYAMLComment(comment *yamlmeta.Comment, nodePos *filepos.Position, opts MetasOpts) (template.Annotation, error) {
	ann, err := template.NewAnnotationFromString(comment.Data, template.MetaOpts{IgnoreUnknown: opts.IgnoreUnknown})
	if err != nil {
		return template.Annotation{}, fmt.Errorf(
			"Non-ytt comment at %s: '#%s': %s. (hint: if this is plain YAML — not a template — consider `--file-mark '<filename>:type=yaml-plain'`)",
			comment.Position.AsString(), comment.Data, err)
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

// extractMetas parses "metas" (i.e. code, values, and/or annotations) from node comments
//
// returns the extracted metas and a copy of "node" with code and value type comments removed.
func extractMetas(node yamlmeta.Node, opts MetasOpts) (Metas, yamlmeta.Node, error) {
	metas := Metas{}

	nonCodeComments := []*yamlmeta.Comment{}
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
			ann.Position = comment.Position
			metas.Annotations = append(metas.Annotations, CommentAndAnnotation{comment, &ann})
			nonCodeComments = append(nonCodeComments, comment)
		}
	}
	digestedNode := node.DeepCopyAsNode()
	digestedNode.SetComments(nonCodeComments)

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
