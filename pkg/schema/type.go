// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

// Type encapsulates a schema that describes a yamlmeta.Node.
type Type interface {
	AssignTypeTo(node yamlmeta.Node) TypeCheck
	CheckType(node yamlmeta.Node) TypeCheck

	GetValueType() Type
	GetDefaultValue() interface{}
	SetDefaultValue(interface{})
	GetDefinitionPosition() *filepos.Position

	GetDescription() string
	GetTitle() string
	GetExample() *Example
	SetDocumentation(Documentation)
	String() string
}

var _ Type = (*DocumentType)(nil)
var _ Type = (*MapType)(nil)
var _ Type = (*MapItemType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*ArrayItemType)(nil)
var _ Type = (*ScalarType)(nil)
var _ Type = (*AnyType)(nil)
var _ Type = (*NullType)(nil)

type DocumentType struct {
	Source       *yamlmeta.Document
	ValueType    Type // typically one of: MapType, ArrayType, ScalarType
	Position     *filepos.Position
	defaultValue interface{}
}
type MapType struct {
	Items         []*MapItemType
	Position      *filepos.Position
	documentation Documentation
}
type MapItemType struct {
	Key          interface{} // usually a string
	ValueType    Type
	Position     *filepos.Position
	defaultValue interface{}
}
type ArrayType struct {
	ItemsType     Type
	Position      *filepos.Position
	defaultValue  interface{}
	documentation Documentation
}
type ArrayItemType struct {
	ValueType    Type
	Position     *filepos.Position
	defaultValue interface{}
}
type ScalarType struct {
	ValueType     interface{}
	Position      *filepos.Position
	defaultValue  interface{}
	documentation Documentation
}
type AnyType struct {
	defaultValue  interface{}
	Position      *filepos.Position
	documentation Documentation
}
type NullType struct {
	ValueType     Type
	Position      *filepos.Position
	documentation Documentation
}

// The total set of supported scalars.
const (
	FloatType  = float64(0)
	StringType = ""
	IntType    = int64(0)
	BoolType   = false
)

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

// GetValueType provides the type of the value
func (n NullType) GetValueType() Type {
	return n.ValueType
}

