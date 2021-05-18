// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
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

// TODO: this could use a  less conflicting name
type TypeAnnotation struct {
	//"schema/type" any=True, one_of=[array]
	any          bool
	itemPosition *filepos.Position
	//listOfAllowedTypes []string
}

type NullableAnnotation struct {
	//TODO: name member variable
	bool
	itemPosition      *filepos.Position
	ProvidedValueType yamlmeta.Type
}

func NewTypeAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return TypeAnnotation{}, fmt.Errorf("expected @%v annotation to have keyword argument and value. Supported key-value pairs are '%v=True', '%v=False'", AnnotationType, TypeAnnotationKwargAny, TypeAnnotationKwargAny)
	}
	typeAnn := TypeAnnotation{itemPosition: pos}
	for _, kwarg := range ann.Kwargs {

		argName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return TypeAnnotation{}, err
		}

		switch argName {
		case TypeAnnotationKwargAny:
			isAnyType, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return TypeAnnotation{},
					fmt.Errorf("processing @%v '%v' argument: %s", AnnotationType, TypeAnnotationKwargAny, err)
			}
			typeAnn.any = isAnyType

		default:
			return TypeAnnotation{}, fmt.Errorf("unknown @%v annotation keyword argument '%v'. Supported kwargs are '%v'", AnnotationType, argName, TypeAnnotationKwargAny)
		}
	}
	return typeAnn, nil
}

func (t *TypeAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if t.any {
		return &AnyType{Position: t.itemPosition}
	}
	return nil
}

func (n *NullableAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if n.bool {
		return &NullType{ValueType: n.ProvidedValueType, Position: n.itemPosition}
	}
	return nil
}

func processNullableAnnotation(item yamlmeta.Node) (*NullableAnnotation, error) {
	templateAnnotations := template.NewAnnotations(item)
	if templateAnnotations.Has(AnnotationNullable) {
		// TODO: item.GetValues >:(
		valueType, err := newCollectionItemValueType(item.GetValues()[0], item.GetPosition())
		if err != nil {
			return nil, err
		}
		return &NullableAnnotation{true, item.GetPosition(), valueType}, nil
	}
	//what does returning an empty annotation mean? why not nil?
	return nil, nil
}

func processTypeAnnotations(item yamlmeta.Node) (*TypeAnnotation, error) {
	templateAnnotations := template.NewAnnotations(item)

	if templateAnnotations.Has(AnnotationType) {
		ann, _ := templateAnnotations[AnnotationType]
		typeAnn, err := NewTypeAnnotation(ann, item.GetPosition())
		if err != nil {
			return nil, NewInvalidSchemaError(item, err.Error(), "")
		}
		return &typeAnn, nil
	}
	//what does returning an empty annotation mean? why not nil?
	return nil, nil
}

func ProcessAnnotations(item yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	tAnn, err := processTypeAnnotations(item)
	if err != nil {
		return nil, err
	}
	if tAnn != nil {
		anns = append(anns, tAnn)
	}
	nAnn, err := processNullableAnnotation(item)
	if err != nil {
		return nil, err
	}
	if nAnn != nil {
		anns = append(anns, nAnn)
	}
	return anns, nil
}

func ConvertAnnotationsToSingleType(anns []Annotation) yamlmeta.Type {
	listOfTypes := []yamlmeta.Type{}
	for _, a := range anns {
		newType := a.NewTypeFromAnn()
		if newType != nil {
			listOfTypes = append(listOfTypes, newType)
		}
	}

	finalType := chooseType(listOfTypes)
	return finalType
}

func chooseType(types []yamlmeta.Type) yamlmeta.Type {
	if len(types) == 0 {
		return nil
	}
	if len(types) == 1 {
		return types[0]
	}
	if len(types) > 1 {
		if n, ok := hasAnyType(types); ok {
			return types[n]
		}
		if n, ok := hasNullType(types); ok {
			return types[n]
		}
	}
	return nil
}

func hasAnyType(types []yamlmeta.Type) (int, bool) {
	for n, t := range types {
		if _, ok := t.(*AnyType); ok {
			return n, true
		}
	}
	return 0, false
}

func hasNullType(types []yamlmeta.Type) (int, bool) {
	for n, t := range types {
		if _, ok := t.(*NullType); ok {
			return n, true
		}
	}
	return 0, false
}
