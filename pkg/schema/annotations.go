// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
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
	AnnotationTitle        template.AnnotationName = "schema/title"
	AnnotationExamples     template.AnnotationName = "schema/examples"
	AnnotationDeprecated   template.AnnotationName = "schema/deprecated"
	TypeAnnotationKwargAny string                  = "any"
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

//DeprecatedAnnotation is a wrapper for a value provided via @schema/deprecated annotation
type DeprecatedAnnotation struct {
	notice string
	pos    *filepos.Position
}

// ExampleAnnotation provides the Examples of a node
type ExampleAnnotation struct {
	examples []Example
	pos      *filepos.Position
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
			expected:     fmt.Sprintf("string"),
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
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationDeprecated, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationDeprecated, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDeprecated),
			expected:     fmt.Sprintf("string"),
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
		//at this point the annotation is processed, and the Starlark evaluated
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
			expected:     fmt.Sprintf("string"),
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
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationDescription, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationDescription, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationDescription),
			expected:     fmt.Sprintf("string"),
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
			expected:     fmt.Sprintf("string"),
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
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("missing value in @%v (by %v)", AnnotationTitle, ann.Position.AsCompactString()),
		}
	case numArgs > 1:
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     fmt.Sprintf("string"),
			found:        fmt.Sprintf("%v values in @%v (by %v)", numArgs, AnnotationTitle, ann.Position.AsCompactString()),
		}
	}

	strVal, err := core.NewStarlarkValue(ann.Args[0]).AsString()
	if err != nil {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationTitle),
			expected:     fmt.Sprintf("string"),
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
			expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
			found:        fmt.Sprintf("keyword argument in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
		}
	}
	if len(ann.Args) == 0 {
		return nil, schemaAssertionError{
			annPositions: []*filepos.Position{ann.Position},
			position:     pos,
			description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
			expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
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
				expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
				found:        fmt.Sprintf("%v for @%v (at %v)", ex.Type(), AnnotationExamples, ann.Position.AsCompactString()),
			}
		}
		switch {
		case len(exampleTuple) == 0:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
				found:        fmt.Sprintf("empty tuple in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
			}
		case len(exampleTuple) == 1:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
				found:        fmt.Sprintf("empty tuple in @%v (by %v)", AnnotationExamples, ann.Position.AsCompactString()),
			}
		case len(exampleTuple) > 2:
			return nil, schemaAssertionError{
				annPositions: []*filepos.Position{ann.Position},
				position:     pos,
				description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
				expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
				found:        fmt.Sprintf("%v-tuple argument in @%v (by %v)", len(exampleTuple), AnnotationExamples, ann.Position.AsCompactString()),
			}
		default:
			description, err := core.NewStarlarkValue(exampleTuple[0]).AsString()
			if err != nil {
				return nil, schemaAssertionError{
					annPositions: []*filepos.Position{ann.Position},
					position:     pos,
					description:  fmt.Sprintf("syntax error in @%v annotation", AnnotationExamples),
					expected:     fmt.Sprintf("2-tuple containing description (string) and example value (of expected type)"),
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
		ann.Position = populateAnnotationPosition(ann.Position, node)

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

func getTypeFromAnnotations(anns []Annotation, pos *filepos.Position) (Type, error) {
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
		annPositions := make([]*filepos.Position, len(conflictingTypeAnns))
		for i, a := range conflictingTypeAnns {
			annPositions[i] = a.GetPosition()
		}
		return nil, schemaAssertionError{
			annPositions: annPositions,
			position:     pos,
			description:  fmt.Sprintf("@%v, and @%v any=True are mutually exclusive", AnnotationNullable, AnnotationType),
			expected:     fmt.Sprintf("one of %v, or %v any=True", AnnotationNullable, AnnotationType),
			found:        fmt.Sprintf("both @%v, and @%v any=True annotations", AnnotationNullable, AnnotationType),
		}
	}

	typeFromAnn, err := conflictingTypeAnns[0].NewTypeFromAnn()
	if err != nil {
		return nil, err
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

func (c checkForAnnotations) Visit(n yamlmeta.Node) error {
	var foundAnns []string
	var foundAnnsPos []*filepos.Position
	nodeAnnotations := template.NewAnnotations(n)
	for _, annotation := range []template.AnnotationName{AnnotationNullable, AnnotationType, AnnotationDefault} {
		if nodeAnnotations.Has(annotation) {
			foundAnns = append(foundAnns, string(annotation))
			annPos := populateAnnotationPosition(nodeAnnotations[annotation].Position, n)
			foundAnnsPos = append(foundAnnsPos, annPos)
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

func populateAnnotationPosition(annPos *filepos.Position, node yamlmeta.Node) *filepos.Position {
	leftPadding := 0
	nodePos := node.GetPosition()
	if nodePos.IsKnown() {
		nodeLine := nodePos.GetLine()
		leftPadding = len(nodeLine) - len(strings.TrimLeft(nodeLine, " "))
	}

	lineString := ""
	nodeComments := node.GetComments()
	for _, c := range nodeComments {
		if c.Position.IsKnown() && c.Position.AsIntString() == fmt.Sprintf("%d", annPos.LineNum()) {
			lineString = fmt.Sprintf("%v#%s", strings.Repeat(" ", leftPadding), c.Data)
		}
	}

	annPos.SetFile(nodePos.GetFile())
	annPos.SetLine(lineString)

	return annPos
}
