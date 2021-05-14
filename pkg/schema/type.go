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

type DocumentType struct {
	Source    *yamlmeta.Document
	ValueType yamlmeta.Type // typically one of: MapType, ArrayType, ScalarType
	Position  *filepos.Position
}
type MapType struct {
	Items    []*MapItemType
	Position *filepos.Position
}
type MapItemType struct {
	Key          interface{} // usually a string
	ValueType    yamlmeta.Type
	DefaultValue interface{}
	Position     *filepos.Position
	//Annotations  TypeAnnotations
}
type ArrayType struct {
	ItemsType yamlmeta.Type
	Position  *filepos.Position
}
type ArrayItemType struct {
	ValueType yamlmeta.Type
	Position  *filepos.Position
}
type ScalarType struct {
	Value    interface{}
	Position *filepos.Position
}
type AnyType struct {
	Position *filepos.Position
}

type NullType struct {
	ValueType yamlmeta.Type
	Position  *filepos.Position
}

func (n NullType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	switch typedItem := typeable.(type) {
	// when nullable annotates an arrayItem or mapItem, the value is of that item can be null.
	// when the value is not null, we must default to

	case *yamlmeta.Map, *yamlmeta.Array:
		childCheck := n.ValueType.AssignTypeTo(typeable)
		chk.Violations = append(chk.Violations, childCheck.Violations...)

	case *yamlmeta.MapItem:
		typeable.SetType(n)
		typeableValue, ok := typedItem.Value.(yamlmeta.Typeable)
		if ok {
			childCheck := n.ValueType.AssignTypeTo(typeableValue)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		}
	//TODO: Is this ArrayItem case needed? dead code, we think
	case *yamlmeta.ArrayItem:
		typeable.SetType(n)
		typeableValue, ok := typedItem.Value.(yamlmeta.Typeable)
		if ok {
			childCheck := n.ValueType.AssignTypeTo(typeableValue)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		}
	}
	return
}

func (n NullType) GetValueType() yamlmeta.Type {
	return n.ValueType
}

func (n NullType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	switch typedItem := node.(type) {
	// Arrays and Maps cannot have 'nil' values, so if node is one of those types,
	// then those will be checked with the proper value type in checkCollectionItem()
	case *yamlmeta.MapItem:
		if typedItem.Value == nil {
			return
		}
		check := n.GetValueType().CheckType(node)
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	// is this needed? maybe panic?
	// a: [#@/nu"a", "b", "b]
	case *yamlmeta.ArrayItem:
		panic("arrayItems cannot be annotated as nullable")
	}
	return
}

func (n NullType) PositionOfDefinition() *filepos.Position {
	return n.Position
}

func (n NullType) String() string {
	return "nullable"
}

//type TypeAnnotations map[structmeta.AnnotationName]interface{}

func (t *DocumentType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (m MapType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (t MapItemType) GetValueType() yamlmeta.Type {
	return t.ValueType
}
func (a ArrayType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (a ArrayItemType) GetValueType() yamlmeta.Type {
	return a.ValueType
}
func (m ScalarType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}
func (a AnyType) GetValueType() yamlmeta.Type {
	return a
}

func (t *DocumentType) PositionOfDefinition() *filepos.Position {
	return t.Position
}
func (m MapType) PositionOfDefinition() *filepos.Position {
	return m.Position
}
func (t MapItemType) PositionOfDefinition() *filepos.Position {
	return t.Position
}
func (a ArrayType) PositionOfDefinition() *filepos.Position {
	return a.Position
}
func (a ArrayItemType) PositionOfDefinition() *filepos.Position {
	return a.Position
}
func (m ScalarType) PositionOfDefinition() *filepos.Position {
	return m.Position
}
func (a AnyType) PositionOfDefinition() *filepos.Position {
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
func (m ScalarType) String() string {
	switch m.Value.(type) {
	case float64:
		return "float"
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		return fmt.Sprintf("%T", m.Value)
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
			NewMismatchedTypeError(node, m))
		return
	}

	for _, item := range nodeMap.Items {
		if !m.AllowsKey(item.Key) {
			chk.Violations = append(chk.Violations,
				NewUnexpectedKeyError(item, m.Position))
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
			NewMismatchedTypeError(node, a))
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

func (m *ScalarType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	value := node.GetValues()[0]
	switch value.(type) {
	case string:
		if _, ok := m.Value.(string); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeError(node, m))
		}
	case float64:
		if _, ok := m.Value.(float64); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeError(node, m))
		}
	case int:
		if _, ok := m.Value.(int); !ok {
			if _, ok = m.Value.(float64); !ok {
				chk.Violations = append(chk.Violations,
					NewMismatchedTypeError(node, m))
			}
		}
	case bool:
		if _, ok := m.Value.(bool); !ok {
			chk.Violations = append(chk.Violations,
				NewMismatchedTypeError(node, m))
		}
	default:
		chk.Violations = append(chk.Violations,
			NewMismatchedTypeError(node, m))
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
			NewMismatchedTypeError(typeable, t))
		return
	}

	typeable.SetType(t)
	typeableChild, ok := doc.Value.(yamlmeta.Typeable)
	if ok || doc.Value == nil {
		if t.ValueType != nil {
			tChild := typeableChild
			if doc.Value == nil {
				switch t.ValueType.(type) {
				case *MapType:
					tChild = &yamlmeta.Map{}
				case *ArrayType:
					tChild = &yamlmeta.Array{}
				default:
					panic("implement me!")
				}
				doc.Value = tChild
			}
			childCheck := t.ValueType.AssignTypeTo(tChild)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		} else {
			chk.Violations = append(chk.Violations,
				fmt.Errorf("data values were found in data values file(s), but schema (%s) has no values defined\n"+
					"(hint: define matching keys from data values files(s) in the schema, or do not enable the schema feature)", t.Position.AsCompactString()))
		}
	} else {

	} // else, at a leaf
	return
}

