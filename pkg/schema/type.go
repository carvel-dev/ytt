// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var _ yamlmeta.Type = (*DocumentType)(nil)
var _ yamlmeta.Type = (*MapType)(nil)
var _ yamlmeta.Type = (*MapItemType)(nil)
var _ yamlmeta.Type = (*ArrayType)(nil)
var _ yamlmeta.Type = (*ArrayItemType)(nil)
var _ yamlmeta.Type = (*AnyType)(nil)
var _ yamlmeta.Type = (*NullType)(nil)
var _ yamlmeta.Type = (*ScalarType)(nil)

type DocumentType struct {
	Source       *yamlmeta.Document
	ValueType    yamlmeta.Type // typically one of: MapType, ArrayType, ScalarType
	Position     *filepos.Position
	defaultValue interface{}
}
type MapType struct {
	Items       []*MapItemType
	Position    *filepos.Position
	description string
}
type MapItemType struct {
	Key          interface{} // usually a string
	ValueType    yamlmeta.Type
	Position     *filepos.Position
	defaultValue interface{}
}
type ArrayType struct {
	ItemsType    yamlmeta.Type
	Position     *filepos.Position
	defaultValue interface{}
	description  string
}
type ArrayItemType struct {
	ValueType    yamlmeta.Type
	Position     *filepos.Position
	defaultValue interface{}
}
type ScalarType struct {
	ValueType    interface{}
	Position     *filepos.Position
	defaultValue interface{}
	description  string
}
type AnyType struct {
	defaultValue interface{}
	Position     *filepos.Position
	description  string
}
type NullType struct {
	ValueType   yamlmeta.Type
	Position    *filepos.Position
	description string
}

// SetDefaultValue sets the default value of the wrapped type to `val`
func (n *NullType) SetDefaultValue(val interface{}) {
	n.GetValueType().SetDefaultValue(val)
}

// SetDefaultValue does nothing
func (a *AnyType) SetDefaultValue(val interface{}) {
	a.defaultValue = val
}

// SetDefaultValue sets the default value of the entire document to `val`
func (t *DocumentType) SetDefaultValue(val interface{}) {
	t.defaultValue = val
}

// SetDefaultValue is ignored as default values should be set on each MapItemType, individually.
func (m *MapType) SetDefaultValue(val interface{}) {
	// TODO: determine if we should set the contents of a MapType by setting the given Map...?
	return
}

// SetDefaultValue sets the default value to `val`
func (t *MapItemType) SetDefaultValue(val interface{}) {
	t.defaultValue = val
}

// SetDefaultValue sets the default value to `val`
func (a *ArrayType) SetDefaultValue(val interface{}) {
	a.defaultValue = val
}

// SetDefaultValue sets the default value to `val`
func (a *ArrayItemType) SetDefaultValue(val interface{}) {
	a.defaultValue = val
}

// SetDefaultValue sets the default value to `val`
func (s *ScalarType) SetDefaultValue(val interface{}) {
	s.defaultValue = val
}

func (n NullType) GetDefaultValue() interface{} {
	return nil
}

func (a AnyType) GetDefaultValue() interface{} {
	if node, ok := a.defaultValue.(yamlmeta.Node); ok {
		return node.DeepCopyAsInterface()
	}
	return a.defaultValue
}

// GetDefaultValue provides the default value
func (s ScalarType) GetDefaultValue() interface{} {
	return s.defaultValue // scalar values are copied (even through an interface{} reference)
}

func (a ArrayItemType) GetDefaultValue() interface{} {
	panic(fmt.Sprintf("Unexpected call to GetDefaultValue() on %+v", a))
}

func (a ArrayType) GetDefaultValue() interface{} {
	return a.defaultValue
}

func (t MapItemType) GetDefaultValue() interface{} {
	return &yamlmeta.MapItem{Key: t.Key, Value: t.defaultValue, Position: t.Position}
}

func (m MapType) GetDefaultValue() interface{} {
	defaultValues := &yamlmeta.Map{Position: m.Position}
	for _, item := range m.Items {
		newItem := item.GetDefaultValue()
		defaultValues.Items = append(defaultValues.Items, newItem.(*yamlmeta.MapItem))
	}
	return defaultValues
}

