// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/validations"
	"carvel.dev/ytt/pkg/yamlmeta"
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
	SetDescription(string)
	GetTitle() string
	SetTitle(string)
	GetExamples() []Example
	SetExamples([]Example)
	IsDeprecated() (bool, string)
	SetDeprecated(bool, string)
	GetValidation() *validations.NodeValidation
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
	validations  *validations.NodeValidation
}

type MapType struct {
	Items         []*MapItemType
	Position      *filepos.Position
	documentation documentation
}

type MapItemType struct {
	Key          interface{} // usually a string
	ValueType    Type
	Position     *filepos.Position
	defaultValue interface{}
	validations  *validations.NodeValidation
}

type ArrayType struct {
	ItemsType     Type
	Position      *filepos.Position
	defaultValue  interface{}
	documentation documentation
}

type ArrayItemType struct {
	ValueType    Type
	Position     *filepos.Position
	defaultValue interface{}
	validations  *validations.NodeValidation
}

type ScalarType struct {
	ValueType     interface{}
	Position      *filepos.Position
	defaultValue  interface{}
	documentation documentation
}

type AnyType struct {
	defaultValue  interface{}
	Position      *filepos.Position
	documentation documentation
}

type NullType struct {
	ValueType     Type
	Position      *filepos.Position
	documentation documentation
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
func (m *MapType) SetDefaultValue(_ interface{}) {
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

// SetDescription sets the description of the type
func (t *DocumentType) SetDescription(_ string) {}

// SetDescription sets the description of the type
func (m *MapType) SetDescription(desc string) {
	m.documentation.description = desc
}

// SetDescription sets the description of the type
func (t *MapItemType) SetDescription(_ string) {}

// SetDescription sets the description of the type
func (a *ArrayType) SetDescription(desc string) {
	a.documentation.description = desc
}

// SetDescription sets the description of the type
func (a *ArrayItemType) SetDescription(_ string) {}

// SetDescription sets the description of the type
func (s *ScalarType) SetDescription(desc string) {
	s.documentation.description = desc
}

// SetDescription sets the description of the type
func (a *AnyType) SetDescription(desc string) {
	a.documentation.description = desc
}

// SetDescription sets the description of the type
func (n *NullType) SetDescription(desc string) {
	n.documentation.description = desc
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

// SetTitle sets the title of the type
func (t *DocumentType) SetTitle(_ string) {}

// SetTitle sets the title of the type
func (m *MapType) SetTitle(title string) {
	m.documentation.title = title
}

// SetTitle sets the title of the type
func (t *MapItemType) SetTitle(_ string) {}

// SetTitle sets the title of the type
func (a *ArrayType) SetTitle(title string) {
	a.documentation.title = title
}

// SetTitle sets the title of the type
func (a *ArrayItemType) SetTitle(_ string) {}

// SetTitle sets the title of the type
func (s *ScalarType) SetTitle(title string) {
	s.documentation.title = title
}

// SetTitle sets the title of the type
func (a *AnyType) SetTitle(title string) {
	a.documentation.title = title
}

// SetTitle sets the title of the type
func (n *NullType) SetTitle(title string) {
	n.documentation.title = title
}

// GetExamples provides descriptive example information
func (t *DocumentType) GetExamples() []Example {
	return nil
}

// GetExamples provides descriptive example information
func (m *MapType) GetExamples() []Example {
	return m.documentation.examples
}

// GetExamples provides descriptive example information
func (t *MapItemType) GetExamples() []Example {
	return nil
}

// GetExamples provides descriptive example information
func (a *ArrayType) GetExamples() []Example {
	return a.documentation.examples
}

// GetExamples provides descriptive example information
func (a *ArrayItemType) GetExamples() []Example {
	return nil
}

// GetExamples provides descriptive example information
func (s *ScalarType) GetExamples() []Example {
	return s.documentation.examples
}

// GetExamples provides descriptive example information
func (a *AnyType) GetExamples() []Example {
	return a.documentation.examples
}

// GetExamples provides descriptive example information
func (n *NullType) GetExamples() []Example {
	return n.documentation.examples
}

// SetExamples sets the description and example of the type
func (t *DocumentType) SetExamples(_ []Example) {}

// SetExamples sets the description and example of the type
func (m *MapType) SetExamples(exs []Example) {
	m.documentation.examples = exs
}

// SetExamples sets the description and example of the type
func (t *MapItemType) SetExamples(_ []Example) {}

// SetExamples sets the description and example of the type
func (a *ArrayType) SetExamples(exs []Example) {
	a.documentation.examples = exs
}

// SetExamples sets the description and example of the type
func (a *ArrayItemType) SetExamples(_ []Example) {}

// SetExamples sets the description and example of the type
func (s *ScalarType) SetExamples(exs []Example) {
	s.documentation.examples = exs
}

// SetExamples sets the description and example of the type
func (a *AnyType) SetExamples(exs []Example) {
	a.documentation.examples = exs
}

// SetExamples sets the description and example of the type
func (n *NullType) SetExamples(exs []Example) {
	n.documentation.examples = exs
}

// IsDeprecated provides deprecated field information
func (t *DocumentType) IsDeprecated() (bool, string) {
	return false, ""
}

// IsDeprecated provides deprecated field information
func (m *MapType) IsDeprecated() (bool, string) {
	return m.documentation.deprecated, m.documentation.deprecationNotice

}

// IsDeprecated provides deprecated field information
func (t *MapItemType) IsDeprecated() (bool, string) {
	return false, ""
}

// IsDeprecated provides deprecated field information
func (a *ArrayType) IsDeprecated() (bool, string) {
	return a.documentation.deprecated, a.documentation.deprecationNotice
}

// IsDeprecated provides deprecated field information
func (a *ArrayItemType) IsDeprecated() (bool, string) {
	return false, ""
}

// IsDeprecated provides deprecated field information
func (s *ScalarType) IsDeprecated() (bool, string) {
	return s.documentation.deprecated, s.documentation.deprecationNotice
}

// IsDeprecated provides deprecated field information
func (a *AnyType) IsDeprecated() (bool, string) {
	return a.documentation.deprecated, a.documentation.deprecationNotice
}

// IsDeprecated provides deprecated field information
func (n *NullType) IsDeprecated() (bool, string) {
	return n.documentation.deprecated, n.documentation.deprecationNotice
}

// SetDeprecated sets the deprecated field value
func (t *DocumentType) SetDeprecated(_ bool, _ string) {}

// SetDeprecated sets the deprecated field value
func (m *MapType) SetDeprecated(deprecated bool, notice string) {
	m.documentation.deprecationNotice = notice
	m.documentation.deprecated = deprecated
}

// SetDeprecated sets the deprecated field value
func (t *MapItemType) SetDeprecated(_ bool, _ string) {}

// SetDeprecated sets the deprecated field value
func (a *ArrayType) SetDeprecated(deprecated bool, notice string) {
	a.documentation.deprecationNotice = notice
	a.documentation.deprecated = deprecated
}

// SetDeprecated sets the deprecated field value
func (a *ArrayItemType) SetDeprecated(_ bool, _ string) {}

// SetDeprecated sets the deprecated field value
func (s *ScalarType) SetDeprecated(deprecated bool, notice string) {
	s.documentation.deprecationNotice = notice
	s.documentation.deprecated = deprecated
}

// SetDeprecated sets the deprecated field value
func (a *AnyType) SetDeprecated(deprecated bool, notice string) {
	a.documentation.deprecationNotice = notice
	a.documentation.deprecated = deprecated
}

// SetDeprecated sets the deprecated field value
func (n *NullType) SetDeprecated(deprecated bool, notice string) {
	n.documentation.deprecationNotice = notice
	n.documentation.deprecated = deprecated
}

// GetValidation provides the validation from @schema/validation for a node
func (t *DocumentType) GetValidation() *validations.NodeValidation {
	return t.validations
}

// GetValidation provides the validation from @schema/validation for a node
func (m *MapType) GetValidation() *validations.NodeValidation {
	return nil
}

// GetValidation provides the validation from @schema/validation for a node
func (t *MapItemType) GetValidation() *validations.NodeValidation {
	return t.validations
}

// GetValidation provides the validation from @schema/validation for a node
func (a *ArrayType) GetValidation() *validations.NodeValidation {
	return nil
}

// GetValidation provides the validation from @schema/validation for a node
func (a *ArrayItemType) GetValidation() *validations.NodeValidation {
	return a.validations
}

// GetValidation provides the validation from @schema/validation for a node
func (s *ScalarType) GetValidation() *validations.NodeValidation {
	return nil
}

// GetValidation provides the validation from @schema/validation for a node
func (a AnyType) GetValidation() *validations.NodeValidation {
	return nil
}

// GetValidation provides the validation from @schema/validation for a node
func (n NullType) GetValidation() *validations.NodeValidation {
	return nil
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
