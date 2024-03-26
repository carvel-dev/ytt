package schema

import (
	"fmt"
	"sort"

	"carvel.dev/ytt/pkg/yamlmeta"
)

// JsonSchemaDocument holds the document type used for creating an JSON Schema document
type JsonSchemaDocument struct {
	docType *DocumentType
}

// NewOpenAPIDocument creates an instance of an OpenAPIDocument based on the given DocumentType
func NewJsonSchemaDocument(docType *DocumentType) *JsonSchemaDocument {
	return &JsonSchemaDocument{docType}
}

// AsDocument generates a new AST of this OpenAPI v3.0.x document, populating the `schemas:` section with the
// type information contained in `docType`.
func (j *JsonSchemaDocument) AsDocument() *yamlmeta.Document {
	openAPIProperties := j.calculateProperties(j.docType)
	openAPIProperties.Items = append(openAPIProperties.Items, &yamlmeta.MapItem{Key: "$schema", Value: "https://json-schema.org/draft/2020-12/schema"})

	return &yamlmeta.Document{Value: openAPIProperties}
}

func (j *JsonSchemaDocument) calculateProperties(schemaVal interface{}) *yamlmeta.Map {
	switch typedValue := schemaVal.(type) {
	case *DocumentType:
		return j.calculateProperties(typedValue.GetValueType())
	case *MapType:
		var items openAPIKeys
		items = append(items, j.collectDocumentation(typedValue)...)
		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: &yamlmeta.Array{Items: []*yamlmeta.ArrayItem{{Value: "object"}}}})
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
		items = append(items, j.collectDocumentation(typedValue)...)

		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: &yamlmeta.Array{Items: []*yamlmeta.ArrayItem{{Value: "array"}}}})
		items = append(items, &yamlmeta.MapItem{Key: defaultProp, Value: typedValue.GetDefaultValue()})

		valueType := typedValue.GetValueType().(*ArrayItemType)
		properties := j.calculateProperties(valueType.GetValueType())
		items = append(items, &yamlmeta.MapItem{Key: itemsProp, Value: properties})

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *ScalarType:
		var items openAPIKeys
		items = append(items, j.collectDocumentation(typedValue)...)
		items = append(items, &yamlmeta.MapItem{Key: defaultProp, Value: typedValue.GetDefaultValue()})

		typeString := j.jsonSchemaTypeFor(typedValue)

		items = append(items, &yamlmeta.MapItem{Key: typeProp, Value: &yamlmeta.Array{Items: []*yamlmeta.ArrayItem{{Value: typeString}}}})

		items = append(items, convertValidations(typedValue.GetValidationMap())...)

		if typedValue.String() == "float" {
			items = append(items, &yamlmeta.MapItem{Key: formatProp, Value: "float"})
		}

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *NullType:
		var items openAPIKeys
		items = append(items, j.collectDocumentation(typedValue)...)

		properties := j.calculateProperties(typedValue.GetValueType())
		// we need to append the "null" type to the list of types
		for i := 0; i < len(properties.Items); i++ {
			if properties.Items[i].Key == "type" {
				typeItem := properties.Items[i]
				// in json schema nullable is represented as an array of types with "null" (as a string!) as one of the types
				typeItem.Value.(*yamlmeta.Array).Items = append(typeItem.Value.(*yamlmeta.Array).Items, &yamlmeta.ArrayItem{Value: "null"})

				items = append(items, typeItem)
			} else {
				items = append(items, properties.Items[i])
			}
		}

		sort.Sort(items)
		return &yamlmeta.Map{Items: items}
	case *AnyType:
		var items openAPIKeys
		items = append(items, j.collectDocumentation(typedValue)...)
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

func (j *JsonSchemaDocument) collectDocumentation(typedValue Type) []*yamlmeta.MapItem {
	var items []*yamlmeta.MapItem
	if typedValue.GetTitle() != "" {
		items = append(items, &yamlmeta.MapItem{Key: titleProp, Value: typedValue.GetTitle()})
	}
	if typedValue.GetDescription() != "" {
		items = append(items, &yamlmeta.MapItem{Key: descriptionProp, Value: typedValue.GetDescription()})
	}
	if isDeprecated, _ := typedValue.IsDeprecated(); isDeprecated {
		items = append(items, &yamlmeta.MapItem{Key: deprecatedProp, Value: isDeprecated})
	}
	examples := typedValue.GetExamples()
	if len(examples) != 0 {
		items = append(items, &yamlmeta.MapItem{Key: exampleDescriptionProp, Value: examples[0].description})
		items = append(items, &yamlmeta.MapItem{Key: exampleProp, Value: examples[0].example})
	}
	return items
}

func (o *JsonSchemaDocument) jsonSchemaTypeFor(astType *ScalarType) string {
	switch astType.ValueType {
	case StringType:
		return "string"
	case FloatType:
		return "number"
	case IntType:
		return "integer"
	case BoolType:
		return "boolean"
	default:
		panic(fmt.Sprintf("Unrecognized type: %T", astType.ValueType))
	}
}
