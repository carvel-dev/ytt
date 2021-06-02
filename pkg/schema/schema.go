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

type NullSchema struct {
}

type DocumentSchema struct {
	Source     *yamlmeta.Document
	defaultDVs *yamlmeta.Document
	Allowed    *DocumentType
}

type DocumentSchemaEnvelope struct {
	Doc *DocumentSchema

	used           bool
	originalLibRef []ref.LibraryRef
	libRef         []ref.LibraryRef
}

func NewDocumentSchema(doc *yamlmeta.Document) (*DocumentSchema, error) {
	docType, err := NewDocumentType(doc)
	if err != nil {
		return nil, err
	}

	schemaDVs, err := defaultDataValues(doc)
	if err != nil {
		return nil, err
	}

	return &DocumentSchema{
		Source:     doc,
		defaultDVs: schemaDVs,
		Allowed:    docType,
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

func NewPermissiveSchema() *DocumentSchema {
	return &DocumentSchema{
		Source: &yamlmeta.Document{},
		Allowed: &DocumentType{
			ValueType: &AnyType{}},
	}
}

func NewDocumentType(doc *yamlmeta.Document) (*DocumentType, error) {
	docType := &DocumentType{Source: doc, Position: doc.Position}
	switch typedDocumentValue := doc.Value.(type) {
	case *yamlmeta.Map:
		valueType, err := NewMapType(typedDocumentValue)
		if err != nil {
			return nil, err
		}

		docType.ValueType = valueType
	case *yamlmeta.Array:
		valueType, err := NewArrayType(typedDocumentValue)
		if err != nil {
			return nil, err
		}
		docType.ValueType = valueType
	}
	return docType, nil
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
	var valueType yamlmeta.Type

	anns, err := collectAnnotations(item)
	if err != nil {
		return nil, err
	}
	typeFromAnns := convertAnnotationsToSingleType(anns)
	if typeFromAnns != nil {
		valueType = typeFromAnns
	} else {
		valueType, err = newCollectionItemValueType(item.Value, item.GetPosition())
		if err != nil {
			return nil, err
		}
	}

	if valueType == nil {
		return nil, NewSchemaError(schemaAssertionError{
			position:    item.GetPosition(),
			description: "null value not allowed here",
			expected:    "non-null value",
			found:       "null value",
			hints:       []string{"in YAML, omitting a value implies null.", "to set the default value to null, annotate with @schema/nullable.", "to allow any value, annotate with @schema/type any=True."},
		})
	}

	defaultValue := item.Value
	if _, ok := item.Value.(*yamlmeta.Array); ok {
		defaultValue = &yamlmeta.Array{}
	}

	return &MapItemType{Key: item.Key, ValueType: valueType, DefaultValue: defaultValue, Position: item.Position}, nil
}

func NewArrayType(a *yamlmeta.Array) (*ArrayType, error) {
	if len(a.Items) != 1 {
		return nil, NewSchemaError(schemaAssertionError{
			position:    a.Position,
			description: "wrong number of items in array definition",
			expected:    "exactly 1 array item, of the desired type",
			found:       fmt.Sprintf("%d array items", len(a.Items)),
			hints:       []string{"in schema, the one item of the array implies the type of its elements.", "in schema, the default value for an array is always an empty list.", "default values can be overridden via a data values overlay."},
		})
	}

	arrayItemType, err := NewArrayItemType(a.Items[0])
	if err != nil {
		return nil, err
	}

	return &ArrayType{ItemsType: arrayItemType, Position: a.Position}, nil
}

func NewArrayItemType(item *yamlmeta.ArrayItem) (*ArrayItemType, error) {
	var valueType yamlmeta.Type

	anns, err := collectAnnotations(item)
	if err != nil {
		return nil, err
	}
	typeFromAnns := convertAnnotationsToSingleType(anns)
	if typeFromAnns != nil {
		if _, ok := typeFromAnns.(*NullType); ok {
			return nil, NewSchemaError(schemaAssertionError{
				position:    item.Position,
				description: "@schema/nullable is not supported on array items",
				expected:    "a valid annotation",
				found:       fmt.Sprintf("@%v", AnnotationNullable),
				hints:       []string{"Remove the @schema/nullable annotation from array item"},
			})
		}
		valueType = typeFromAnns
	} else {
		valueType, err = newCollectionItemValueType(item.Value, item.Position)
		if err != nil {
			return nil, err
		}
	}

	return &ArrayItemType{ValueType: valueType, Position: item.Position}, nil
}

func newCollectionItemValueType(collectionItemValue interface{}, position *filepos.Position) (yamlmeta.Type, error) {
	switch typedContent := collectionItemValue.(type) {
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
		return &ScalarType{Value: *new(string), Position: position}, nil
	case float64:
		return &ScalarType{Value: *new(float64), Position: position}, nil
	case int, int64, uint64:
		return &ScalarType{Value: *new(int), Position: position}, nil
	case bool:
		return &ScalarType{Value: *new(bool), Position: position}, nil
	case nil:
		return nil, nil
	}

	return nil, fmt.Errorf("Collection item type did not match any known types")
}

func defaultDataValues(doc *yamlmeta.Document) (*yamlmeta.Document, error) {
	docCopy := doc.DeepCopyAsNode()
	for _, value := range docCopy.GetValues() {
		if valueAsANode, ok := value.(yamlmeta.Node); ok {
			setDefaultValues(valueAsANode)
		}
	}

	return docCopy.(*yamlmeta.Document), nil
}

func setDefaultValues(node yamlmeta.Node) {
	switch typedNode := node.(type) {
	case *yamlmeta.Map:
		for _, value := range typedNode.Items {
			setDefaultValues(value)
		}
	case *yamlmeta.MapItem:
		anns := template.NewAnnotations(typedNode)
		if anns.Has(AnnotationNullable) {
			typedNode.Value = nil
		}
		if valueAsANode, ok := typedNode.Value.(yamlmeta.Node); ok {
			setDefaultValues(valueAsANode)
		}
	case *yamlmeta.Array:
		typedNode.Items = []*yamlmeta.ArrayItem{}
	}
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

func (n NullSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{}
}

func (s *DocumentSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return s.Allowed.AssignTypeTo(typeable)
}

func (n NullSchema) DefaultDataValues() *yamlmeta.Document {
	return nil
}

func (s *DocumentSchema) DefaultDataValues() *yamlmeta.Document {
	return s.defaultDVs
}

func (s *DocumentSchema) deepCopy() *DocumentSchema {
	return &DocumentSchema{
		Source:     s.Source.DeepCopy(),
		defaultDVs: s.defaultDVs.DeepCopy(),
		Allowed:    s.Allowed,
	}
}

func (n NullSchema) ValidateWithValues(valuesFilesCount int) error {
	if valuesFilesCount > 0 {
		return fmt.Errorf("Schema feature is enabled but no schema document was provided")
	}
	return nil
}

func (s *DocumentSchema) ValidateWithValues(valuesFilesCount int) error {
	return nil
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