func (m *MapType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapNode, ok := typeable.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeError(typeable, m))
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

		val := &yamlmeta.MapItem{
			Key:      item.Key,
			Value:    item.DefaultValue,
			Position: item.Position,
		}
		childCheck := item.AssignTypeTo(val)
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
	typeableValue, ok := mapItem.Value.(yamlmeta.Typeable)
	if ok {
		childCheck := t.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}

func (a *ArrayType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayNode, ok := typeable.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeError(typeable, a))
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
	typeableValue, ok := arrayItem.Value.(yamlmeta.Typeable)
	if ok {
		childCheck := a.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}

func (m *ScalarType) AssignTypeTo(typeable yamlmeta.Typeable) yamlmeta.TypeCheck {
	return yamlmeta.TypeCheck{[]error{NewMismatchedTypeError(typeable, m)}}
}

func (a AnyType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {

	switch typedItem := typeable.(type) {
	case *yamlmeta.Array:
		typeable.SetType(&ArrayType{ItemsType: &ArrayItemType{ValueType: a, Position: a.Position}})
		for _, arrayItem := range typedItem.Items {
			a.AssignTypeTo(arrayItem)
		}
	case *yamlmeta.Map:
		var mItemTypeS []*MapItemType
		for _, mapItem := range typedItem.Items {
			a.AssignTypeTo(mapItem)
			mItemType := &MapItemType{Key: mapItem.Key, ValueType: a, Position: a.Position}
			mItemTypeS = append(mItemTypeS, mItemType)
		}
		typeable.SetType(&MapType{Items: mItemTypeS})
	case *yamlmeta.ArrayItem:
		typeable.SetType(&ArrayItemType{ValueType: a, Position: a.Position})
		typeableValue, ok := typedItem.Value.(yamlmeta.Typeable)
		if ok {
			a.AssignTypeTo(typeableValue)
		}
	case *yamlmeta.MapItem:
		typeable.SetType(&MapItemType{Key: typedItem.Key, ValueType: a, Position: a.Position})
		typeableValue, ok := typedItem.Value.(yamlmeta.Typeable)
		if ok {
			a.AssignTypeTo(typeableValue)
		}
	}

	return
}

func (m *MapType) AllowsKey(key interface{}) bool {
	for _, item := range m.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

//func combineTypes(types []yamlmeta.Type, defaultValueType yamlmeta.Type) (yamlmeta.Type, error) {
//	// is there 1 type in types?
//	//   return type
//	// is there two or more in the array?
//	//   call compare types on first two types,
//	//   append result with rest of array
//	//   return combinetypes() on result
//	numOfTypes := len(types)
//	if numOfTypes == 0 {
//		panic("whoops")
//	}
//	if numOfTypes == 1 {
//		return types[0], nil
//	}
//	if numOfTypes > 1 {
//		dominantType := compareTypes(types[0], types[1])
//		rest := append(types[2:], dominantType)
//		return combineTypes(rest, defaultValueType)
//	}
//	return nil, fmt.Errorf("reached the end of combine types")
//}
//
//func compareTypes(type1, type2 yamlmeta.Type) yamlmeta.Type {
//
//	any := false
//	nullable := false
//	switch type1.(type) {
//	case AnyType:
//		any = true
//	case NullType:
//		nullable = true
//	default:
//		panic("another whoops")
//	}
//	switch type2.(type) {
//	case AnyType:
//		any = true
//	case NullType:
//		nullable = true
//	default:
//		panic("another whoops")
//	}
//	if any && nullable {
//		panic("expected to find one of @schema/nullable, or @schema/type, but found both")
//	}
//	if any {
//		return AnyType{type1.PositionOfDefinition()}
//	}
//	if nullable {
//		return NullType{type1.PositionOfDefinition()}
//	}
//	panic("no types?")
//}
