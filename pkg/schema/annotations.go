// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"sort"

	"github.com/k14s/ytt/pkg/filepos"

	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNullable     structmeta.AnnotationName = "schema/nullable"
	AnnotationType         structmeta.AnnotationName = "schema/type"
	TypeAnnotationKwargAny string                    = "any"
)

type Annotation interface {
	NewTypeFromAnn() yamlmeta.Type
}

type TypeAnnotation struct {
	any          bool
	itemPosition *filepos.Position
}

type NullableAnnotation struct {
	itemPosition      *filepos.Position
	ProvidedValueType yamlmeta.Type
}

func NewTypeAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return &TypeAnnotation{}, fmt.Errorf("expected @%v annotation to have keyword argument and value. Supported key-value pairs are '%v=True', '%v=False'", AnnotationType, TypeAnnotationKwargAny, TypeAnnotationKwargAny)
	}
	typeAnn := &TypeAnnotation{itemPosition: pos}
	for _, kwarg := range ann.Kwargs {
		argName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return &TypeAnnotation{}, err
		}

		switch argName {
		case TypeAnnotationKwargAny:
			isAnyType, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return &TypeAnnotation{},
					fmt.Errorf("processing @%v '%v' argument: %s", AnnotationType, TypeAnnotationKwargAny, err)
			}
			typeAnn.any = isAnyType

		default:
			return &TypeAnnotation{}, fmt.Errorf("unknown @%v annotation keyword argument '%v'. Supported kwargs are '%v'", AnnotationType, argName, TypeAnnotationKwargAny)
		}
	}
	return typeAnn, nil
}

func NewNullableAnnotation(valueType yamlmeta.Type, pos *filepos.Position) (*NullableAnnotation, error) {
	return &NullableAnnotation{pos, valueType}, nil
}

func (t *TypeAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if t.any {
		return &AnyType{Position: t.itemPosition}
	}
	return nil
}

func (n *NullableAnnotation) NewTypeFromAnn() yamlmeta.Type {
	return &NullType{ValueType: n.ProvidedValueType, Position: n.itemPosition}
}

func ProcessAnnotations(item yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []structmeta.AnnotationName{AnnotationType, AnnotationNullable} {
		ann, err := processOptionalAnnotations(item, annotation)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

func processOptionalAnnotations(node yamlmeta.Node, annotation structmeta.AnnotationName) (Annotation, error) {
	nodeAnnotations := template.NewAnnotations(node)

	if nodeAnnotations.Has(annotation) {
		switch annotation {
		case AnnotationNullable:
			wrappedValueType, err := newCollectionItemValueType(node.GetValues()[0], node.GetPosition())
			if err != nil {
				return nil, err
			}
			nullAnn, err := NewNullableAnnotation(wrappedValueType, node.GetPosition())
			if err != nil {
				return nil, NewInvalidSchemaError(node, err.Error(), "")
			}
			return nullAnn, nil
		case AnnotationType:
			ann, _ := nodeAnnotations[AnnotationType]
			typeAnn, err := NewTypeAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, NewInvalidSchemaError(node, err.Error(), "")
			}
			return typeAnn, nil
		}
	}

	return nil, nil
}

func ConvertAnnotationsToSingleType(anns []Annotation) yamlmeta.Type {
	listOfTypes := []yamlmeta.Type{}
	for _, a := range anns {
		newType := a.NewTypeFromAnn()
		if newType != nil {
			listOfTypes = append(listOfTypes, newType)
		}
	}

	finalType := chooseStrongestType(listOfTypes)
	return finalType
}

func chooseStrongestType(types []yamlmeta.Type) yamlmeta.Type {
	if len(types) == 0 {
		return nil
	}

	sort.Slice(types, func(i, j int) bool {
		if _, ok := types[i].(*AnyType); ok {
			return true
		}
		return false
	})

	return types[0]
}
