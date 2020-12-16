// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type AnySchema struct {
}

type DocumentSchema struct {
	Name    string
	Source  *yamlmeta.Document
	Allowed *DocumentType
}

func NewDocumentSchema(doc *yamlmeta.Document) (*DocumentSchema, error) {
	docType := &DocumentType{Source: doc}

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
	return &DocumentSchema{
		Name:    "dataValues",
		Source:  doc,
		Allowed: docType,
	}, nil
}

func NewMapType(m *yamlmeta.Map) (*MapType, error) {
	mapType := &MapType{}

	for _, mapItem := range m.Items {
		mapItemType, err := NewMapItemType(mapItem)
		if err != nil {
			return nil, err
		}
		mapType.Items = append(mapType.Items, mapItemType)
	}
	annotations := template.NewAnnotations(m)
	if _, nullable := annotations[AnnotationSchemaNullable]; nullable {
		mapType.Items = nil
	}

	return mapType, nil
}

func NewMapItemType(item *yamlmeta.MapItem) (*MapItemType, error) {
	valueType, err := newCollectionItemValueType(item.Value)
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
	}
	annotations := make(TypeAnnotations)
	for key, val := range templateAnnotations {
		annotations[key] = val
	}

	return &MapItemType{Key: item.Key, ValueType: valueType, DefaultValue: defaultValue, Position: item.Position, Annotations: annotations}, nil
}

func NewArrayType(a *yamlmeta.Array) (*ArrayType, error) {
	// These really are distinct use cases. In the empty list, perhaps the user is unaware that arrays must be typed. In the >1 scenario, they may be expecting the given items to be the defaults.
	if len(a.Items) == 0 {
		return nil, fmt.Errorf("Expected one item in array (describing the type of its elements) at %s", a.Position.AsCompactString())
	}
	if len(a.Items) > 1 {
		return nil, fmt.Errorf("Expected one item (found %v) in array (describing the type of its elements) at %s", len(a.Items), a.Position.AsCompactString())
	}

	arrayItemType, err := NewArrayItemType(a.Items[0])
	if err != nil {
		return nil, err
	}

	return &ArrayType{ItemsType: arrayItemType}, nil
}

func NewArrayItemType(item *yamlmeta.ArrayItem) (*ArrayItemType, error) {
	valueType, err := newCollectionItemValueType(item.Value)
	if err != nil {
		return nil, err
	}

	annotations := template.NewAnnotations(item)

	if _, found := annotations[AnnotationSchemaNullable]; found {
		return nil, fmt.Errorf("Array items cannot be annotated with #@schema/nullable (%s). If this behaviour would be valuable, please submit an issue on https://github.com/vmware-tanzu/carvel-ytt", item.GetPosition().AsCompactString())
	}

	return &ArrayItemType{ValueType: valueType}, nil
}

func newCollectionItemValueType(collectionItemValue interface{}) (yamlmeta.Type, error) {
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
		return &ScalarType{Type: *new(string)}, nil
	case int:
		return &ScalarType{Type: *new(int)}, nil
	case bool:
		return &ScalarType{Type: *new(bool)}, nil
	}

	return nil, fmt.Errorf("Collection item type did not match any known types")
}

func (as *AnySchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{}
}

func (s *DocumentSchema) AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return s.Allowed.AssignTypeTo(typeable)
}
