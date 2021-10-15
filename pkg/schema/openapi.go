// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func NewOpenAPIDocument(docType *DocumentType) *yamlmeta.Document {
	openAPIProperties := calculateProperties(docType)

	headerDoc := yamlmeta.Document{Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
		{Key: "openapi", Value: "3.0.0"},
		{Key: "info", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "version", Value: "0.1.0"},
			{Key: "title", Value: "Openapi schema generated from ytt Data Values Schema"},
		}},
		},
		{Key: "paths", Value: &yamlmeta.Map{}},
		{Key: "components", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "schemas", Value: openAPIProperties},
		}},
		},
	}}}
	return &headerDoc
}

func calculateProperties(schemaVal interface{}) yamlmeta.Node {
	switch typedValue := schemaVal.(type) {
	case *DocumentType:
		return calculateProperties(typedValue.GetValueType())
	case *MapType:
		typeString := convertTypeToString(typedValue.String())

		var properties []*yamlmeta.MapItem
		for _, i := range typedValue.Items {
			mi := yamlmeta.MapItem{Key: i.Key, Value: calculateProperties(i.GetValueType())}
			properties = append(properties, &mi)
		}
		newMap := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: typeString},
			{Key: "additionalProperties", Value: false},
			{Key: "properties", Value: &yamlmeta.Map{Items: properties}},
		}}
		return &newMap
	case *ScalarType:
		typeString := convertTypeToString(typedValue.String())
		defaultVal := typedValue.GetDefaultValue()
		newMap := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: typeString},
			{Key: "default", Value: defaultVal},
		}}
		if typedValue.String() == "float" {
			newMap.Items = append(newMap.Items, &yamlmeta.MapItem{Key: "format", Value: "float"})
		}
		return &newMap

	}
	return nil
}

func convertTypeToString(typeString string) string {
	switch typeString {
	case "string":
		return "string"
	case "boolean":
		return "boolean"
	case "integer":
		return "integer"
	case "float":
		return "number"
	case "map":
		return "object"
	case "array":
		return "array"
	default:
		return typeString
	}
}