func (t DocumentType) GetDefaultValue() interface{} {
	return &yamlmeta.Document{Value: t.defaultValue, Position: t.Position}
}

func (n NullType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	childCheck := n.ValueType.AssignTypeTo(typeable)
	chk.Violations = append(chk.Violations, childCheck.Violations...)
	return
}

func (n NullType) GetValueType() yamlmeta.Type {
	return n.ValueType
}

func (n NullType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	if len(node.GetValues()) == 1 && node.GetValues()[0] == nil {
		return
	}

	check := n.GetValueType().CheckType(node)
	chk.Violations = check.Violations

	return
}

func (n NullType) GetDefinitionPosition() *filepos.Position {
	return n.Position
}

func (n NullType) String() string {
	return "null"
}

func (t *DocumentType) GetValueType() yamlmeta.Type {
	return t.ValueType
}
func (m MapType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (t MapItemType) GetValueType() yamlmeta.Type {
	return t.ValueType
}
func (a ArrayType) GetValueType() yamlmeta.Type {
	return a.ItemsType
}
func (a ArrayItemType) GetValueType() yamlmeta.Type {
	return a.ValueType
}

// GetValueType provides the type of the value
func (s ScalarType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (a AnyType) GetValueType() yamlmeta.Type {
	return &a
}

func (t *DocumentType) GetDefinitionPosition() *filepos.Position {
	return t.Position
}
func (m MapType) GetDefinitionPosition() *filepos.Position {
	return m.Position
}
func (t MapItemType) GetDefinitionPosition() *filepos.Position {
	return t.Position
}
func (a ArrayType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}
func (a ArrayItemType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}

// GetDefinitionPosition provides the file position
func (s ScalarType) GetDefinitionPosition() *filepos.Position {
	return s.Position
}
func (a AnyType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}

func (t *DocumentType) String() string {
	return "document"
}
func (m MapType) String() string {
	return "map"
}
func (t MapItemType) String() string {
	return fmt.Sprintf("%s: %s", t.Key, t.ValueType.String())
}
func (a ArrayType) String() string {
	return "array"
}
func (a ArrayItemType) String() string {
	return fmt.Sprintf("- %s", a.ValueType.String())
}
func (s ScalarType) String() string {
	switch s.ValueType.(type) {
	case float64:
		return "float"
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		return fmt.Sprintf("%T", s.ValueType)
	}
}
func (a AnyType) String() string {
	return "any"
}

func (t *DocumentType) CheckType(_ yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	return
}

func (m *MapType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	nodeMap, ok := node.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(node, m))
		return
	}

	for _, item := range nodeMap.Items {
		if !m.AllowsKey(item.Key) {
			chk.Violations = append(chk.Violations,
				NewUnexpectedKeyAssertionError(item, m.Position, m.AllowedKeys()))
		}
	}
	return
}

func (t *MapItemType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	_, ok := node.(*yamlmeta.MapItem)
	if !ok {
		// A Map must've yielded a non-MapItem which is not valid YAML
		panic(fmt.Sprintf("MapItem type check was called on a non-MapItem: %#v", node))
	}

	return
}

func (a *ArrayType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	_, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(node, a))
	}
	return
}

func (a *ArrayItemType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	_, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		// An Array must've yielded a non-ArrayItem which is not valid YAML
		panic(fmt.Sprintf("ArrayItem type check was called on a non-ArrayItem: %#v", node))
	}
	return
}

// CheckType validates the type of the node and the type of the value
func (s *ScalarType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	value := node.GetValues()[0]
	switch value.(type) {
	case string:
		if _, ok := s.ValueType.(string); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeAssertionError(node, s))
		}
	case float64:
		if _, ok := s.ValueType.(float64); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeAssertionError(node, s))
		}
	case int, int64, uint64:
		if _, ok := s.ValueType.(int); !ok {
			if _, ok = s.ValueType.(float64); !ok {
				chk.Violations = append(chk.Violations,
					NewMismatchedTypeAssertionError(node, s))
			}
		}
	case bool:
		if _, ok := s.ValueType.(bool); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeAssertionError(node, s))
		}
	default:
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(node, s))
	}
	return
}