// GetDefaultValue provides the default value
func (t DocumentType) GetDefaultValue() interface{} {
	return &yamlmeta.Document{Value: t.defaultValue, Position: t.Position}
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
func (t MapItemType) GetDefaultValue() interface{} {
	return &yamlmeta.MapItem{Key: t.Key, Value: t.defaultValue, Position: t.Position}
}

// GetDefaultValue provides the default value
func (a ArrayType) GetDefaultValue() interface{} {
	return a.defaultValue
}

// GetDefaultValue provides the default value
func (a ArrayItemType) GetDefaultValue() interface{} {
	panic(fmt.Sprintf("Unexpected call to GetDefaultValue() on %+v", a))
}

// GetDefaultValue provides the default value
func (s ScalarType) GetDefaultValue() interface{} {
	return s.defaultValue // scalar values are copied (even through an interface{} reference)
}

// GetDefaultValue provides the default value
func (a AnyType) GetDefaultValue() interface{} {
	if node, ok := a.defaultValue.(yamlmeta.Node); ok {
		return node.DeepCopyAsInterface()
	}
	return a.defaultValue
}

// GetDefaultValue provides the default value
func (n NullType) GetDefaultValue() interface{} {
	return nil
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

// SetDefaultValue does nothing
func (a *AnyType) SetDefaultValue(val interface{}) {
	a.defaultValue = val
}

// SetDefaultValue sets the default value of the wrapped type to `val`
func (n *NullType) SetDefaultValue(val interface{}) {
	n.GetValueType().SetDefaultValue(val)
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (t *DocumentType) GetDefinitionPosition() *filepos.Position {
	return t.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (m MapType) GetDefinitionPosition() *filepos.Position {
	return m.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (t MapItemType) GetDefinitionPosition() *filepos.Position {
	return t.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (a ArrayType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (a ArrayItemType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (s ScalarType) GetDefinitionPosition() *filepos.Position {
	return s.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (a AnyType) GetDefinitionPosition() *filepos.Position {
	return a.Position
}

// GetDefinitionPosition reports the location in source schema that contains this type definition.
func (n NullType) GetDefinitionPosition() *filepos.Position {
	return n.Position
}

// GetDescription provides descriptive information
func (t *DocumentType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (m *MapType) GetDescription() string {
	return m.documentation.description
}

// GetDescription provides descriptive information
func (t *MapItemType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (a *ArrayType) GetDescription() string {
	return a.documentation.description
}

// GetDescription provides descriptive information
func (a *ArrayItemType) GetDescription() string {
	return ""
}

// GetDescription provides descriptive information
func (s *ScalarType) GetDescription() string {
	return s.documentation.description
}

// GetDescription provides descriptive information
func (a *AnyType) GetDescription() string {
	return a.documentation.description
}

// GetDescription provides descriptive information
func (n *NullType) GetDescription() string {
	return n.documentation.description
}

// GetTitle provides title information
func (t *DocumentType) GetTitle() string {
	return ""
}

// GetTitle provides title information
func (m *MapType) GetTitle() string {
	return m.documentation.title
}

// GetTitle provides title information
func (t *MapItemType) GetTitle() string {
	return ""
}

// GetTitle provides title information
func (a *ArrayType) GetTitle() string {
	return a.documentation.title
}

// GetTitle provides title information
func (a *ArrayItemType) GetTitle() string {
	return ""
}

// GetTitle provides title information
func (s *ScalarType) GetTitle() string {
	return s.documentation.title
}

// GetTitle provides title information
func (a *AnyType) GetTitle() string {
	return a.documentation.title
}

// GetTitle provides title information
func (n *NullType) GetTitle() string {
	return n.documentation.title
}

// GetExample provides descriptive example information
func (t *DocumentType) GetExample() *Example {
	return nil
}

// GetExample provides descriptive example information
func (m *MapType) GetExample() *Example {
	return m.documentation.Example
}

// GetExample provides descriptive example information
func (t *MapItemType) GetExample() *Example {
	return nil
}

// GetExample provides descriptive example information
func (a *ArrayType) GetExample() *Example {
	return a.documentation.Example
}

// GetExample provides descriptive example information
func (a *ArrayItemType) GetExample() *Example {
	return nil
}

// GetExample provides descriptive example information
func (s *ScalarType) GetExample() *Example {
	return s.documentation.Example
}

// GetExample provides descriptive example information
func (a *AnyType) GetExample() *Example {
	return a.documentation.Example
}

// GetExample provides descriptive example information
func (n *NullType) GetExample() *Example {
	return n.documentation.Example
}

// SetDocumentation sets the description and example of the type
func (t *DocumentType) SetDocumentation(data Documentation) {}

// SetDocumentation sets the description and example of the type
func (m *MapType) SetDocumentation(data Documentation) {
	m.documentation = data
}

// SetDocumentation sets the description and example of the type
func (t *MapItemType) SetDocumentation(data Documentation) {}

// SetDocumentation sets the description and example of the type
func (a *ArrayType) SetDocumentation(data Documentation) {
	a.documentation = data
}

// SetDocumentation sets the description and example of the type
func (a *ArrayItemType) SetDocumentation(data Documentation) {}

// SetDocumentation sets the description and example of the type
func (s *ScalarType) SetDocumentation(data Documentation) {
	s.documentation = data
}

// SetDocumentation sets the description and example of the type
func (a *AnyType) SetDocumentation(data Documentation) {
	a.documentation = data
}

// SetDocumentation sets the description and example of the type
func (n *NullType) SetDocumentation(data Documentation) {
	n.documentation = data
}

// String produces a user-friendly name of the expected type.
func (t *DocumentType) String() string {
	return yamlmeta.TypeName(&yamlmeta.Document{})
}

// String produces a user-friendly name of the expected type.
func (m MapType) String() string {
	return yamlmeta.TypeName(&yamlmeta.Map{})
}

// String produces a user-friendly name of the expected type.
func (t MapItemType) String() string {
	return fmt.Sprintf("%s: %s", t.Key, t.ValueType.String())
}

// String produces a user-friendly name of the expected type.
func (a ArrayType) String() string {
	return yamlmeta.TypeName(&yamlmeta.Array{})
}

// String produces a user-friendly name of the expected type.
func (a ArrayItemType) String() string {
	return fmt.Sprintf("- %s", a.ValueType.String())
}

// String produces a user-friendly name of the expected type.
func (s ScalarType) String() string {
	return yamlmeta.TypeName(s.ValueType)
}

// String produces a user-friendly name of the expected type.
func (a AnyType) String() string {
	return "any"
}

// String produces a user-friendly name of the expected type.
func (n NullType) String() string {
	return "null"
}
