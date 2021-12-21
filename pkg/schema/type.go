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

// AssignTypeTo assigns this NullType's wrapped Type to `node`.
func (n NullType) AssignTypeTo(node yamlmeta.Node) (chk TypeCheck) {
	childCheck := n.ValueType.AssignTypeTo(node)
	chk.Violations = append(chk.Violations, childCheck.Violations...)
	return
}

// GetValueType provides the type of the value
func (n NullType) GetValueType() Type {
	return n.ValueType
}

// CheckType checks the type of `node` against this NullType
// If `node`'s value is null, this check passes
// If `node`'s value is not null, then it is checked against this NullType's wrapped Type.
func (n NullType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
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

// CheckType checks the type of `node` against this Type.
func (t *DocumentType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
	return
}

// CheckType checks the type of `node` against this MapType.
// If `node` is not a yamlmeta.Map, `chk` contains a violation describing this mismatch
// If a contained yamlmeta.MapItem is not allowed by this MapType, `chk` contains a corresponding violation
func (m *MapType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
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

// CheckType checks the type of `node` against this MapItemType
// If `node` is not a yamlmeta.MapItem, `chk` contains a violation describing the mismatch
func (t *MapItemType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
	_, ok := node.(*yamlmeta.MapItem)
	if !ok {
		// A Map must've yielded a non-MapItem which is not valid YAML
		panic(fmt.Sprintf("Map item type check was called on a non-map item: %#v", node))
	}

	return
}

// CheckType checks the type of `node` against this ArrayType
// If `node` is not a yamlmeta.Array, `chk` contains a violation describing the mismatch
func (a *ArrayType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
	_, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeAssertionError(node, a))
	}
	return
}

// CheckType checks the type of `node` against this ArrayItemType
// If `node` is not a yamlmeta.ArrayItem, `chk` contains a violation describing the mismatch
func (a *ArrayItemType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
	_, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		// An Array must've yielded a non-ArrayItem which is not valid YAML
		panic(fmt.Sprintf("Array item type check was called on a non-array item: %#v", node))
	}
	return
}

// CheckType checks the type of `node`'s `value`, which is expected to be a scalar type.
// If the value is not a recognized scalar type, `chk` contains a corresponding violation
// If the value is not of the type specified in this ScalarType, `chk` contains a violation describing the mismatch
func (s *ScalarType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
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

// CheckType is a no-op because AnyType allows any value.
// `chk` will always be an empty TypeCheck.
func (a AnyType) CheckType(node yamlmeta.Node) (chk TypeCheck) {
	return
}

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

func contains(haystack []interface{}, needle interface{}) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
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

// GetType retrieves schema metadata from `n`, typically set previously via SetType().
func GetType(n yamlmeta.Node) Type {
	t := n.GetMeta("schema/type")
	if t == nil {
		return nil
	}
	return t.(Type)
}

// SetType attaches schema metadata to `n`, typically later retrieved via GetType().
func SetType(n yamlmeta.Node, t Type) {
	n.SetMeta("schema/type", t)
}

func nodeValueTypeAsString(n yamlmeta.Node) string {
	switch typed := n.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Map, *yamlmeta.Array:
		return yamlmeta.TypeName(typed)
	default:
		return yamlmeta.TypeName(typed.GetValues()[0])
	}
}
