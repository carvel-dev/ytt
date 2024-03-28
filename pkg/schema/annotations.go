// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/validations"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/starlark-go/starlark"
)

// Declare @schema/... annotation names
const (
	AnnotationNullable     template.AnnotationName = "schema/nullable"
	AnnotationType         template.AnnotationName = "schema/type"
	AnnotationDefault      template.AnnotationName = "schema/default"
	AnnotationDescription  template.AnnotationName = "schema/desc"
	AnnotationTitle        template.AnnotationName = "schema/title"
	AnnotationExamples     template.AnnotationName = "schema/examples"
	AnnotationDeprecated   template.AnnotationName = "schema/deprecated"
	TypeAnnotationKwargAny string                  = "any"
	AnnotationValidation   template.AnnotationName = "schema/validation"
)

type Annotation interface {
	NewTypeFromAnn() (Type, error)
	GetPosition() *filepos.Position
}

type TypeAnnotation struct {
	any  bool
	node yamlmeta.Node
	pos  *filepos.Position
}

type NullableAnnotation struct {
	node yamlmeta.Node
	pos  *filepos.Position
}

// DefaultAnnotation is a wrapper for a value provided via @schema/default annotation
type DefaultAnnotation struct {
	val interface{}
	pos *filepos.Position
}

// DescriptionAnnotation documents the purpose of a node
type DescriptionAnnotation struct {
	description string
	pos         *filepos.Position
}

// TitleAnnotation provides title of a node
type TitleAnnotation struct {
	title string
	pos   *filepos.Position
}

// DeprecatedAnnotation is a wrapper for a value provided via @schema/deprecated annotation
type DeprecatedAnnotation struct {
	notice string
	pos    *filepos.Position
}

// ExampleAnnotation provides the Examples of a node
type ExampleAnnotation struct {
	examples []Example
	pos      *filepos.Position
}

// ValidationAnnotation is a wrapper for validations provided via @schema/validation annotation
type ValidationAnnotation struct {
	validation *validations.NodeValidation
	pos        *filepos.Position
}

// Example contains a yaml example and its description
type Example struct {
	description string
	example     interface{}
}

// documentation holds metadata about a Type, provided via documentation annotations
type documentation struct {
	title             string
	description       string
	deprecated        bool
	deprecationNotice string
	examples          []Example
}

// NewTypeAnnotation checks the keyword argument provided via @schema/type annotation, and returns wrapper for the annotated node.
func NewTypeAnnotation(ann template.NodeAnnotation, node yamlmeta.Node) (*TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     node.GetPosition(),
			description:  fmt.Sprintf("expected @%v annotation to have keyword argument and value", AnnotationType),
			expected:     "valid keyword argument and value",
			found:        fmt.Sprintf("missing keyword argument and value (by %s)", ann.Position.AsCompactString()),
			hints:        []string{fmt.Sprintf("Supported key-value pairs are '%v=True', '%v=False'", TypeAnnotationKwargAny, TypeAnnotationKwargAny)},
		}
	}
	typeAnn := &TypeAnnotation{node: node, pos: ann.Position}
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
					annPositions: []*filepos.Position{ann.Position},
					position:     node.GetPosition(),
					description:  "unknown @schema/type annotation keyword argument",
					expected:     "starlark.Bool",
					found:        fmt.Sprintf("%T (by %s)", kwarg[1], ann.Position.AsCompactString()),
					hints:        []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
				}
			}
			typeAnn.any = isAnyType

		default:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     node.GetPosition(),
				description:  "unknown @schema/type annotation keyword argument",
				expected:     "A valid kwarg",
				found:        fmt.Sprintf("%s (by %s)", argName, ann.Position.AsCompactString()),
				hints:        []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
			}
		}
	}
	return typeAnn, nil
}

// NewNullableAnnotation checks that there are no arguments, and returns wrapper for the annotated node.
func NewNullableAnnotation(ann template.NodeAnnotation, node yamlmeta.Node) (*NullableAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     node.GetPosition(),
			description:  fmt.Sprintf("expected @%v annotation to not contain any keyword arguments", AnnotationNullable),
			expected:     "starlark.Bool",
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationNullable, ann.Position.AsCompactString()),
			hints:        []string{fmt.Sprintf("Supported kwargs are '%v'", TypeAnnotationKwargAny)},
		}
	}

	return &NullableAnnotation{node, ann.Position}, nil
}

// NewDeprecatedAnnotation validates the value from the AnnotationDeprecated, and returns the value
func NewDeprecatedAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*DeprecatedAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     "string",
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationDeprecated, ann.Position.AsCompactString()),
			hints:        []string{"this annotation only accepts one argument: a string."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     "string",
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationDeprecated, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     "string",
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationDeprecated, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     "string",
			found:        fmt.Sprintf("Non-string value in @%v (by %v)", AnnotationDeprecated, ann.Position.AsCompactString()),
		}
	}
	return &DeprecatedAnnotation{strVal, ann.Position}, nil
}

