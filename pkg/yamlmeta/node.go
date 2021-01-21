// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"encoding/json"
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func (ds *DocumentSet) GetPosition() *filepos.Position { return ds.Position }
func (d *Document) GetPosition() *filepos.Position     { return d.Position }
func (m *Map) GetPosition() *filepos.Position          { return m.Position }
func (mi *MapItem) GetPosition() *filepos.Position     { return mi.Position }
func (a *Array) GetPosition() *filepos.Position        { return a.Position }
func (ai *ArrayItem) GetPosition() *filepos.Position   { return ai.Position }
func (s *Scalar) GetPosition() *filepos.Position       { return s.Position }

func (ds *DocumentSet) ValueTypeAsString() string { return "documentSet" }
func (d *Document) ValueTypeAsString() string     { return typeToString(d.Value) }
func (m *Map) ValueTypeAsString() string          { return "map" }
func (mi *MapItem) ValueTypeAsString() string     { return typeToString(mi.Value) }
func (a *Array) ValueTypeAsString() string        { return "array" }
func (ai *ArrayItem) ValueTypeAsString() string   { return typeToString(ai.Value) }
func (s *Scalar) ValueTypeAsString() string       { return typeToString(s.Value) }

func typeToString(value interface{}) string {
	// TODO: this functions is duplicated
	switch value.(type) {
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func (ds *DocumentSet) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a documentset")
}

func (d *Document) SetValue(val interface{}) error {
	d.Value = val
	return nil
}

func (m *Map) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a map")
}

func (mi *MapItem) SetValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot set map-or-array-item value (%T) into mapitem", val)
	}
	mi.Value = val
	return nil
}

func (a *Array) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on an array")
}

func (ai *ArrayItem) SetValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot set map-or-array-item value (%T) into arrayitem", val)
	}
	ai.Value = val
	return nil
}

func (ds *DocumentSet) ResetValue() { ds.Items = nil }
func (d *Document) ResetValue()     { d.Value = nil }
func (m *Map) ResetValue()          { m.Items = nil }
func (mi *MapItem) ResetValue()     { mi.Value = nil }
func (a *Array) ResetValue()        { a.Items = nil }
func (ai *ArrayItem) ResetValue()   { ai.Value = nil }

func (ds *DocumentSet) AddValue(val interface{}) error {
	if item, ok := val.(*Document); ok {
		ds.Items = append(ds.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-document value (%T) into documentset", val)
}

func (d *Document) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into document", val)
	}
	d.Value = val
	return nil
}

func (m *Map) AddValue(val interface{}) error {
	if item, ok := val.(*MapItem); ok {
		m.Items = append(m.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-map-item value (%T) into map", val)
}

func (mi *MapItem) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into mapitem", val)
	}
	mi.Value = val
	return nil
}

func (a *Array) AddValue(val interface{}) error {
	if item, ok := val.(*ArrayItem); ok {
		a.Items = append(a.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-array-item value (%T) into array", val)
}

func (ai *ArrayItem) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into arrayitem", val)
	}
	ai.Value = val
	return nil
}

func isMapOrArrayItem(val interface{}) bool {
	switch val.(type) {
	case *MapItem, *ArrayItem:
		return true
	default:
		return false
	}
}

func (ds *DocumentSet) GetValues() []interface{} {
	var result []interface{}
	for _, item := range ds.Items {
		result = append(result, item)
	}
	return result
}

func (d *Document) GetValues() []interface{} { return []interface{}{d.Value} }

func (m *Map) GetValues() []interface{} {
	var result []interface{}
	for _, item := range m.Items {
		result = append(result, item)
	}
	return result
}

func (mi *MapItem) GetValues() []interface{} { return []interface{}{mi.Value} }

func (a *Array) GetValues() []interface{} {
	var result []interface{}
	for _, item := range a.Items {
		result = append(result, item)
	}
	return result
}

func (ai *ArrayItem) GetValues() []interface{} { return []interface{}{ai.Value} }
func (s *Scalar) GetValues() []interface{}     { return []interface{}{s.Value} }

func (ds *DocumentSet) GetMetas() []*Meta { return ds.Metas }
func (d *Document) GetMetas() []*Meta     { return d.Metas }
func (m *Map) GetMetas() []*Meta          { return m.Metas }
func (mi *MapItem) GetMetas() []*Meta     { return mi.Metas }
func (a *Array) GetMetas() []*Meta        { return a.Metas }
func (ai *ArrayItem) GetMetas() []*Meta   { return ai.Metas }

func (ds *DocumentSet) addMeta(meta *Meta) { ds.Metas = append(ds.Metas, meta) }
func (d *Document) addMeta(meta *Meta)     { d.Metas = append(d.Metas, meta) }
func (m *Map) addMeta(meta *Meta) {
	panic(fmt.Sprintf("Attempted to attach metadata (%s) to Map (%v); maps cannot carry metadata", meta.Data, m))
}
func (mi *MapItem) addMeta(meta *Meta) { mi.Metas = append(mi.Metas, meta) }
func (a *Array) addMeta(meta *Meta) {
	panic(fmt.Sprintf("Attempted to attach metadata (%s) to Array (%v); arrays cannot carry metadata", meta.Data, a))
}
func (ai *ArrayItem) addMeta(meta *Meta) { ai.Metas = append(ai.Metas, meta) }

