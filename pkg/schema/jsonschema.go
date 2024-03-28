// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"sort"

	"carvel.dev/ytt/pkg/yamlmeta"
)

// JSONSchemaDocument holds the document type used for creating an JSON Schema document
type JSONSchemaDocument struct {
	docType *DocumentType
}

// NewJSONSchemaDocument creates an instance of an OpenAPIDocument based on the given DocumentType
func NewJSONSchemaDocument(docType *DocumentType) *JSONSchemaDocument {
	return &JSONSchemaDocument{docType}
}

// AsDocument generates a new AST of this OpenAPI v3.0.x document, populating the `schemas:` section with the
// type information contained in `docType`.
func (j *JSONSchemaDocument) AsDocument() *yamlmeta.Document {
	jsonSchemaProperties := j.calculateProperties(j.docType)

	jsonSchemaProperties.Items = append(
		[]*yamlmeta.MapItem{
			{Key: "$schema", Value: "https://json-schema.org/draft/2020-12/schema"},
			{Key: "$id", Value: "https://example.biz/schema/ytt/data-values.json"},
			{Key: "description", Value: "Schema for data values, generated by ytt"},
		},
		jsonSchemaProperties.Items...,
	)

	return &yamlmeta.Document{Value: jsonSchemaProperties}
}

func (j *JSONSchemaDocument) calculateProperties(schemaVal interface{}) *yamlmeta.Map {
	switch typedValue := schemaVal.(type) {
	case *DocumentType:
		return j.calculateProperties(typedValue.GetValueType())
	case *MapType:
		var items openAPIKeys
		items = append(items, collectDocumentation(typedValue)...)
		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: "object"})
		items = append(items, &yamlmeta.MapItem{Key: additionalPropsProp, Value: false})

		var properties []*yamlmeta.MapItem
		for _, i := range typedValue.Items {
			mi := yamlmeta.MapItem{Key: i.Key, Value: j.calculateProperties(i.GetValueType())}
			properties = append(properties, &mi)
		}
		items = append(items, &yamlmeta.MapItem{Key: propertiesProp, Value: &yamlmeta.Map{Items: properties}})

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *ArrayType:
		var items openAPIKeys
		items = append(items, collectDocumentation(typedValue)...)

		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: "array"})
		items = append(items, &yamlmeta.MapItem{Key: defaultProp, Value: typedValue.GetDefaultValue()})

		valueType := typedValue.GetValueType().(*ArrayItemType)
		properties := j.calculateProperties(valueType.GetValueType())
		items = append(items, &yamlmeta.MapItem{Key: itemsProp, Value: properties})

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *ScalarType:
		var items openAPIKeys
		items = append(items, collectDocumentation(typedValue)...)
		items = append(items, &yamlmeta.MapItem{Key: defaultProp, Value: typedValue.GetDefaultValue()})

		typeString := schemaTypeFor(typedValue)
		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: typeString})

		items = append(items, convertValidations(typedValue.GetValidationMap())...)

		if typedValue.String() == "float" {
			items = append(items, &yamlmeta.MapItem{Key: formatProp, Value: "float"})
		}

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *NullType:
		var items openAPIKeys
		items = append(items, collectDocumentation(typedValue)...)

		properties := j.calculateProperties(typedValue.GetValueType())
		// we need to append the "null" type to the list of types
		for i := 0; i < len(properties.Items); i++ {
			if properties.Items[i].Key == "type" {
				// this is a map item with a single valeu, we now need to convert it to an array
				typeItem := properties.Items[i]
				nullableItem := &yamlmeta.MapItem{Key: "type", Value: &yamlmeta.Array{Items: []*yamlmeta.ArrayItem{
					{Value: typeItem.Value}, // this is the original type
					{Value: "null"},
				}}}

				items = append(items, nullableItem)
			} else {
				items = append(items, properties.Items[i])
			}
		}

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *AnyType:
		var items openAPIKeys
		items = append(items, collectDocumentation(typedValue)...)
		// AnyType must allow all value types, and need to explicitly list them for json schema
		items = append(items, &yamlmeta.MapItem{Key: typeProp,
			Value: &yamlmeta.Array{Items: []*yamlmeta.ArrayItem{
				{Value: "null"},
				{Value: "string"},
				{Value: "number"},
				{Value: "object"},
				{Value: "array"},
				{Value: "boolean"},
			}},
		})
		items = append(items, &yamlmeta.MapItem{Key: defaultProp, Value: typedValue.GetDefaultValue()})

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	default:
		panic(fmt.Sprintf("Unrecognized type %T", schemaVal))
	}
}