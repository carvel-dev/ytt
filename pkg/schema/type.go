// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationSchemaNullable structmeta.AnnotationName = "schema/nullable"
)

var _ yamlmeta.Type = (*DocumentType)(nil)
var _ yamlmeta.Type = (*MapType)(nil)
var _ yamlmeta.Type = (*MapItemType)(nil)
var _ yamlmeta.Type = (*ArrayType)(nil)
var _ yamlmeta.Type = (*ArrayItemType)(nil)

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
	Annotations  TypeAnnotations
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
	Type     interface{}
	Position *filepos.Position
}

type TypeAnnotations map[structmeta.AnnotationName]interface{}

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
	// TODO: is this correct? returns a func() yamlmeta.Type (?)
	return a.ItemsType
	//panic("Not implemented because it is unreachable")
}
func (a ArrayItemType) GetValueType() yamlmeta.Type {
	return a.ValueType
}
func (m ScalarType) GetValueType() yamlmeta.Type {
	panic("Not implemented because it is unreachable")
}

func (t *DocumentType) CheckType(_ yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	return
}
func (m *MapType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	nodeMap, ok := node.(*yamlmeta.Map)
	if !ok {
		// TODO: Can this be removed? Can we replace "map item" with a call to typeToString?
		//scalar, ok := node.(*yamlmeta.Scalar)
		//if ok {
		//	chk.Violations = append(chk.Violations,
		//		NewTypeError(scalar.ValueTypeAsString(), "map item", node.GetPosition(), m.Position))
		//} else {
		chk.Violations = append(chk.Violations,
			NewTypeError(node.ValueTypeAsString(), "map item", node.GetPosition(), m.Position))
		//}
		return
	}

	for _, item := range nodeMap.Items {
		if !m.AllowsKey(item.Key) {
			chk.Violations = append(chk.Violations,
				NewUnexpectedKeyError(item.Key.(string), item.Position, m.Position))
		}
	}
	return
}
func (t MapItemType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	mapItem, ok := node.(*yamlmeta.MapItem)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewTypeError(node.ValueTypeAsString(),
				//TODO: Is this the right expected type for this error?
				typeToString(t.GetValueType()),
				node.GetPosition(),
				t.Position))
		return
	}
	if mapItem.Value == nil && !t.IsNullable() {
		chk.Violations = append(chk.Violations,
			NewTypeError(node.ValueTypeAsString(),
				typeToString(t.GetValueType()),
				node.GetPosition(),
				t.Position))
	}

	return
}
func (a ArrayType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	_, ok := node.(*yamlmeta.Array)
	if !ok {
		//scalar, ok := node.(*yamlmeta.Scalar)
		//if ok {
		//	chk.Violations = append(chk.Violations,
		//NewTypeError(scalar.ValueTypeAsString(), typeToString(a.GetValueType), node.GetPosition(), a.Position))
		//} else {
		chk.Violations = append(chk.Violations,
			//TODO: expected can be made typeToString(a.GetValueType) ?
			NewTypeError(node.ValueTypeAsString(), "array element", node.GetPosition(), a.Position))
		//}
	}
	return
}
func (a ArrayItemType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	_, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewTypeError(node.ValueTypeAsString(), typeToString(&yamlmeta.ArrayItem{}), node.GetPosition(), a.Position))
		return
	}
	return
}
func (m ScalarType) CheckType(node yamlmeta.TypeWithValues) (chk yamlmeta.TypeCheck) {
	value := node.GetValues()[0]
	switch itemValueType := value.(type) {
	case string:
		if _, ok := m.Type.(string); !ok {
			chk.Violations = append(chk.Violations,
				NewTypeError(node.ValueTypeAsString(), typeToString(m.Type), node.GetPosition(), m.Position))
		}
	case int:
		if _, ok := m.Type.(int); !ok {
			chk.Violations = append(chk.Violations,
				NewTypeError(node.ValueTypeAsString(), typeToString(m.Type), node.GetPosition(), m.Position))
		}
	case bool:
		if _, ok := m.Type.(bool); !ok {
			chk.Violations = append(chk.Violations,
				NewTypeError(node.ValueTypeAsString(), typeToString(m.Type), node.GetPosition(), m.Position))
		}
	default:
		chk.Violations = append(chk.Violations,
			NewTypeError(typeToString(itemValueType), typeToString(m.Type), node.GetPosition(), m.Position))
	}
	return
}

