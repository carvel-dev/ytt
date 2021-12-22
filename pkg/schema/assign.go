// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/yamlmeta"
)

// AssignTypeTo assigns this schema metadata to `node`.
// If `node` is not a yamlmeta.Document, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (t *DocumentType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	doc, ok := node.(*yamlmeta.Document)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(node, t))
		return
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
// If `node` is not a yamlmeta.Map, `chk` contains a violation describing the mismatch
// If `node`'s yamlmeta.MapItem's cannot be assigned their corresponding MapItemType, `chk` contains a violation describing the mismatch
// If `node` is missing any yamlmeta.MapItem's specified in this MapType, they are added to `node`.
func (m *MapType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	mapNode, ok := node.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, m))
		return
	}
	var foundKeys []interface{}
	SetType(node.(yamlmeta.Node), m)
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
	return
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

// AssignTypeTo assigns this schema metadata to `node`.
// If `node` is not a yamlmeta.MapItem, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (t *MapItemType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	mapItem, ok := node.(*yamlmeta.MapItem)
	if !ok {
		panic(fmt.Sprintf("Maps can only contain map items; this is a %s (attempting to assign %#v to %#v)", yamlmeta.TypeName(node), t, node))
	}
	SetType(node.(yamlmeta.Node), t)
	valueNode, isNode := mapItem.Value.(yamlmeta.Node)
	if isNode {
		childCheck := t.ValueType.AssignTypeTo(valueNode)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return
}

// AssignTypeTo assigns this schema metadata to `node`.
// If `node` is not a yamlmeta.Array, `chk` contains a violation describing the mismatch
// For each `node`'s yamlmeta.ArrayItem's that cannot be assigned this ArrayType's ArrayItemType, `chk` contains a violation describing the mismatch
func (a *ArrayType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	arrayNode, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, a))
		return
	}
	SetType(node.(yamlmeta.Node), a)
	for _, arrayItem := range arrayNode.Items {
		childCheck := a.ItemsType.AssignTypeTo(arrayItem)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	}
	return
}

// AssignTypeTo assigns this schema metadata to `node`.
// If `node` is not a yamlmeta.ArrayItem, `chk` contains a violation describing the mismatch
// If `node`'s value is not of the same structure (i.e. yamlmeta.Node type), `chk` contains a violation describing this mismatch
func (a *ArrayItemType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	arrayItem, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		panic(fmt.Sprintf("Arrays can only contain array items; this is a %s (attempting to assign %#v to %#v)", yamlmeta.TypeName(node), a, node))
	}
	SetType(node.(yamlmeta.Node), a)
	valueNode, isNode := arrayItem.Value.(yamlmeta.Node)
	if isNode {
		childCheck := a.ValueType.AssignTypeTo(valueNode)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return
}

// AssignTypeTo returns a violation describing the type mismatch, given that ScalarType will never accept a yamlmeta.Node
func (s *ScalarType) AssignTypeTo(node yamlmeta.Node) TypeCheck {
	return TypeCheck{[]error{NewMismatchedTypeAssertionError(node, s)}}
}

// AssignTypeTo is a no-op given that AnyType allows all types.
func (a AnyType) AssignTypeTo(yamlmeta.Node) (chk TypeCheck) {
	return
}

// AssignTypeTo assigns this NullType's wrapped Type to `node`.
func (n NullType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	childCheck := n.ValueType.AssignTypeTo(node)
	chk.Violations = append(chk.Violations, childCheck.Violations...)
	return
}
