// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
	"github.com/k14s/ytt/pkg/filepos"
)

type Type interface {
	// Checks whether `value` is permitted within (applicable to map types only)
	CheckAllows(item *MapItem) TypeCheck
	AssignTypeTo(typeable Typeable) TypeCheck
}

var _ Type = (*DocumentType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*MapItemType)(nil)

type Typeable interface {
	// TODO: extract methods common to Node and Typeable to a shared interface?
	GetPosition() *filepos.Position
	GetValues() []interface{} // ie children

	SetType(Type)
}

var _ Typeable = (*Document)(nil)
var _ Typeable = (*Map)(nil)
var _ Typeable = (*MapItem)(nil)

func (n *Document) SetType(t Type) { n.Type = t }
func (n *Map) SetType(t Type)      { n.Type = t }
func (n *MapItem) SetType(t Type)  { n.Type = t }

type DocumentType struct {
	Source    *Document
	ValueType Type // typically one of: MapType, ArrayType, ScalarType
}
type MapType struct {
	Items []*MapItemType
}
type MapItemType struct {
	Key       interface{} // usually a string
	ValueType Type
}

type ScalarType struct {
	Type interface{}
}

func (t *DocumentType) CheckAllows(item *MapItem) TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a Document.")
}
func (m MapItemType) CheckAllows(item *MapItem) TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a MapItem.")
}

func (m ScalarType) CheckAllows(item *MapItem) TypeCheck {
	panic("Attempt to check if a MapItem is allowed as a value of a ScalarType.")
}

func (t *DocumentType) AssignTypeTo(typeable Typeable) (chk TypeCheck) {
	doc, ok := typeable.(*Document)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &Document{}, typeable)}
		return
	}
	typeable.SetType(t)
	typeableChild, ok := doc.Value.(Typeable)
	if ok {
		if t.ValueType != nil {
			childCheck := t.ValueType.AssignTypeTo(typeableChild)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		} else {
			chk.Violations = []string{fmt.Sprintf("Expected node at %s to be %s, but was a %T", typeableChild.GetPosition().AsCompactString(), "nil", typeableChild)}
		}
	} else {

	} // else, at a leaf
	return
}
func (t *MapType) AssignTypeTo(typeable Typeable) (chk TypeCheck) {
	mapNode, ok := typeable.(*Map)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &Map{}, typeable)}
		return
	}
	typeable.SetType(t)
	for _, mapItem := range mapNode.Items {
		for _, itemType := range t.Items {
			if mapItem.Key == itemType.Key {
				childCheck := itemType.AssignTypeTo(mapItem)
				chk.Violations = append(chk.Violations, childCheck.Violations...)
				break
			}
		}
	}
	return
}
func (t *MapItemType) AssignTypeTo(typeable Typeable) (chk TypeCheck) {
	mapItem, ok := typeable.(*MapItem)
	if !ok {
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &MapItem{}, typeable)}
		return
	}
	typeable.SetType(t)
	typeableValue, ok := mapItem.Value.(Typeable)
	if ok {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}

func (t *ScalarType) AssignTypeTo(typeable Typeable) (chk TypeCheck) {
	switch t.Type.(type) {
	case int:
		typeable.SetType(t)
	case string:
		typeable.SetType(t)
	default:
		chk.Violations = []string{fmt.Sprintf("Expected node at %s to be a %T, but was a %T", typeable.GetPosition().AsCompactString(), &ScalarType{}, typeable)}
	}
	return
}

func (t *MapType) AllowsKey(key interface{}) bool {
	for _, item := range t.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func (t *MapType) CheckAllows(item *MapItem) TypeCheck {
	typeCheck := TypeCheck{}

	if !t.AllowsKey(item.Key) {
		typeCheck.Violations = append(typeCheck.Violations, fmt.Sprintf("Map item '%s' at %s is not defined in schema", item.Key, item.Position.AsCompactString()))
	}
	return typeCheck
}
