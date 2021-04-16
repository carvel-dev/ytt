// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationLibraryRef = "library/ref"

	schemaLibrarySep            = "@"
	schemaLibraryAliasIndicator = "~"
)

type AnySchema struct {
}

type NullSchema struct {
}

type DocumentSchema struct {
	Name       string
	Source     *yamlmeta.Document
	defaultDVs *yamlmeta.Document
	Allowed    *DocumentType

	originalLibRef []LibRefPiece
	libRef         []LibRefPiece
}

type LibRefPiece struct {
	Path  string
	Alias string
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

	libRef, err := getSchemaLibRef(doc)
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
	valueType, err := newCollectionItemValueType(item.Value, item.Position)
	if err != nil {
		return nil, err
	}

	defaultValue := item.Value
	if _, ok := item.Value.(*yamlmeta.Array); ok {
		defaultValue = &yamlmeta.Array{}
	}

	templateAnnotations := template.NewAnnotations(item)
	if _, nullable := templateAnnotations[AnnotationSchemaNullable]; nullable {
		defaultValue = nil
	} else if valueType == nil {
		return nil, NewInvalidSchemaError(item,
			"null value is not allowed in schema (no type can be inferred from it)",
			"to default to null, specify a value of the desired type and annotate with @schema/nullable")
	}
	annotations := make(TypeAnnotations)
	for key, val := range templateAnnotations {
		annotations[key] = val
	}

	return &MapItemType{Key: item.Key, ValueType: valueType, DefaultValue: defaultValue, Position: item.Position, Annotations: annotations}, nil
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
	valueType, err := newCollectionItemValueType(item.Value, item.Position)
	if err != nil {
		return nil, err
	}

	annotations := template.NewAnnotations(item)

	if _, found := annotations[AnnotationSchemaNullable]; found {
		return nil, NewInvalidSchemaError(item, fmt.Sprintf("@%s is not supported on array items", AnnotationSchemaNullable), "")
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
		if anns.Has(AnnotationSchemaNullable) {
			typedNode.Value = nil
		}
		if valueAsANode, ok := typedNode.Value.(yamlmeta.Node); ok {
			setDefaultValues(valueAsANode)
		}
	case *yamlmeta.Array:
		typedNode.Items = []*yamlmeta.ArrayItem{}
	}
}

func getSchemaLibRef(doc *yamlmeta.Document) ([]LibRefPiece, error) {
	var libRef []LibRefPiece
	anns := template.NewAnnotations(doc)

	if hasLibAnn := anns.Has(AnnotationLibraryRef); hasLibAnn {
		libArgs := anns.Args(AnnotationLibraryRef)
		if l := libArgs.Len(); l != 1 {
			return nil, fmt.Errorf("Expected %s annotation to have one arg, got %d", AnnotationLibraryRef, l)
		}
		argString, err := core.NewStarlarkValue(libArgs[0]).AsString()
		if err != nil {
			return nil, err
		}
		if argString == "" {
			return nil, fmt.Errorf("Expected library ref to not be empty")
		}

		if !strings.HasPrefix(argString, schemaLibrarySep) {
			return nil, fmt.Errorf("Expected library ref to start with '%s'", schemaLibrarySep)
		}

		for _, refPiece := range strings.Split(argString, schemaLibrarySep)[1:] {
			pathAndAlias := strings.Split(refPiece, schemaLibraryAliasIndicator)
			switch l := len(pathAndAlias); {
			case l == 1:
				libRef = append(libRef, LibRefPiece{Path: pathAndAlias[0]})

			case l == 2:
				if pathAndAlias[1] == "" {
					return nil, fmt.Errorf("Expected library alias to not be empty")
				}

				libRef = append(libRef, LibRefPiece{Path: pathAndAlias[0], Alias: pathAndAlias[1]})

			default:
				return nil, fmt.Errorf("Expected library ref to have form: '@path', '@~alias', or '@path~alias', got: '%s'", argString)
			}
		}

	}
	return libRef, nil
}

func (as *AnySchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{}
}

func (n NullSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{}
}

func (s *DocumentSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return s.Allowed.AssignTypeTo(typeable)
}

func (as *AnySchema) DefaultDataValues() *yamlmeta.Document {
	return nil
}

func (n NullSchema) DefaultDataValues() *yamlmeta.Document {
	return nil
}

func (s *DocumentSchema) DefaultDataValues() *yamlmeta.Document {
	return s.defaultDVs
}

func (as *AnySchema) ValidateWithValues(valuesFilesCount int) error {
	return nil
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

func (s *DocumentSchema) HasLibRef() bool {
	return len(s.libRef) > 0
}

func (s *DocumentSchema) UsedInLibrary(expectedRefPiece LibRefPiece) {
	if len(s.libRef) == 0 {
		return
	}
	if s.libRef[0].Path != expectedRefPiece.Path {
		return
	}
	s.libRef = nil
}
