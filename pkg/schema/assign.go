// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"carvel.dev/ytt/pkg/validations"
	"carvel.dev/ytt/pkg/yamlmeta"
)

// AssignTypeTo assigns this schema metadata to `node`.
//
// If `node` is not a yamlmeta.Document, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (t *DocumentType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	doc, ok := node.(*yamlmeta.Document)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, t))
		return chk
	}
	SetType(doc, t)
	valueNode, isNode := doc.Value.(yamlmeta.Node)
	if isNode {
		childCheck := t.ValueType.AssignTypeTo(valueNode)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is a scalar
	return chk
}

// AssignTypeTo assigns this schema metadata to `node`.
//
// If `node` is not a yamlmeta.Map, `chk` contains a violation describing the mismatch
// If `node`'s yamlmeta.MapItem's cannot be assigned their corresponding MapItemType, `chk` contains a violation describing the mismatch
// If `node` is missing any yamlmeta.MapItem's specified in this MapType, they are added to `node`.
func (m *MapType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	mapNode, ok := node.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, m))
		return chk
	}
	var foundKeys []interface{}
	SetType(node, m)
	for _, mapItem := range mapNode.Items {
		for _, itemType := range m.Items {
			if mapItem.Key == itemType.Key {
				foundKeys = append(foundKeys, itemType.Key)
				childCheck := itemType.AssignTypeTo(mapItem)
				chk.Violations = append(chk.Violations, childCheck.Violations...)
				break
			}
		}
	}

	m.applySchemaDefaults(foundKeys, chk, mapNode)
	return chk
}

func (m *MapType) applySchemaDefaults(foundKeys []interface{}, chk TypeCheck, mapNode *yamlmeta.Map) {
	for _, item := range m.Items {
		if contains(foundKeys, item.Key) {
			continue
		}

		val := item.GetDefaultValue()
		childCheck := item.AssignTypeTo(val.(*yamlmeta.MapItem))
		chk.Violations = append(chk.Violations, childCheck.Violations...)
		err := mapNode.AddValue(val)
		if err != nil {
			panic(fmt.Sprintf("Internal inconsistency: adding map item: %s", err))
		}
	}
}

func contains(haystack []interface{}, needle interface{}) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
}

// AssignTypeTo assigns this schema metadata to `node`.
//
// If `node` is not a yamlmeta.MapItem, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (t *MapItemType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	mapItem, ok := node.(*yamlmeta.MapItem)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, t))
		return chk
	}
	SetType(node, t)
	valueNode, isNode := mapItem.Value.(yamlmeta.Node)
	if isNode {
		childCheck := t.ValueType.AssignTypeTo(valueNode)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return chk
}

// AssignTypeTo assigns this schema metadata to `node`.
//
// If `node` is not a yamlmeta.Array, `chk` contains a violation describing the mismatch
// For each `node`'s yamlmeta.ArrayItem's that cannot be assigned this ArrayType's ArrayItemType, `chk` contains a violation describing the mismatch
func (a *ArrayType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	arrayNode, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, a))
		return chk
	}
	SetType(node, a)
	for _, arrayItem := range arrayNode.Items {
		childCheck := a.ItemsType.AssignTypeTo(arrayItem)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	}
	return chk
}

// AssignTypeTo assigns this schema metadata to `node`.
//
// If `node` is not a yamlmeta.ArrayItem, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (a *ArrayItemType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	arrayItem, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, a))
		return chk
	}
	SetType(node, a)
	valueNode, isNode := arrayItem.Value.(yamlmeta.Node)
	if isNode {
		childCheck := a.ValueType.AssignTypeTo(valueNode)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return chk
}

// AssignTypeTo returns a violation describing the type mismatch, given that ScalarType will never accept a yamlmeta.Node
func (s *ScalarType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	return TypeCheck{[]error{NewMismatchedTypeAssertionError(node, s)}}
}

// AssignTypeTo is a no-op given that AnyType allows all types.
func (a AnyType) AssignTypeTo(yamlmeta.Node) TypeCheck {
	return TypeCheck{}
}

// AssignTypeTo assigns this NullType's wrapped Type to `node`.
func (n NullType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	childCheck := n.ValueType.AssignTypeTo(node)
	chk.Violations = append(chk.Violations, childCheck.Violations...)
	return chk
}

// AssignSchemaValidations implements the visitor interface to set validations from the schema type
type AssignSchemaValidations struct{}

// Visit Extracts the validations from Node's Type and sets them in Node's meta
// This visitor returns nil if node has no assigned type or when the execution is completed
func (AssignSchemaValidations) Visit(node yamlmeta.Node) error {
	if schemaType := GetType(node); schemaType != nil {
		if v := schemaType.GetValidation(); v != nil {
			validations.Set(node, []validations.NodeValidation{*v})
		}
	}
	return nil
}
