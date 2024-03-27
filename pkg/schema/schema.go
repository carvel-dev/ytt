// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/validations"
	"carvel.dev/ytt/pkg/yamlmeta"
)

// NewDocumentType constructs a complete DocumentType based on the contents of a schema YAML document.
func NewDocumentType(doc *yamlmeta.Document) (*DocumentType, error) {
	typeOfValue, err := getType(doc)
	if err != nil {
		return nil, err
	}

	defaultValue, err := getValue(doc, typeOfValue)
	if err != nil {
		return nil, err
	}

	v, err := getValidation(doc)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &DocumentType{Source: doc, Position: doc.Position, ValueType: typeOfValue, defaultValue: defaultValue, validations: v}, nil
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

	v, err := getValidation(item)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &MapItemType{Key: item.Key, ValueType: typeOfValue, defaultValue: defaultValue, Position: item.Position, validations: v}, nil
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

	v, err := getValidation(item)
	if err != nil {
		return nil, err
	}

	typeOfValue.SetDefaultValue(defaultValue)

	return &ArrayItemType{ValueType: typeOfValue, defaultValue: defaultValue, Position: item.GetPosition(), validations: v}, nil
}

func getType(node yamlmeta.Node) (Type, error) {
	var typeOfValue Type

	anns, err := collectTypeAnnotations(node)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}
	typeOfValue, err = getTypeFromAnnotations(anns)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}

	if typeOfValue == nil {
		typeOfValue, err = InferTypeFromValue(node.GetValues()[0], node.GetPosition())
		if err != nil {
			return nil, err
		}
	}
	err = valueTypeAllowsItemValue(typeOfValue, node.GetValues()[0], node.GetPosition())
	if err != nil {
		return nil, err
	}

	docAnns, err := collectDocumentationAnnotations(node)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}
	err = setDocumentationFromAnns(docAnns, typeOfValue)
	if err != nil {
		return nil, err
	}

	return typeOfValue, nil
}

func getValue(node yamlmeta.Node, t Type) (interface{}, error) {
	anns, err := collectValueAnnotations(node, t)
	if err != nil {
		return nil, NewSchemaError("Invalid schema", err)
	}

	for _, ann := range anns {
		if defaultAnn, ok := ann.(*DefaultAnnotation); ok {
			return getValueFromAnn(defaultAnn, t)
		}
	}

	return t.GetDefaultValue(), nil
}

func getValidation(node yamlmeta.Node) (*validations.NodeValidation, error) {
	validationAnn, err := processValidationAnnotation(node)
	if err != nil {
		return nil, err
	}

	if validationAnn != nil {
		return validationAnn.GetValidation(), nil
	}
	return nil, nil
}

// getValueFromAnn extracts the value from the annotation and validates its type
func getValueFromAnn(defaultAnn *DefaultAnnotation, t Type) (interface{}, error) {
	var typeCheck TypeCheck

	defaultValue := defaultAnn.Val()
	if node, ok := defaultValue.(yamlmeta.Node); ok {
		defaultNode := node.DeepCopyAsNode()
		typeCheck = t.AssignTypeTo(defaultNode)
		if !typeCheck.HasViolations() {
			typeCheck = CheckNode(defaultNode)
		}
		defaultValue = defaultNode
	} else {
		typeCheck = t.CheckType(&yamlmeta.MapItem{Value: defaultValue, Position: t.GetDefinitionPosition()})
	}
	if typeCheck.HasViolations() {
		var violations []error
		// add violating annotation position to error
		for _, err := range typeCheck.Violations {
			if typeCheckAssertionErr, ok := err.(schemaAssertionError); ok {
				typeCheckAssertionErr.annPositions = []*filepos.Position{defaultAnn.GetPosition()}
				typeCheckAssertionErr.found = typeCheckAssertionErr.found + fmt.Sprintf(" (at %v)", defaultAnn.GetPosition().AsCompactString())
				violations = append(violations, typeCheckAssertionErr)
			} else {
				violations = append(violations, err)
			}
		}
		return nil, NewSchemaError(fmt.Sprintf("Invalid schema - @%v is wrong type", AnnotationDefault), violations...)
	}

	return defaultValue, nil
}

// InferTypeFromValue detects and constructs an instance of the Type of `value`.
func InferTypeFromValue(value interface{}, position *filepos.Position) (Type, error) {
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
		return &ScalarType{ValueType: StringType, defaultValue: typedContent, Position: position}, nil
	case float64:
		return &ScalarType{ValueType: FloatType, defaultValue: typedContent, Position: position}, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return &ScalarType{ValueType: IntType, defaultValue: typedContent, Position: position}, nil
	case bool:
		return &ScalarType{ValueType: BoolType, defaultValue: typedContent, Position: position}, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("Expected value '%s' to be a map, array, or scalar, but was %s", value, yamlmeta.TypeName(typedContent))
	}
}

func valueTypeAllowsItemValue(explicitType Type, itemValue interface{}, position *filepos.Position) error {
	switch explicitType.(type) {
	case *AnyType:
		if node, ok := itemValue.(yamlmeta.Node); ok {
			// search children for annotations
			err := yamlmeta.Walk(node, checkForAnnotations{})
			if err != nil {
				return err
			}
		}
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