func (ds *DocumentSet) GetAnnotations() interface{} { return ds.annotations }
func (d *Document) GetAnnotations() interface{}     { return d.annotations }
func (m *Map) GetAnnotations() interface{}          { return m.annotations }
func (mi *MapItem) GetAnnotations() interface{}     { return mi.annotations }
func (a *Array) GetAnnotations() interface{}        { return a.annotations }
func (ai *ArrayItem) GetAnnotations() interface{}   { return ai.annotations }

func (ds *DocumentSet) SetAnnotations(anns interface{}) { ds.annotations = anns }
func (d *Document) SetAnnotations(anns interface{})     { d.annotations = anns }
func (m *Map) SetAnnotations(anns interface{})          { m.annotations = anns }
func (mi *MapItem) SetAnnotations(anns interface{})     { mi.annotations = anns }
func (a *Array) SetAnnotations(anns interface{})        { a.annotations = anns }
func (ai *ArrayItem) SetAnnotations(anns interface{})   { ai.annotations = anns }

type TypeCheck struct {
	Violations []error
}

func (tc *TypeCheck) LoadContext(doc *Document) {
	if !tc.HasViolations() {
		return
	}

	for _, violation := range tc.Violations {
		if v, ok := violation.(violationWithContext); ok {
			v.SetContext(doc)
		}
	}
}

func (tc TypeCheck) Error() string {
	if !tc.HasViolations() {
		return ""
	}

	msg := ""
	for _, err := range tc.Violations {
		msg += err.Error() + "\n"
	}
	return msg
}

func (tc *TypeCheck) HasViolations() bool {
	return len(tc.Violations) > 0
}

type violationWithContext interface {
	SetContext(doc *Document) error
}

func (ds *DocumentSet) Check() TypeCheck { return TypeCheck{} }
func (d *Document) Check() (chk TypeCheck) {
	switch typedContents := d.Value.(type) {
	case Node:
		chk = typedContents.Check()
	}

	return chk
}
func (m *Map) Check() (chk TypeCheck) {
	check := m.Type.CheckType(m)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
		return
	}

	for _, item := range m.Items {
		check = item.Check()
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}
func (mi *MapItem) Check() (chk TypeCheck) {
	check := mi.Type.CheckType(mi)
	if check.HasViolations() {
		chk.Violations = check.Violations
		return
	}

	// If the current value of the item is null
	// there is no extra validation needed Type wise
	if mi.Value == nil {
		return
	}

	check = checkCollectionItem(mi.Value, mi.Type.GetValueType(), mi.Position)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return
}
func (a *Array) Check() (chk TypeCheck) {
	for _, item := range a.Items {
		check := item.Check()
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}
func (ai *ArrayItem) Check() (chk TypeCheck) {
	// TODO: This check only ensures that the ai is of ArrayItem type
	//       which we know because if it was not we would not assign
	//       the type to it.
	//       Given this maybe we can completely remove this check
	//       Lets not forget that the check of the type of the item
	//       is done by checkCollectionItem
	chk = ai.Type.CheckType(ai)
	if chk.HasViolations() {
		return
	}

	check := checkCollectionItem(ai.Value, ai.Type.GetValueType(), ai.Position)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return chk
}

func checkCollectionItem(value interface{}, valueType Type, position *filepos.Position) (chk TypeCheck) {
	switch typedValue := value.(type) {
	case *Map:
		check := typedValue.Check()
		chk.Violations = append(chk.Violations, check.Violations...)

	case *Array:
		check := typedValue.Check()
		chk.Violations = append(chk.Violations, check.Violations...)
	default:
		chk = valueType.CheckType(&Scalar{Value: value, Position: position})
	}
	return chk
}

// Below methods disallow marshaling of nodes directly
var _ []yaml.Marshaler = []yaml.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

func (ds *DocumentSet) MarshalYAML() (interface{}, error) { panic("Unexpected marshaling of docset") }
func (d *Document) MarshalYAML() (interface{}, error)     { panic("Unexpected marshaling of doc") }
func (m *Map) MarshalYAML() (interface{}, error)          { panic("Unexpected marshaling of map") }
func (mi *MapItem) MarshalYAML() (interface{}, error)     { panic("Unexpected marshaling of mapitem") }
func (a *Array) MarshalYAML() (interface{}, error)        { panic("Unexpected marshaling of array") }
func (ai *ArrayItem) MarshalYAML() (interface{}, error)   { panic("Unexpected marshaling of arrayitem") }

// Below methods disallow marshaling of nodes directly
var _ []json.Marshaler = []json.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

func (ds *DocumentSet) MarshalJSON() ([]byte, error) { panic("Unexpected marshaling of docset") }
func (d *Document) MarshalJSON() ([]byte, error)     { panic("Unexpected marshaling of doc") }
func (m *Map) MarshalJSON() ([]byte, error)          { panic("Unexpected marshaling of map") }
func (mi *MapItem) MarshalJSON() ([]byte, error)     { panic("Unexpected marshaling of mapitem") }
func (a *Array) MarshalJSON() ([]byte, error)        { panic("Unexpected marshaling of array") }
func (ai *ArrayItem) MarshalJSON() ([]byte, error)   { panic("Unexpected marshaling of arrayitem") }

func (ds *DocumentSet) _private() {}
func (d *Document) _private()     {}
func (m *Map) _private()          {}
func (mi *MapItem) _private()     {}
func (a *Array) _private()        {}
func (ai *ArrayItem) _private()   {}
