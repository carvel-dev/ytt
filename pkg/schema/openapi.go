// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/yamlmeta"
)

// OpenAPIDocument TODO:
type OpenAPIDocument struct {
	docType *DocumentType
}

// NewOpenAPIDocument creates an instance of an OpenAPIDocument based on the given DocumentType
func NewOpenAPIDocument(docType *DocumentType) *OpenAPIDocument {
	return &OpenAPIDocument{docType}
}

// AsDocument generates a new AST of this OpenAPI v3.0.x document, populating the `schemas:` section with the
// type information contained in `docType`.
func (o *OpenAPIDocument) AsDocument() *yamlmeta.Document {
	openAPIProperties := o.calculateProperties(o.docType)

	return &yamlmeta.Document{Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
		{Key: "openapi", Value: "3.0.0"},
		{Key: "info", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "version", Value: "0.1.0"},
			{Key: "title", Value: "Openapi schema generated from ytt Data Values Schema"},
		}},
		},
		{Key: "paths", Value: &yamlmeta.Map{}},
		{Key: "components", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "schemas", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
				{Key: "dataValues", Value: openAPIProperties},
			}}},
		}},
		},
	}}}
}

func (o *OpenAPIDocument) calculateProperties(schemaVal interface{}) yamlmeta.Node {
	switch typedValue := schemaVal.(type) {
	case *DocumentType:
		return o.calculateProperties(typedValue.GetValueType())
	case *MapType:
		var properties []*yamlmeta.MapItem
		for _, i := range typedValue.Items {
			mi := yamlmeta.MapItem{Key: i.Key, Value: o.calculateProperties(i.GetValueType())}
			properties = append(properties, &mi)
		}
		newMap := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: "object"},
			{Key: "additionalProperties", Value: false},
			{Key: "properties", Value: &yamlmeta.Map{Items: properties}},
		}}
		return &newMap
	case *ScalarType:
		typeString := o.openAPITypeFor(typedValue)
		defaultVal := typedValue.GetDefaultValue()
		newMap := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: typeString},
			{Key: "default", Value: defaultVal},
		}}
		if typedValue.String() == "float" {
			newMap.Items = append(newMap.Items, &yamlmeta.MapItem{Key: "format", Value: "float"})
		}
		return &newMap
	default:
		panic(fmt.Sprintf("Unrecognized type %T", schemaVal))
	}
	return nil
}

func (o *OpenAPIDocument) openAPITypeFor(astType *ScalarType) string {
	switch astType.ValueType.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		panic(fmt.Sprintf("Unrecognized type: %T", astType.ValueType))
	}
}
