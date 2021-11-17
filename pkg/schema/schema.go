// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/workspace/ref"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type DocumentSchema struct {
	Source     *yamlmeta.Document
	defaultDVs *yamlmeta.Document
	DocType    *DocumentType
}

type DocumentSchemaEnvelope struct {
	Doc *DocumentSchema

	used           bool
	originalLibRef []ref.LibraryRef
	libRef         []ref.LibraryRef
}

func NewDocumentSchema(doc *yamlmeta.Document) (*DocumentSchema, error) {
	docType, err := inferTypeFromValue(doc, doc.Position)
	if err != nil {
		return nil, err
	}

	schemaDVs := docType.GetDefaultValue()

	return &DocumentSchema{
		Source:     doc,
		defaultDVs: schemaDVs.(*yamlmeta.Document),
		DocType:    docType.(*DocumentType),
	}, nil
}

func NewDocumentSchemaEnvelope(doc *yamlmeta.Document) (*DocumentSchemaEnvelope, error) {
	libRef, err := getSchemaLibRef(ref.LibraryRefExtractor{}, doc)
	if err != nil {
		return nil, err
	}

	schema, err := NewDocumentSchema(doc)
	if err != nil {
		return nil, err
	}

	return &DocumentSchemaEnvelope{
		Doc:            schema,
		originalLibRef: libRef,
		libRef:         libRef,
	}, nil
}

// NewNullSchema provides the "Null Object" value of Schema. This is used in the case where no schema was provided.
func NewNullSchema() *DocumentSchema {
	return &DocumentSchema{
		Source: &yamlmeta.Document{},
		DocType: &DocumentType{
			ValueType: &AnyType{}},
	}
}

func NewDocumentType(doc *yamlmeta.Document) (*DocumentType, error) {
	typeOfValue, err := getType(doc)
	if err != nil {
		return nil, err
	}

	defaultValue, err := getValue(doc, typeOfValue)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &DocumentType{Source: doc, Position: doc.Position, ValueType: typeOfValue, defaultValue: defaultValue}, nil
}

func NewMapType(m *yamlmeta.Map) (*MapType, error) {
	mapType := &MapType{Position: m.Position}

	for _, mapItem := range m.Items {
		mapItemType, err := NewMapItemType(mapItem)
		if err != nil {
			return nil, err
		}
		mapType.Items = append(mapType.Items, mapItemType)
	}

	return mapType, nil
}

func NewMapItemType(item *yamlmeta.MapItem) (*MapItemType, error) {
	typeOfValue, err := getType(item)
	if err != nil {
		return nil, err
	}

	defaultValue, err := getValue(item, typeOfValue)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &MapItemType{Key: item.Key, ValueType: typeOfValue, defaultValue: defaultValue, Position: item.Position}, nil
}

func NewArrayType(a *yamlmeta.Array) (*ArrayType, error) {
	if len(a.Items) != 1 {
		return nil, NewSchemaError("Invalid schema - wrong number of items in array definition", schemaAssertionError{
			position: a.Position,
			expected: "exactly 1 array item, of the desired type",
			found:    fmt.Sprintf("%d array items", len(a.Items)),
			hints:    []string{"in schema, the one item of the array implies the type of its elements.", "in schema, the default value for an array is always an empty list.", "default values can be overridden via a data values overlay."},
		})
	}

	arrayItemType, err := NewArrayItemType(a.Items[0])
	if err != nil {
		return nil, err
	}

	return &ArrayType{ItemsType: arrayItemType, defaultValue: &yamlmeta.Array{}, Position: a.Position}, nil
}

func NewArrayItemType(item *yamlmeta.ArrayItem) (*ArrayItemType, error) {
	typeOfValue, err := getType(item)
	if err != nil {
		return nil, err
	}

	defaultValue, err := getValue(item, typeOfValue)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &ArrayItemType{ValueType: typeOfValue, defaultValue: defaultValue, Position: item.GetPosition()}, nil
}

func getType(node yamlmeta.Node) (yamlmeta.Type, error) {
	var typeOfValue yamlmeta.Type

	anns, err := collectTypeAnnotations(node)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}
	typeOfValue, err = getTypeFromAnnotations(anns, node.GetPosition())
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}

	if typeOfValue == nil {
		typeOfValue, err = inferTypeFromValue(node.GetValues()[0], node.GetPosition())
		if err != nil {
			return nil, err
		}
	}
	docAnns, err := collectDocumentationAnnotations(node)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}
	for _, ann := range docAnns {
		if desc, ok := ann.(*DescriptionAnnotation); ok {
			typeOfValue.SetDescription(desc.description)
		}
	}

	err = valueTypeAllowsItemValue(typeOfValue, node.GetValues()[0], node.GetPosition())
	if err != nil {
		return nil, err
	}

	return typeOfValue, nil
}

func getValue(node yamlmeta.Node, t yamlmeta.Type) (interface{}, error) {
	anns, err := collectValueAnnotations(node, t)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}

	for _, ann := range anns {
		if defaultAnn, ok := ann.(*DefaultAnnotation); ok {
			return getValueFromAnn(defaultAnn, t)
		}
	}

	if _, ok := t.(*AnyType); ok {
		return node.GetValues()[0], nil
	}

	return t.GetDefaultValue(), nil
}