func typeToString(value interface{}) string {
	// TODO: this functions is duplicated
	switch typedValue := value.(type) {
	case *ScalarType:
		return typeToString(typedValue.Type)
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		if t, ok := value.(yamlmeta.TypeWithValues); ok {
			return t.ValueTypeAsString()
		}

		return fmt.Sprintf("%T", value)
	}
}

func (t *DocumentType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	doc, ok := typeable.(*yamlmeta.Document)
	if !ok {
		chk.Violations = append(chk.Violations,
			//TODO: Can we replace the %Ts with typeToString?
			NewTypeError(fmt.Sprintf("%T", typeable),
				fmt.Sprintf("%T", &yamlmeta.Document{}),
				typeable.GetPosition(),
				t.Position))
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
				default:
					chk.Violations = append(chk.Violations,
						NewTypeError(fmt.Sprintf("%T", t.ValueType),
							typeToString(&yamlmeta.Map{}),
							typeableChild.GetPosition(),
							t.Position))
				}
				doc.Value = tChild
			}
			childCheck := t.ValueType.AssignTypeTo(tChild)
			chk.Violations = append(chk.Violations, childCheck.Violations...)
		} else {
			// TODO: this message points to the document start (---), which is kind of correct: the document
			//       by default becomes a map which isn't allowed, but the error is probably confusing
			//       Maybe this error should be at a higher level. Maybe not even this format?
			chk.Violations = append(chk.Violations,
				NewTypeError(typeToString(typeableChild),
					"nil",
					typeableChild.GetPosition(),
					t.Position))
		}
	} else {

	} // else, at a leaf
	return
}
func (m *MapType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	mapNode, ok := typeable.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewTypeError(typeToString(typeable),
				typeToString(&yamlmeta.Map{}),
				typeable.GetPosition(),
				m.Position))
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
		chk.Violations = append(chk.Violations,
			NewTypeError(fmt.Sprintf("%T", typeable),
				fmt.Sprintf("%T", &yamlmeta.MapItem{}),
				typeable.GetPosition(),
				t.Position))
		return
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
		chk.Violations = append(chk.Violations,
			NewTypeError(typeToString(typeable),
				typeToString(&yamlmeta.Array{}),
				typeable.GetPosition(),
				a.Position))
		return
	}
	typeable.SetType(a)
	for _, arrayItem := range arrayNode.Items {
		childCheck := a.ItemsType.AssignTypeTo(arrayItem)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	}
	return
}
func (a ArrayItemType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	arrayItem, ok := typeable.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = append(chk.Violations,
			NewTypeError(typeToString(typeable),
				typeToString(&yamlmeta.ArrayItem{}),
				typeable.GetPosition(),
				a.Position))
		return
	}
	typeable.SetType(a)
	typeableValue, ok := arrayItem.Value.(yamlmeta.Typeable)
	if ok {
		childCheck := a.ValueType.AssignTypeTo(typeableValue)
		chk.Violations = append(chk.Violations, childCheck.Violations...)
	} // else, at a leaf
	return
}
func (m *ScalarType) AssignTypeTo(typeable yamlmeta.Typeable) (chk yamlmeta.TypeCheck) {
	switch m.Type.(type) {
	case int:
		typeable.SetType(m)
	case string:
		typeable.SetType(m)
	default:
		chk.Violations = append(chk.Violations,
			NewTypeError(typeToString(typeable),
				typeToString(&ScalarType{}),
				typeable.GetPosition(),
				m.Position))
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

func (t MapItemType) IsNullable() bool {
	_, found := t.Annotations[AnnotationSchemaNullable]
	return found
}