// NewDefaultAnnotation checks the argument provided via @schema/default annotation, and returns wrapper for that value.
func NewDefaultAnnotation(ann template.NodeAnnotation, effectiveType Type, pos *filepos.Position) (*DefaultAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:     fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationDefault, ann.Position.AsCompactString()),
			hints: []string{
				"this annotation only accepts one argument: the default value.",
				"value must be in Starlark format, e.g.: {'key': 'value'}, True."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:     fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationDefault, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDefault),
			expected:     fmt.Sprintf("%s (by %s)", effectiveType.String(), effectiveType.GetDefinitionPosition().AsCompactString()),
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationDefault, ann.Position.AsCompactString()),
		}
	}

	val, err := core.NewStarlarkValue(ann.Args[0]).AsGoValue()
	if err != nil {
		// at this point the annotation is processed, and the Starlark evaluated
		panic(err)
	}
	return &DefaultAnnotation{yamlmeta.NewASTFromInterfaceWithPosition(val, pos), ann.Position}, nil
}

// NewDescriptionAnnotation validates the value from the AnnotationDescription, and returns the value
func NewDescriptionAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*DescriptionAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     "string",
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationDescription, ann.Position.AsCompactString()),
			hints:        []string{"this annotation only accepts one argument: a string."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     "string",
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationDescription, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     "string",
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationDescription, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     "string",
			found:        fmt.Sprintf("Non-string value in @%v (by %v)", AnnotationDescription, ann.Position.AsCompactString()),
		}
	}
	return &DescriptionAnnotation{strVal, ann.Position}, nil
}

// NewTitleAnnotation validates the value from the AnnotationTitle, and returns the value
func NewTitleAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*TitleAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     "string",
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationTitle, ann.Position.AsCompactString()),
			hints:        []string{"this annotation only accepts one argument: a string."},
		}
	}
	switch numArgs := len(ann.Args); {
	case numArgs == 0:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     "string",
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationTitle, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     "string",
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationTitle, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     "string",
			found:        fmt.Sprintf("Non-string value in @%v (by %v)", AnnotationTitle, ann.Position.AsCompactString()),
		}
	}
	return &TitleAnnotation{strVal, ann.Position}, nil
}

// NewExampleAnnotation validates the value(s) from the AnnotationExamples, and returns the value(s)
func NewExampleAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*ExampleAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
			expected:     "2-tuple containing description (string) and example value (of expected type)",
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
		}
	}
	if len(ann.Args) == 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
			expected:     "2-tuple containing description (string) and example value (of expected type)",
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
		}
	}

	var examples []Example
	for _, ex := range ann.Args {
		exampleTuple, ok := ex.(starlark.Tuple)
		if !ok {
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     "2-tuple containing description (string) and example value (of expected type)",
				found:        fmt.Sprintf("%v for @%v (at %v)", ex.Type(), AnnotationExamples, ann.Position.AsCompactString()),
			}
		}
		switch {
		case len(exampleTuple) == 0:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     "2-tuple containing description (string) and example value (of expected type)",
				found:        fmt.Sprintf("empty tuple in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
			}
		case len(exampleTuple) == 1:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     "2-tuple containing description (string) and example value (of expected type)",
				found:        fmt.Sprintf("empty tuple in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
			}
		case len(exampleTuple) > 2:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     "2-tuple containing description (string) and example value (of expected type)",
				found:        fmt.Sprintf("%v-tuple argument in @%v (by %v)", len(exampleTuple), AnnotationExamples, ann.Position.AsCompactString()),
			}
		default:
			description, err := core.NewStarlarkValue(exampleTuple[0]).AsString()
			if err != nil {
				return nil, schemaAssertionError{
					annPositions: []*filepos.Position{ann.Position},
					position:     pos,
					description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
					expected:     "2-tuple containing description (string) and example value (of expected type)",
					found:        fmt.Sprintf("%v value for @%v (at %v)", exampleTuple[0].Type(), AnnotationExamples, ann.Position.AsCompactString()),
				}
			}
			exampleVal, err := core.NewStarlarkValue(exampleTuple[1]).AsGoValue()
			if err != nil {
				panic(err)
			}
			examples = append(examples, Example{description, yamlmeta.NewASTFromInterfaceWithPosition(exampleVal, pos)})
		}
	}
	return &ExampleAnnotation{examples, ann.Position}, nil
}

