// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

// Type encapsulates a schema describing a yamlmeta.Node.
type Type interface {
	AssignTypeTo(node yamlmeta.Node) TypeCheck
	GetValueType() Type
	GetDefaultValue() interface{}
	SetDefaultValue(interface{})
	CheckType(node yamlmeta.Node) TypeCheck
	GetDefinitionPosition() *filepos.Position
	String() string
	SetDescription(string)
	GetDescription() string
}

var _ Type = (*DocumentType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*MapItemType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*ArrayItemType)(nil)
var _ Type = (*AnyType)(nil)
var _ Type = (*NullType)(nil)
var _ Type = (*ScalarType)(nil)

type DocumentType struct {
	Source       *yamlmeta.Document
	ValueType    Type // typically one of: MapType, ArrayType, ScalarType
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
	ValueType    Type
	Position     *filepos.Position
	defaultValue interface{}
}
type ArrayType struct {
	ItemsType    Type
	Position     *filepos.Position
	defaultValue interface{}
	description  string
}
type ArrayItemType struct {
	ValueType    Type
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
	ValueType   Type
	Position    *filepos.Position
	description string
}

// The total set of supported scalars.
const (
	FloatType  = float64(0)
	StringType = ""
	IntType    = int64(0)
	BoolType   = false
)

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

// GetDefaultValue provides the default value
func (n NullType) GetDefaultValue() interface{} {
	return nil
}

// GetDefaultValue provides the default value
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

// GetDefaultValue provides the default value
func (a ArrayType) GetDefaultValue() interface{} {
	return a.defaultValue
}

// GetDefaultValue provides the default value
func (t MapItemType) GetDefaultValue() interface{} {
	return &yamlmeta.MapItem{Key: t.Key, Value: t.defaultValue, Position: t.Position}
}

// GetDefaultValue provides the default value
func (m MapType) GetDefaultValue() interface{} {
	defaultValues := &yamlmeta.Map{Position: m.Position}
	for _, item := range m.Items {
		newItem := item.GetDefaultValue()
		defaultValues.Items = append(defaultValues.Items, newItem.(*yamlmeta.MapItem))
	}
	return defaultValues
}

// GetDefaultValue provides the default value
func (t DocumentType) GetDefaultValue() interface{} {
	return &yamlmeta.Document{Value: t.defaultValue, Position: t.Position}
}

// GetValueType provides the type of the value
func (n NullType) GetValueType() Type {
	return n.ValueType
}

func (n NullType) GetDefinitionPosition() *filepos.Position {
	return n.Position
}

func (n NullType) String() string {
	return "null"
}

// GetValueType provides the type of the value
func (t *DocumentType) GetValueType() Type {
	return t.ValueType
}

// GetValueType provides the type of the value
func (m MapType) GetValueType() Type {
	panic("Not implemented because it is unreachable")
}

// GetValueType provides the type of the value
func (t MapItemType) GetValueType() Type {
	return t.ValueType
}

// GetValueType provides the type of the value
func (a ArrayType) GetValueType() Type {
	return a.ItemsType
}

// GetValueType provides the type of the value
func (a ArrayItemType) GetValueType() Type {
	return a.ValueType
}

// GetValueType provides the type of the value
func (s ScalarType) GetValueType() Type {
	panic("Not implemented because it is unreachable")
}

// GetValueType provides the type of the value
func (a AnyType) GetValueType() Type {
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
	return yamlmeta.TypeName(&yamlmeta.Document{})
}
func (m MapType) String() string {
	return yamlmeta.TypeName(&yamlmeta.Map{})
}
func (t MapItemType) String() string {
	return fmt.Sprintf("%s: %s", t.Key, t.ValueType.String())
}
func (a ArrayType) String() string {
	return yamlmeta.TypeName(&yamlmeta.Array{})
}
func (a ArrayItemType) String() string {
	return fmt.Sprintf("- %s", a.ValueType.String())
}
func (s ScalarType) String() string {
	return yamlmeta.TypeName(s.ValueType)
}
func (a AnyType) String() string {
	return "any"
}

func contains(haystack []interface{}, needle interface{}) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
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

func nodeValueTypeAsString(n yamlmeta.Node) string {
	switch typed := n.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Map, *yamlmeta.Array:
		return yamlmeta.TypeName(typed)
	default:
		return yamlmeta.TypeName(typed.GetValues()[0])
	}
}
