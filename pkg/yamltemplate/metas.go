// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"
	"strings"
	"unicode"

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

func NewTemplateMetaFromYAMLComment(comment *yamlmeta.Comment, opts MetasOpts) (template.Meta, error) {
	meta, err := template.NewMetaFromString(comment.Data, template.MetaOpts{IgnoreUnknown: opts.IgnoreUnknown})
	if err != nil {
		return template.Meta{}, fmt.Errorf(
			"Non-ytt comment at %s: '#%s': %s. (hint: if this is plain YAML — not a template — consider `--file-mark '<filename>:type=yaml-plain'`)",
			comment.Position.AsString(), comment.Data, err)
	}
	return meta, nil
}

func NewMetas(node yamlmeta.Node, opts MetasOpts) (Metas, error) {
	metas := Metas{}

	for _, comment := range node.GetComments() {
		meta, err := NewTemplateMetaFromYAMLComment(comment, opts)
		if err != nil {
			return metas, err
		}

		for _, ann := range meta.Annotations {
			if len(ann.Name) == 0 {
				// Default code and value annotations to make templates less verbose
				ann.Name = template.AnnotationCode

				if node.GetPosition().IsKnown() {
					if comment.Position.LineNum() == node.GetPosition().LineNum() {
						if len(node.GetValues()) > 0 && node.GetValues()[0] != nil {
							return metas, fmt.Errorf(
								"Expected YAML node at %s to have either computed or YAML value, but found both",
								comment.Position.AsString())
						}

						ann.Name = template.AnnotationValue
					}
				}
			}

			switch ann.Name {
			case template.AnnotationValue:
				metas.Values = append(metas.Values, &yamlmeta.Comment{
					Position: comment.Position,
					Data:     ann.Content,
				})

			case template.AnnotationCode:
				if metas.needsEnds > 0 {
					return metas, fmt.Errorf(
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
				// ignore

			default:
				metas.Annotations = append(metas.Annotations, CommentAndAnnotation{comment, ann})
			}
		}
	}

	return metas, nil
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
