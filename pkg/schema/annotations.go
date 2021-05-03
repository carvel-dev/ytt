// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"sort"

	"github.com/k14s/ytt/pkg/filepos"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNullable     template.AnnotationName = "schema/nullable"
	AnnotationType         template.AnnotationName = "schema/type"
	TypeAnnotationKwargAny string                  = "any"
)

type Annotation interface {
	NewTypeFromAnn() yamlmeta.Type
}

type TypeAnnotation struct {
	any          bool
	inferredType yamlmeta.Type
	itemPosition *filepos.Position
}

type NullableAnnotation struct {
	providedValueType yamlmeta.Type
	itemPosition      *filepos.Position
}

func NewTypeAnnotation(ann template.NodeAnnotation, inferredType yamlmeta.Type, pos *filepos.Position) (*TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("expected @%v annotation to have keyword argument and value", AnnotationType),
			expected:    "valid keyword argument and value",
			found:       "missing keyword argument and value",
			hints:       []string{fmt.Sprintf("Supported key-value pairs are '%v=True', '%v=False'", TypeAnnotationKwargAny, TypeAnnotationKwargAny)},
		}
	}
	typeAnn := &TypeAnnotation{inferredType: inferredType, itemPosition: pos}
	for _, kwarg := range ann.Kwargs {
		argName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return nil, err
		}

		switch argName {
		case TypeAnnotationKwargAny:
			isAnyType, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return nil, schemaAssertionError{
					position:    pos,
					description: "unknown @schema/type annotation keyword argument",
					expected:    "starlark.Bool",
					found:       fmt.Sprintf("%T", kwarg[1]),
					hints:       []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
				}
			}
			typeAnn.any = isAnyType

		default:
			return nil, schemaAssertionError{
				position:    pos,
				description: "unknown @schema/type annotation keyword argument",
				expected:    "A valid kwarg",
				found:       argName,
				hints:       []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
			}
		}
	}
	return typeAnn, nil
}

func NewNullableAnnotation(ann template.NodeAnnotation, valueType yamlmeta.Type, pos *filepos.Position) (*NullableAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, fmt.Errorf("expected @%v annotation to not contain any keyword arguments", AnnotationNullable)
	}

	return &NullableAnnotation{valueType, pos}, nil
}

func (t *TypeAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if t.any {
		return &AnyType{ValueType: t.inferredType, Position: t.itemPosition}
	}
	return nil
}

func (t *TypeAnnotation) IsAny() bool {
	return t.any
}

func (n *NullableAnnotation) NewTypeFromAnn() yamlmeta.Type {
	return &NullType{ValueType: n.providedValueType, Position: n.itemPosition}
}

func collectAnnotations(item yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []template.AnnotationName{AnnotationType, AnnotationNullable} {
		ann, err := processOptionalAnnotation(item, annotation)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

func processOptionalAnnotation(node yamlmeta.Node, optionalAnnotation template.AnnotationName) (Annotation, error) {
	nodeAnnotations := template.NewAnnotations(node)

	if nodeAnnotations.Has(optionalAnnotation) {
		ann, _ := nodeAnnotations[optionalAnnotation]

		wrappedValueType, err := inferTypeFromValue(node.GetValues()[0], node.GetPosition())
		if err != nil {
			return nil, err
		}

		switch optionalAnnotation {
		case AnnotationNullable:
			nullAnn, err := NewNullableAnnotation(ann, wrappedValueType, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return nullAnn, nil
		case AnnotationType:
			typeAnn, err := NewTypeAnnotation(ann, wrappedValueType, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return typeAnn, nil
		}
	}

	return nil, nil
}

func getTypeFromAnnotations(anns []Annotation) yamlmeta.Type {
	annsCopy := append([]Annotation{}, anns...)

	if len(annsCopy) == 0 {
		return nil
	}

	// allow Configuration Author to annotate "nullable" as a fallback if "any" is false.
	preferAnyTypeOverNullableType := func(i, j int) bool {
		if typeAnn, ok := annsCopy[i].(*TypeAnnotation); ok && typeAnn.IsAny() {
			return true
		}
		return false
	}

	sort.Slice(annsCopy, preferAnyTypeOverNullableType)
	return annsCopy[0].NewTypeFromAnn()
}