func (a AnyType) CheckType(_ yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	return
}

func (t *DocumentType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	doc, ok := typeable.(*yamlmeta.Document)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(typeable, t))
		return
	}
	doc.SetType(t)
	typeableValue, isNode := doc.Value.(yamlmeta.Typeable)
	if isNode {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is a scalar
	return chk
}

func (m *MapType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapNode, ok := typeable.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(typeable, m))
		return
	}
	var foundKeys []interface{}
	typeable.SetType(m)
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

func (m *MapType) applySchemaDefaults(foundKeys []interface{}, chk yamlmeta.TypeCheck, mapNode *yamlmeta.Map) {
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

func (t *MapItemType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapItem, ok := typeable.(*yamlmeta.MapItem)
	if !ok {
		panic(fmt.Sprintf("Attempt to assign type to a non-map-item (children of Maps can only be MapItems). type=%#v; typeable=%#v", t, typeable))
	}
	typeable.SetType(t)
	typeableValue, isNode := mapItem.Value.(yamlmeta.Typeable)
	if isNode {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return
}

func (a *ArrayType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayNode, ok := typeable.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(typeable, a))
		return
	}
	typeable.SetType(a)
	for _, arrayItem := range arrayNode.Items {
		childCheck := a.ItemsType.AssignTypeTo(arrayItem)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	}
	return
}

func (a *ArrayItemType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayItem, ok := typeable.(*yamlmeta.ArrayItem)
	if !ok {
		panic(fmt.Sprintf("Attempt to assign type to a non-array-item (children of Arrays can only be ArrayItems). type=%#v; typeable=%#v", a, typeable))
	}
	typeable.SetType(a)
	typeableValue, isNode := arrayItem.Value.(yamlmeta.Typeable)
	if isNode {
		childCheck := a.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, is scalar
	return
}

// AssignTypeTo validates that the type is compatible and assigns it to the type
func (s *ScalarType) AssignTypeTo(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{[]error{NewMismatchedTypeAssertionError(typeable, s)}}
}

func (a AnyType) AssignTypeTo(yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	return
}

// GetDescription provides descriptive information
func (t *DocumentType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (m *MapType) GetDescription() string {
	return m.description
}

// GetDescription provides descriptive information
func (t *MapItemType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (a *ArrayType) GetDescription() string {
	return a.description
}

// GetDescription provides descriptive information
func (a *ArrayItemType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (s *ScalarType) GetDescription() string {
	return s.description
}

// GetDescription provides descriptive information
func (a *AnyType) GetDescription() string {
	return a.description
}

// GetDescription provides descriptive information
func (n *NullType) GetDescription() string {
	return n.description
}

// SetDescription sets the description of the type
func (t *DocumentType) SetDescription(desc string) {}

// SetDescription sets the description of the type
func (m *MapType) SetDescription(desc string) {
	m.description = desc
}

// SetDescription sets the description of the type
func (t *MapItemType) SetDescription(desc string) {}

// SetDescription sets the description of the type
func (a *ArrayType) SetDescription(desc string) {
	a.description = desc
}

// SetDescription sets the description of the type
func (a *ArrayItemType) SetDescription(desc string) {}

// SetDescription sets the description of the type
func (s *ScalarType) SetDescription(desc string) {
	s.description = desc
}

// SetDescription sets the description of the type
func (a *AnyType) SetDescription(desc string) {
	a.description = desc
}

// SetDescription sets the description of the type
func (n *NullType) SetDescription(desc string) {
	n.description = desc
}

func (m *MapType) AllowsKey(key interface{}) bool {
	for _, item := range m.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

// AllowedKeys returns the set of keys (in string format) permitted in this map.
func (m *MapType) AllowedKeys() []string {
	var keysAsString []string

	for _, item := range m.Items {
		keysAsString = append(keysAsString, fmt.Sprintf("%s", item.Key))
	}

	return keysAsString
}
