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
	AnnotationDescription  template.AnnotationName = "schema/desc"
	TypeAnnotationKwargAny string                  = "any"
)

type Annotation interface {
	NewTypeFromAnn() (yamlmeta.Type, error)
}

type TypeAnnotation struct {
	any  bool
	node yamlmeta.Node
}

type NullableAnnotation struct {
	node yamlmeta.Node
}

// DefaultAnnotation is a wrapper for a value provided via @schema/default annotation
type DefaultAnnotation struct {
	val interface{}
}

// DescriptionAnnotation documents the purpose of a node
type DescriptionAnnotation struct {
	description string
}

// NewTypeAnnotation checks the keyword argument provided via @schema/type annotation, and returns wrapper for the annotated node.
func NewTypeAnnotation(ann template.NodeAnnotation, node yamlmeta.Node) (*TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return nil, schemaAssertionError{
			position:    node.GetPosition(),
			description: fmt.Sprintf("expected @%v annotation to have keyword argument and value", AnnotationType),
			expected:    "valid keyword argument and value",
			found:       "missing keyword argument and value",
			hints:       []string{fmt.Sprintf("Supported key-value pairs are '%v=True', '%v=False'", TypeAnnotationKwargAny, TypeAnnotationKwargAny)},
		}
	}
	typeAnn := &TypeAnnotation{node: node}
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
					position:    node.GetPosition(),
					description: "unknown @schema/type annotation keyword argument",
					expected:    "starlark.Bool",
					found:       fmt.Sprintf("%T", kwarg[1]),
					hints:       []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
				}
			}
			typeAnn.any = isAnyType

		default:
			return nil, schemaAssertionError{
				position:    node.GetPosition(),
				description: "unknown @schema/type annotation keyword argument",
				expected:    "A valid kwarg",
				found:       argName,
				hints:       []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
			}
		}
	}
	return typeAnn, nil
}

// NewNullableAnnotation checks that there are no arguments, and returns wrapper for the annotated node.
func NewNullableAnnotation(ann template.NodeAnnotation, node yamlmeta.Node) (*NullableAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, fmt.Errorf("expected @%v annotation to not contain any keyword arguments", AnnotationNullable)
	}

	return &NullableAnnotation{node: node}, nil
}

// NewDefaultAnnotation checks the argument provided via @schema/default annotation, and returns wrapper for that value.
func NewDefaultAnnotation(ann template.NodeAnnotation, effectiveType yamlmeta.Type, pos *filepos.Position) (*DefaultAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:    fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
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
			expected:    fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
			found:       fmt.Sprintf("missing value (in @%v above this item)", AnnotationDefault),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:    fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
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

// NewDescriptionAnnotation validates the value from the AnnotationDescription, and returns the value
func NewDescriptionAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*DescriptionAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:    fmt.Sprintf("string"),
			found:       fmt.Sprintf("keyword argument (in @%v above this item)", AnnotationDescription),
			hints:       []string{"this annotation only accepts one argument: a string."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:    fmt.Sprintf("string"),
			found:       fmt.Sprintf("missing value (in @%v above this item)", AnnotationDescription),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:    fmt.Sprintf("string"),
			found:       fmt.Sprintf("%v values (in @%v above this item)", numArgs, AnnotationDescription),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			position:    pos,
			description: fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:    fmt.Sprintf("string"),
			found:       fmt.Sprintf("Non-string value (in @%v above this item)", AnnotationDescription),
		}
	}
	return &DescriptionAnnotation{strVal}, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (t *TypeAnnotation) NewTypeFromAnn() (yamlmeta.Type, error) {
	if t.any {
		return &AnyType{defaultValue: t.node.GetValues()[0], Position: t.node.GetPosition()}, nil
	}
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (n *NullableAnnotation) NewTypeFromAnn() (yamlmeta.Type, error) {
	inferredType, err := inferTypeFromValue(n.node.GetValues()[0], n.node.GetPosition())
	if err != nil {
		return nil, err
	}
	return &NullType{ValueType: inferredType, Position: n.node.GetPosition()}, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (n *DefaultAnnotation) NewTypeFromAnn() (yamlmeta.Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation. DescriptionAnnotation has no type information.
func (n *DescriptionAnnotation) NewTypeFromAnn() (yamlmeta.Type, error) {
	return nil, nil
}

func (t *TypeAnnotation) IsAny() bool {
	return t.any
}

// Val returns default value specified in annotation.
func (n *DefaultAnnotation) Val() interface{} {
	return n.val
}

func collectTypeAnnotations(node yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []template.AnnotationName{AnnotationType, AnnotationNullable} {
		ann, err := processOptionalAnnotation(node, annotation, nil)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

func collectValueAnnotations(node yamlmeta.Node, effectiveType yamlmeta.Type) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []template.AnnotationName{AnnotationNullable, AnnotationDefault} {
		ann, err := processOptionalAnnotation(node, annotation, effectiveType)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

// collectDocumentationAnnotations provides annotations that are used for documentation purposes
func collectDocumentationAnnotations(node yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []template.AnnotationName{AnnotationDescription} {
		ann, err := processOptionalAnnotation(node, annotation, nil)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

func processOptionalAnnotation(node yamlmeta.Node, optionalAnnotation template.AnnotationName, effectiveType yamlmeta.Type) (Annotation, error) {
	nodeAnnotations := template.NewAnnotations(node)

	if nodeAnnotations.Has(optionalAnnotation) {
		ann := nodeAnnotations[optionalAnnotation]

		switch optionalAnnotation {
		case AnnotationNullable:
			nullAnn, err := NewNullableAnnotation(ann, node)
			if err != nil {
				return nil, err
			}
			return nullAnn, nil
		case AnnotationType:
			typeAnn, err := NewTypeAnnotation(ann, node)
			if err != nil {
				return nil, err
			}
			return typeAnn, nil
		case AnnotationDefault:
			switch node.(type) {
			case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
				return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v not supported on %s", AnnotationDefault, node.DisplayName()),
					schemaAssertionError{position: node.GetPosition()})
			case *yamlmeta.ArrayItem:
				return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v not supported on array item", AnnotationDefault),
					schemaAssertionError{
						position: node.GetPosition(),
						hints:    []string{"do you mean to set a default value for the array?", "set an array's default by annotating its parent."},
					})
			}
			defaultAnn, err := NewDefaultAnnotation(ann, effectiveType, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return defaultAnn, nil
		case AnnotationDescription:
			descAnn, err := NewDescriptionAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return descAnn, nil
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
		typeFromAnn, err := annsCopy[0].NewTypeFromAnn()
		if err != nil {
			return nil, err
		}
		return typeFromAnn, nil
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

	typeFromAnn, err := conflictingTypeAnns[0].NewTypeFromAnn()
	if err != nil {
		return nil, err
	}
	return typeFromAnn, nil
}
