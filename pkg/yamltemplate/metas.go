package yamltemplate

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/k14s/ytt/pkg/structmeta"
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
	Block       []*yamlmeta.Meta // meant to execute some code
	Values      []*yamlmeta.Meta // meant to return interpolated value
	Annotations []MetaAndAnnotation
	needsEnds   int
}

type MetaAndAnnotation struct {
	Meta       *yamlmeta.Meta
	Annotation *structmeta.Annotation
}

type MetasOpts struct {
	IgnoreUnknown bool
}

func NewStructMetaFromMeta(meta *yamlmeta.Meta, opts MetasOpts) (structmeta.Meta, error) {
	structMeta, err := structmeta.NewMetaFromString(meta.Data, structmeta.MetaOpts{IgnoreUnknown: opts.IgnoreUnknown})
	if err != nil {
		return structmeta.Meta{}, fmt.Errorf(
			"Unknown comment syntax at %s: '%s': %s", meta.Position.AsString(), meta.Data, err)
	}
	return structMeta, nil
}

func NewMetas(node yamlmeta.Node, opts MetasOpts) (Metas, error) {
	metas := Metas{}

	for _, meta := range node.GetMetas() {
		structMeta, err := NewStructMetaFromMeta(meta, opts)
		if err != nil {
			return metas, err
		}

		for _, ann := range structMeta.Annotations {
			if len(ann.Name) == 0 {
				// Default code and value annotations to make templates less verbose
				ann.Name = template.AnnotationCode

				if node.GetPosition().IsKnown() {
					if meta.Position.Line() == node.GetPosition().Line() {
						if len(node.GetValues()) > 0 && node.GetValues()[0] != nil {
							return metas, fmt.Errorf(
								"Expected YAML node at %s to have either computed or YAML value, but found both",
								meta.Position.AsString())
						}

						ann.Name = template.AnnotationValue
					}
				}
			}

			switch ann.Name {
			case template.AnnotationValue:
				metas.Values = append(metas.Values, &yamlmeta.Meta{
					Position: meta.Position,
					Data:     ann.Content,
				})

			case template.AnnotationCode:
				if metas.needsEnds > 0 {
					return metas, fmt.Errorf(
						"Unexpected code at %s after use of '*/end', expected YAML node",
						meta.Position.AsString())
				}

				code := ann.Content
				spacePrefix := metas.spacePrefix(code)

				for keyword, replacementKeyword := range nodeSpecificKeywords {
					if strings.HasPrefix(code, spacePrefix+keyword) {
						metas.needsEnds += 1
						code = strings.Replace(code, spacePrefix+keyword, spacePrefix+replacementKeyword, 1)
					}
				}

				metas.Block = append(metas.Block, &yamlmeta.Meta{
					Position: meta.Position,
					Data:     code,
				})

			case template.AnnotationComment:
				// ignore

			default:
				metas.Annotations = append(metas.Annotations, MetaAndAnnotation{meta, ann})
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
