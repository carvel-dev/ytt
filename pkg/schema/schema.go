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
	Name       string
	Source     *yamlmeta.Document
	defaultDVs *yamlmeta.Document
	Allowed    *DocumentType
	used       bool

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

	libRef, err := getSchemaLibRef(ref.LibraryRefExtractor{}, doc)
	if err != nil {
		return nil, err
	}

	return &DocumentSchema{
		Name:           "dataValues",
		Source:         doc,
		defaultDVs:     schemaDVs,
		Allowed:        docType,
		originalLibRef: libRef,
		libRef:         libRef,
	}, nil
}

func NewPermissiveSchema() *DocumentSchema {
	return &DocumentSchema{
		Name:   "anyDataValues",
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
		return nil, NewInvalidSchemaError(item,
			"null value is not allowed in schema (no type can be inferred from it)",
			[]string{
				"in YAML, omitting a value implies null.",
				"to set the default value to null, annotate with @schema/nullable.",
				"to allow any value, annotate with @schema/type any=True.",
			})
	}

	defaultValue := item.Value
	if _, ok := item.Value.(*yamlmeta.Array); ok {
		defaultValue = &yamlmeta.Array{}
	}

	return &MapItemType{Key: item.Key, ValueType: valueType, DefaultValue: defaultValue, Position: item.Position}, nil
}

func NewArrayType(a *yamlmeta.Array) (*ArrayType, error) {
	// what's most useful to hint at depends on the author's input.
	if len(a.Items) == 0 {
		// assumption: the user likely does not understand that the shape of the elements are dependent on this item
		return nil, NewInvalidArrayDefinitionError(a, "in a schema, the item of an array defines the type of its elements; its default value is an empty list")
	}
	if len(a.Items) > 1 {
		// assumption: the user wants to supply defaults and (incorrectly) assumed they should go in schema
		return nil, NewInvalidArrayDefinitionError(a, "to add elements to the default value of an array (i.e. an empty list), declare them in a @data/values document")
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
			return nil, NewInvalidSchemaError(item, fmt.Sprintf("@%s is not supported on array items", AnnotationNullable), nil)
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
	case int:
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

func (n NullSchema) ValidateWithValues(valuesFilesCount int) error {
	if valuesFilesCount > 0 {
		return fmt.Errorf("Schema feature is enabled but no schema document was provided")
	}
	return nil
}

func (s *DocumentSchema) ValidateWithValues(valuesFilesCount int) error {
	return nil
}

func (s *DocumentSchema) deepCopy() *DocumentSchema {
	var copiedPieces []ref.LibraryRef
	copiedPieces = append(copiedPieces, s.libRef...)
	return &DocumentSchema{
		Name:           s.Name,
		Source:         s.Source.DeepCopy(),
		defaultDVs:     s.defaultDVs.DeepCopy(),
		Allowed:        s.Allowed,
		originalLibRef: s.originalLibRef,
		libRef:         copiedPieces,
	}
}

func (s *DocumentSchema) Desc() string {
	var desc []string
	for _, refPiece := range s.originalLibRef {
		desc = append(desc, refPiece.AsString())
	}
	return fmt.Sprintf("Schema belonging to library '%s%s' on %s", "@",
		strings.Join(desc, "@"), s.Source.Position.AsString())
}

func (s *DocumentSchema) IsUsed() bool { return s.used }
func (s *DocumentSchema) markUsed()    { s.used = true }

func (s *DocumentSchema) IntendedForAnotherLibrary() bool {
	return len(s.libRef) > 0
}

func (s *DocumentSchema) UsedInLibrary(expectedRefPiece ref.LibraryRef) (*DocumentSchema, bool) {
	if !s.IntendedForAnotherLibrary() {
		s.markUsed()

		return s.deepCopy(), true
	}

	if !s.libRef[0].Matches(expectedRefPiece) {
		return nil, false
	}
	s.markUsed()
	childSchema := s.deepCopy()
	childSchema.libRef = childSchema.libRef[1:]
	return childSchema, !childSchema.IntendedForAnotherLibrary()
}
