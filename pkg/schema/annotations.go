// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

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
	NewTypeFromAnn(item yamlmeta.Node) yamlmeta.Type
}

// TODO: this could use a  less conflicting name
type TypeAnnotation struct {
	//"schema/type" any=True, one_of=[array]
	any bool
	//listOfAllowedTypes []string
}

type NullableAnnotation struct {
	//TODO: name member variable
	bool
	ProvidedValueType yamlmeta.Type
}

func NewTypeAnnotation(ann template.NodeAnnotation) (TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return TypeAnnotation{}, fmt.Errorf("expected @%v annotation to have keyword argument and value. Supported key-value pairs are '%v=True', '%v=False'", AnnotationType, TypeAnnotationKwargAny, TypeAnnotationKwargAny)
	}
	typeAnn := TypeAnnotation{}
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

func (t *TypeAnnotation) NewTypeFromAnn(item yamlmeta.Node) yamlmeta.Type {
	if t.any {
		return &AnyType{Position: item.GetPosition()}
	}
	return nil
}

func (n *NullableAnnotation) NewTypeFromAnn(item yamlmeta.Node) yamlmeta.Type {
	if n.bool {
		return &NullType{ValueType: n.ProvidedValueType, Position: item.GetPosition()}
	}
	return nil
}

func processNullableAnnotation(item yamlmeta.Node) (*NullableAnnotation, error) {
	templateAnnotations := template.NewAnnotations(item)
	if templateAnnotations.Has(AnnotationNullable) {
		// TODO: item.GetValues :(
		valueType, err := newCollectionItemValueType(item.GetValues()[0], item.GetPosition())
		if err != nil {
			return nil, err
		}
		return &NullableAnnotation{true, valueType}, nil
	}
	//what does returning an empty annotation mean? why not nil?
	return nil, nil
}

func processTypeAnnotations(item yamlmeta.Node) (*TypeAnnotation, error) {
	templateAnnotations := template.NewAnnotations(item)

	if templateAnnotations.Has(AnnotationType) {
		ann, _ := templateAnnotations[AnnotationType]
		typeAnn, err := NewTypeAnnotation(ann)
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

func ConvertAnnotationsToType(anns []Annotation, item yamlmeta.Node) (yamlmeta.Type, error) {
	valueType, err := newCollectionItemValueType(item.GetValues()[0], item.GetPosition())
	if err != nil {
		return nil, err
	}
	listOfTypes := []yamlmeta.Type{valueType}
	for _, a := range anns {
		newType := a.NewTypeFromAnn(item)
		listOfTypes = append(listOfTypes, newType)
	}

	//finalType, err := combineTypes(listOfTypes)
	//if err != nil {
	//	return nil, err
	//}

	return valueType, nil
}
