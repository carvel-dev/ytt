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

type TypeAnnotation struct {
	//"schema/type" any=True, one_of=[array]
	any                bool
	listOfAllowedTypes []string
}

type NullableAnnotation struct {
	bool
	providedValueType yamlmeta.Type
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
		return &NullType{ValueType: n.providedValueType, Position: item.GetPosition()}
	}
	return nil
}

func processNullableAnnotation(item yamlmeta.Node, valueType yamlmeta.Type) (*NullableAnnotation, error) {
	templateAnnotations := template.NewAnnotations(item)
	if templateAnnotations.Has(AnnotationNullable) {
		return &NullableAnnotation{true, valueType}, nil
	}
	//what does returning an empty annotation mean? why not nil?
	return &NullableAnnotation{}, nil
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
	return &TypeAnnotation{}, nil
}

//func processAnnotations(item yamlmeta.Node) ([]Annotation, error) {
//	templateAnnotations := template.NewAnnotations(item)
//	var allAnnotations []Annotation
//	for _, annName := range getPossibleAnnotationNames() {
//		if templateAnnotations.Has(annName) {
//			processedAnn, err := callCorrectAnnMaker(annName, item)
//			if err != nil {
//				return nil, err
//			}
//			allAnnotations = append(allAnnotations, processedAnn)
//		}
//	}
//	return allAnnotations, nil
//
//}
//
//func getPossibleAnnotationNames() []structmeta.AnnotationName {
//	return []structmeta.AnnotationName{AnnotationType, AnnotationNullable}
//}
//
//func callCorrectAnnMaker(annName structmeta.AnnotationName, item yamlmeta.Node) (Annotation, error) {
//	switch annName {
//	case AnnotationType:
//		return processTypeAnnotations(item)
//	case AnnotationNullable:
//		return processNullableAnnotation(item)
//	default:
//		return nil, fmt.Errorf("HERE")
//	}
//}
