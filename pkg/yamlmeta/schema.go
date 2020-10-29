// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
)

type Schema interface {
	AssignType(typeable Typeable) TypeCheck
}

var _ Schema = &AnySchema{}
var _ Schema = &DocumentSchema{}

type AnySchema struct {
}

type DocumentSchema struct {
	Name    string
	Source  *Document
	Allowed *DocumentType
}

func NewDocumentSchema(doc *Document) (*DocumentSchema, error) {
	docType := &DocumentType{Source: doc}

	switch typedDocumentValue := doc.Value.(type) {
	case *Map:
		valueType, err := NewMapType(typedDocumentValue)
		if err != nil {
			return nil, err
		}

		docType.ValueType = valueType
	case *Array:
		valueType, err := NewArrayType(typedDocumentValue)
		if err != nil {
			return nil, err
		}

		docType.ValueType = valueType
	}
	return &DocumentSchema{
		Name:    "dataValues",
		Source:  doc,
		Allowed: docType,
	}, nil
}

func NewMapType(m *Map) (*MapType, error) {
	mapType := &MapType{}

	for _, mapItem := range m.Items {
		mapItemType, err := NewMapItemType(mapItem)
		if err != nil {
			return nil, err
		}
		mapType.Items = append(mapType.Items, mapItemType)
	}
	return mapType, nil
}

func NewMapItemType(item *MapItem) (*MapItemType, error) {
	valueType, err := newCollectionItemValueType(item.Value)
	if err != nil {
		return nil, err
	}
	return &MapItemType{Key: item.Key, ValueType: valueType}, nil
}

func NewArrayType(a *Array) (*ArrayType, error) {
	if len(a.Items) != 1 {
		return nil, fmt.Errorf("Expected only one element in the array to determine type, but found %v", len(a.Items))
	}

	arrayItemType, err := NewArrayItemType(a.Items[0])
	if err != nil {
		return nil, err
	}

	return &ArrayType{ItemsType: arrayItemType}, nil
}

func NewArrayItemType(item *ArrayItem) (*ArrayItemType, error) {
	valueType, err := newCollectionItemValueType(item.Value)
	if err != nil {
		return nil, err
	}
	return &ArrayItemType{ValueType: valueType}, nil
}

func newCollectionItemValueType(collectionItemValue interface{}) (Type, error) {
	switch typedContent := collectionItemValue.(type) {
	case *Map:
		mapType, err := NewMapType(typedContent)
		if err != nil {
			return nil, err
		}
		return mapType, nil
	case *Array:
		arrayType, err := NewArrayType(typedContent)
		if err != nil {
			return nil, err
		}
		return arrayType, nil
	case string:
		return &ScalarType{Type: *new(string)}, nil
	case int:
		return &ScalarType{Type: *new(int)}, nil
	case bool:
		return &ScalarType{Type: *new(bool)}, nil
	}

	return nil, fmt.Errorf("Collection item type did not match any known types")
}

func (as *AnySchema) AssignType(typeable Typeable) TypeCheck { return TypeCheck{} }

func (s *DocumentSchema) AssignType(typeable Typeable) TypeCheck {
	return s.Allowed.AssignTypeTo(typeable)
}