// getValueFromAnn extracts the value from the annotation and validates its type
func getValueFromAnn(defaultAnn *DefaultAnnotation, t yamlmeta.Type) (interface{}, error) {
	var typeCheck yamlmeta.TypeCheck

	defaultValue := defaultAnn.Val()
	if node, ok := defaultValue.(yamlmeta.Node); ok {
		defaultValue = node.DeepCopyAsInterface()
		typeCheck = t.AssignTypeTo(defaultValue.(yamlmeta.Typeable))
	} else {
		typeCheck = t.CheckType(&yamlmeta.Scalar{Value: defaultValue, Position: t.GetDefinitionPosition()})
	}
	if typeCheck.HasViolations() {
		return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v is wrong type", AnnotationDefault), typeCheck.Violations...)
	}

	return defaultValue, nil
}

func inferTypeFromValue(value interface{}, position *filepos.Position) (yamlmeta.Type, error) {
	switch typedContent := value.(type) {
	case *yamlmeta.Document:
		docType, err := NewDocumentType(typedContent)
		if err != nil {
			return nil, err
		}
		return docType, nil
	case *yamlmeta.Map:
		mapType, err := NewMapType(typedContent)
		if err != nil {
			return nil, err
		}
		return mapType, nil
	case *yamlmeta.Array:
		arrayType, err := NewArrayType(typedContent)
		if err != nil {
			return nil, err
		}
		return arrayType, nil
	case string:
		return &ScalarType{ValueType: *new(string), defaultValue: typedContent, Position: position}, nil
	case float64:
		return &ScalarType{ValueType: *new(float64), defaultValue: typedContent, Position: position}, nil
	case int, int64, uint64:
		return &ScalarType{ValueType: *new(int), defaultValue: typedContent, Position: position}, nil
	case bool:
		return &ScalarType{ValueType: *new(bool), defaultValue: typedContent, Position: position}, nil
	case nil:
		return nil, nil
	}

	return nil, fmt.Errorf("Expected value '%s' to be a map, array, or scalar, but was %T", value, value)
}

func valueTypeAllowsItemValue(explicitType yamlmeta.Type, itemValue interface{}, position *filepos.Position) error {
	switch explicitType.(type) {
	case *AnyType:
		return nil
	default:
		if itemValue == nil {
			return NewSchemaError("Invalid schema - null value not allowed here", schemaAssertionError{
				position: position,
				expected: "non-null value",
				found:    "null value",
				hints:    []string{"in YAML, omitting a value implies null.", "to set the default value to null, annotate with @schema/nullable.", "to allow any value, annotate with @schema/type any=True."},
			})
		}
	}
	return nil
}

type ExtractLibRefs interface {
	FromAnnotation(template.NodeAnnotations) ([]ref.LibraryRef, error)
}

func getSchemaLibRef(libRefs ExtractLibRefs, doc *yamlmeta.Document) ([]ref.LibraryRef, error) {
	anns := template.NewAnnotations(doc)
	libRef, err := libRefs.FromAnnotation(anns)
	if err != nil {
		return nil, err
	}
	return libRef, nil
}

func (s *DocumentSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return s.DocType.AssignTypeTo(typeable)
}

func (s *DocumentSchema) DefaultDataValues() *yamlmeta.Document {
	return s.defaultDVs
}

// GetDocumentType returns a reference to the DocumentType that is the root of this Schema.
func (s *DocumentSchema) GetDocumentType() *DocumentType {
	return s.DocType
}

func (s *DocumentSchema) deepCopy() *DocumentSchema {
	return &DocumentSchema{
		Source:     s.Source.DeepCopy(),
		defaultDVs: s.defaultDVs.DeepCopy(),
		DocType:    s.DocType,
	}
}

func (e *DocumentSchemaEnvelope) Source() *yamlmeta.Document {
	return e.Doc.Source
}

func (e *DocumentSchemaEnvelope) Desc() string {
	var desc []string
	for _, refPiece := range e.originalLibRef {
		desc = append(desc, refPiece.AsString())
	}
	return fmt.Sprintf("Schema belonging to library '%s%s' on %s", "@",
		strings.Join(desc, "@"), e.Source().Position.AsString())
}

func (e *DocumentSchemaEnvelope) IsUsed() bool { return e.used }

func (e *DocumentSchemaEnvelope) IntendedForAnotherLibrary() bool {
	return len(e.libRef) > 0
}

func (e *DocumentSchemaEnvelope) UsedInLibrary(expectedRefPiece ref.LibraryRef) (*DocumentSchemaEnvelope, bool) {
	if !e.IntendedForAnotherLibrary() {
		e.markUsed()

		return e.deepCopy(), true
	}

	if !e.libRef[0].Matches(expectedRefPiece) {
		return nil, false
	}
	e.markUsed()
	childSchemaProcessing := e.deepCopy()
	childSchemaProcessing.libRef = childSchemaProcessing.libRef[1:]
	return childSchemaProcessing, !childSchemaProcessing.IntendedForAnotherLibrary()
}

func (e *DocumentSchemaEnvelope) markUsed() { e.used = true }

func (e *DocumentSchemaEnvelope) deepCopy() *DocumentSchemaEnvelope {
	var copiedPieces []ref.LibraryRef
	copiedPieces = append(copiedPieces, e.libRef...)
	return &DocumentSchemaEnvelope{
		Doc:            e.Doc.deepCopy(),
		originalLibRef: e.originalLibRef,
		libRef:         copiedPieces,
	}
}