// NewValidationAnnotation checks the values provided via @schema/validation annotation, and returns wrapper for the validation defined
func NewValidationAnnotation(ann template.NodeAnnotation) (*ValidationAnnotation, error) {
	validation, err := validations.NewValidationFromAnn(ann)
	if err != nil {
		return nil, fmt.Errorf("Invalid @%s annotation - %s", AnnotationValidation, err.Error())
	}

	return &ValidationAnnotation{validation, ann.Position}, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (t *TypeAnnotation) NewTypeFromAnn() (Type, error) {
	if t.any {
		return &AnyType{defaultValue: t.node.GetValues()[0], Position: t.node.GetPosition()}, nil
	}
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (n *NullableAnnotation) NewTypeFromAnn() (Type, error) {
	inferredType, err := InferTypeFromValue(n.node.GetValues()[0], n.node.GetPosition())
	if err != nil {
		return nil, err
	}
	return &NullType{ValueType: inferredType, Position: n.node.GetPosition()}, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (d *DefaultAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (e *ExampleAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation. DeprecatedAnnotation has no type information.
func (d *DeprecatedAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation. DescriptionAnnotation has no type information.
func (d *DescriptionAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation. TitleAnnotation has no type information.
func (t *TitleAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// NewTypeFromAnn returns type information given by annotation.
func (v *ValidationAnnotation) NewTypeFromAnn() (Type, error) {
	return nil, nil
}

// GetPosition returns position of the source comment used to create this annotation.
func (n *NullableAnnotation) GetPosition() *filepos.Position {
	return n.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (t *TypeAnnotation) GetPosition() *filepos.Position {
	return t.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (d *DefaultAnnotation) GetPosition() *filepos.Position {
	return d.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (d *DeprecatedAnnotation) GetPosition() *filepos.Position {
	return d.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (d *DescriptionAnnotation) GetPosition() *filepos.Position {
	return d.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (e *ExampleAnnotation) GetPosition() *filepos.Position {
	return e.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (t *TitleAnnotation) GetPosition() *filepos.Position {
	return t.pos
}

// GetPosition returns position of the source comment used to create this annotation.
func (v *ValidationAnnotation) GetPosition() *filepos.Position {
	return nil
}

// GetValidation gets the NodeValidation created from @schema/validation annotation
func (v *ValidationAnnotation) GetValidation() *validations.NodeValidation {
	return v.validation
}

func (t *TypeAnnotation) IsAny() bool {
	return t.any
}

// Val returns default value specified in annotation.
func (d *DefaultAnnotation) Val() interface{} {
	return d.val
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

func collectValueAnnotations(node yamlmeta.Node, effectiveType Type) ([]Annotation, error) {
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

	for _, annotation := range []template.AnnotationName{AnnotationDescription, AnnotationTitle, AnnotationExamples, AnnotationDeprecated} {
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

func processOptionalAnnotation(node yamlmeta.Node, optionalAnnotation template.AnnotationName, effectiveType Type) (Annotation, error) {
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
				return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v not supported on %s", AnnotationDefault, yamlmeta.TypeName(node)),
					schemaAssertionError{
						annPositions: []*filepos.Position{ann.Position},
						position:     node.GetPosition(),
					})
			case *yamlmeta.ArrayItem:
				return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v not supported on array item", AnnotationDefault),
					schemaAssertionError{
						annPositions: []*filepos.Position{ann.Position},
						position:     node.GetPosition(),
						hints:        []string{"do you mean to set a default value for the array?", "set an array's default by annotating its parent."},
					})
			}
			defaultAnn, err := NewDefaultAnnotation(ann, effectiveType, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return defaultAnn, nil
		case AnnotationDeprecated:
			if _, ok := node.(*yamlmeta.Document); ok {
				return nil, schemaAssertionError{
					description:  fmt.Sprintf("@%v not supported on a %s", AnnotationDeprecated, yamlmeta.TypeName(node)),
					annPositions: []*filepos.Position{ann.Position},
					position:     node.GetPosition(),
					hints:        []string{"do you mean to deprecate the entire schema document?", "use schema/deprecated on individual keys."},
				}
			}
			deprAnn, err := NewDeprecatedAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return deprAnn, nil
		case AnnotationDescription:
			descAnn, err := NewDescriptionAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return descAnn, nil
		case AnnotationExamples:
			exampleAnn, err := NewExampleAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return exampleAnn, nil
		case AnnotationTitle:
			titleAnn, err := NewTitleAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, err
			}
			return titleAnn, nil
		}
	}

	return nil, nil
}

func processValidationAnnotation(node yamlmeta.Node) (*ValidationAnnotation, error) {
	nodeAnnotations := template.NewAnnotations(node)
	if nodeAnnotations.Has(AnnotationValidation) {
		validationAnn, err := NewValidationAnnotation(nodeAnnotations[AnnotationValidation])
		if err != nil {
			return nil, err
		}
		return validationAnn, nil
	}
	return nil, nil
}

func getTypeFromAnnotations(anns []Annotation) (Type, error) {
	annsCopy := append([]Annotation{}, anns...)

	var typeFromAnn Type
	for _, ann := range annsCopy {
		switch typedAnn := ann.(type) {
		case *TypeAnnotation:
			if typedAnn.IsAny() {
				var err error
				typeFromAnn, err = typedAnn.NewTypeFromAnn()
				if err != nil {
					return nil, err
				}
			}
		case *NullableAnnotation:
			if typeFromAnn == nil {
				var err error
				typeFromAnn, err = typedAnn.NewTypeFromAnn()
				if err != nil {
					return nil, err
				}
			} else {
				typeFromAnn.SetDefaultValue(nil)
			}
		default:
			continue
		}
	}

	return typeFromAnn, nil
}

func setDocumentationFromAnns(docAnns []Annotation, typeOfValue Type) error {
	for _, a := range docAnns {
		switch ann := a.(type) {
		case *TitleAnnotation:
			typeOfValue.SetTitle(ann.title)
		case *DescriptionAnnotation:
			typeOfValue.SetDescription(ann.description)
		case *DeprecatedAnnotation:
			typeOfValue.SetDeprecated(true, ann.notice)
		case *ExampleAnnotation:
			err := checkExamplesValue(ann, typeOfValue)
			if err != nil {
				return err
			}
			typeOfValue.SetExamples(ann.examples)
		}
	}
	return nil
}

func checkExamplesValue(ann *ExampleAnnotation, typeOfValue Type) error {
	var typeCheck TypeCheck
	for _, ex := range ann.examples {
		if node, ok := ex.example.(yamlmeta.Node); ok {
			defaultValue := node.DeepCopyAsNode()
			chk := typeOfValue.AssignTypeTo(defaultValue)
			if !chk.HasViolations() {
				chk = CheckNode(defaultValue)
			}
			typeCheck.Violations = append(typeCheck.Violations, chk.Violations...)
		} else {
			chk := typeOfValue.CheckType(&yamlmeta.MapItem{Value: ex.example, Position: typeOfValue.GetDefinitionPosition()})
			typeCheck.Violations = append(typeCheck.Violations, chk.Violations...)
		}
	}
	if typeCheck.HasViolations() {
		var violations []error
		// add violating annotation position to error
		for _, err := range typeCheck.Violations {
			if typeCheckAssertionErr, ok := err.(schemaAssertionError); ok {
				typeCheckAssertionErr.annPositions = []*filepos.Position{ann.GetPosition()}
				typeCheckAssertionErr.found = typeCheckAssertionErr.found + fmt.Sprintf(" (by %v)", ann.GetPosition().AsCompactString())
				violations = append(violations, typeCheckAssertionErr)
			} else {
				violations = append(violations, err)
			}
		}
		return NewSchemaError(fmt.Sprintf("Invalid schema - @%v has wrong type", AnnotationExamples), violations...)
	}
	return nil
}

type checkForAnnotations struct{}

// Visit if `node` is annotated with `@schema/nullable`, `@schema/type`, or `@schema/default` (AnnotationNullable, AnnotationType, AnnotationDefault).
// Used when checking a node's children for undesired annotations.
//
// This visitor returns an error if any listed annotation is found,
// otherwise it will return nil.
func (c checkForAnnotations) Visit(n yamlmeta.Node) error {
	var foundAnns []string
	var foundAnnsPos []*filepos.Position
	nodeAnnotations := template.NewAnnotations(n)
	for _, annName := range []template.AnnotationName{AnnotationNullable, AnnotationType, AnnotationDefault} {
		if nodeAnnotations.Has(annName) {
			foundAnns = append(foundAnns, string(annName))
			foundAnnsPos = append(foundAnnsPos, nodeAnnotations[annName].Position)
		}
	}
	if len(foundAnnsPos) > 0 {
		foundText := strings.Join(foundAnns, ", @")
		return NewSchemaError("Invalid schema", schemaAssertionError{
			annPositions: foundAnnsPos,
			position:     n.GetPosition(),
			description:  "Schema was specified within an \"any type\" fragment",
			expected:     "no '@schema/...' on nodes within a node annotated '@schema/type any=True'",
			found:        fmt.Sprintf("@%v annotation(s)", foundText),
			hints:        []string{"an \"any type\" fragment has no constraints; nested schema annotations conflict and are disallowed"},
		})
	}

	return nil
}
