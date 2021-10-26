// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNullable     template.AnnotationName = "schema/nullable"
	AnnotationType         template.AnnotationName = "schema/type"
	AnnotationDefault      template.AnnotationName = "schema/default"
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

// DefaultAnnotation is a wrapper for a value provided via @schema/default annotation
type DefaultAnnotation struct {
	val interface{}
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

// NewDefaultAnnotation checks the argument provided via @schema/default annotation, and returns wrapper for that value.
func NewDefaultAnnotation(ann template.NodeAnnotation, inferredType yamlmeta.Type, pos *filepos.Position) (*DefaultAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:    fmt.Sprintf("%s (by %s)", inferredType.String(), inferredType.GetDefinitionPosition().AsCompactString()),
			found:       fmt.Sprintf("(keyword argument in @%v above this item)", AnnotationDefault),
			hints: []string{
				"this annotation only accepts one argument: the default value.",
				"value must be in Starlark format, e.g.: {'key': 'value'}, True."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:    fmt.Sprintf("%s (by %s)", inferredType.String(), inferredType.GetDefinitionPosition().AsCompactString()),
			found:       fmt.Sprintf("missing value (in @%v above this item)", AnnotationDefault),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:    fmt.Sprintf("%s (by %s)", inferredType.String(), inferredType.GetDefinitionPosition().AsCompactString()),
			found:       fmt.Sprintf("%v values (in @%v above this item)", numArgs, AnnotationDefault),
		}
	}

	val, err := core.NewStarlarkValue(ann.Args[0]).AsGoValue()
	if err != nil {
		//at this point the annotation is processed, and the Starlark evaluated
		panic(err)
	}
	return &DefaultAnnotation{yamlmeta.NewASTFromInterfaceWithPosition(val, pos)}, nil
}

func (t *TypeAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if t.any {
		return &AnyType{ValueType: t.inferredType, Position: t.itemPosition}
	}
	return nil
}

// NewTypeFromAnn returns type information given by annotation.
func (n *NullableAnnotation) NewTypeFromAnn() yamlmeta.Type {
	return &NullType{ValueType: n.providedValueType, Position: n.itemPosition}
}

// NewTypeFromAnn returns type information given by annotation.
func (n *DefaultAnnotation) NewTypeFromAnn() yamlmeta.Type {
	return nil
}

func (t *TypeAnnotation) IsAny() bool {
	return t.any
}

// Val returns default value specified in annotation.
func (n *DefaultAnnotation) Val() interface{} {
	return n.val
}

func collectAnnotations(item yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []template.AnnotationName{AnnotationType, AnnotationNullable, AnnotationDefault} {
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
		case AnnotationDefault:
			defaultAnn, err := NewDefaultAnnotation(ann, wrappedValueType, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return defaultAnn, nil
		}
	}

	return nil, nil
}

func getTypeFromAnnotations(anns []Annotation, pos *filepos.Position) (yamlmeta.Type, error) {
	annsCopy := append([]Annotation{}, anns...)

	if len(annsCopy) == 0 {
		return nil, nil
	}
	if len(annsCopy) == 1 {
		return annsCopy[0].NewTypeFromAnn(), nil
	}

	var conflictingTypeAnns []Annotation
	for _, ann := range annsCopy {
		switch typedAnn := ann.(type) {
		case *NullableAnnotation:
			conflictingTypeAnns = append(conflictingTypeAnns, ann)
		case *TypeAnnotation:
			if typedAnn.IsAny() {
				conflictingTypeAnns = append(conflictingTypeAnns, ann)
			}
		default:
			continue
		}
	}

	if len(conflictingTypeAnns) > 1 {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("@%v, and @%v any=True are mutually exclusive", AnnotationNullable, AnnotationType),
			expected:    fmt.Sprintf("one of %v, or %v any=True", AnnotationNullable, AnnotationType),
			found:       fmt.Sprintf("both @%v, and @%v any=True annotations", AnnotationNullable, AnnotationType),
		}
	}

	return conflictingTypeAnns[0].NewTypeFromAnn(), nil
}
