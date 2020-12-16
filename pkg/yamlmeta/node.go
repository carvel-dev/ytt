// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"encoding/json"
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func (n *DocumentSet) GetPosition() *filepos.Position { return n.Position }
func (n *Document) GetPosition() *filepos.Position    { return n.Position }
func (n *Map) GetPosition() *filepos.Position         { return n.Position }
func (n *MapItem) GetPosition() *filepos.Position     { return n.Position }
func (n *Array) GetPosition() *filepos.Position       { return n.Position }
func (n *ArrayItem) GetPosition() *filepos.Position   { return n.Position }

func (n *DocumentSet) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a documentset")
}

func (n *Document) SetValue(val interface{}) error {
	n.Value = val
	return nil
}

func (n *Map) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a map")
}

func (n *MapItem) SetValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot set map-or-array-item value (%T) into mapitem", val)
	}
	n.Value = val
	return nil
}

func (n *Array) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on an array")
}

func (n *ArrayItem) SetValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot set map-or-array-item value (%T) into arrayitem", val)
	}
	n.Value = val
	return nil
}

func (n *DocumentSet) ResetValue() { n.Items = nil }
func (n *Document) ResetValue()    { n.Value = nil }
func (n *Map) ResetValue()         { n.Items = nil }
func (n *MapItem) ResetValue()     { n.Value = nil }
func (n *Array) ResetValue()       { n.Items = nil }
func (n *ArrayItem) ResetValue()   { n.Value = nil }

func (n *DocumentSet) AddValue(val interface{}) error {
	if item, ok := val.(*Document); ok {
		n.Items = append(n.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-document value (%T) into documentset", val)
}

func (n *Document) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into document", val)
	}
	n.Value = val
	return nil
}

func (n *Map) AddValue(val interface{}) error {
	if item, ok := val.(*MapItem); ok {
		n.Items = append(n.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-map-item value (%T) into map", val)
}

func (n *MapItem) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into mapitem", val)
	}
	n.Value = val
	return nil
}

func (n *Array) AddValue(val interface{}) error {
	if item, ok := val.(*ArrayItem); ok {
		n.Items = append(n.Items, item)
		return nil
	}
	return fmt.Errorf("cannot add non-array-item value (%T) into array", val)
}

func (n *ArrayItem) AddValue(val interface{}) error {
	if isMapOrArrayItem(val) {
		return fmt.Errorf("cannot add map-or-array-item value (%T) into arrayitem", val)
	}
	n.Value = val
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

func (n *DocumentSet) GetValues() []interface{} {
	var result []interface{}
	for _, item := range n.Items {
		result = append(result, item)
	}
	return result
}

func (n *Document) GetValues() []interface{} { return []interface{}{n.Value} }

func (n *Map) GetValues() []interface{} {
	var result []interface{}
	for _, item := range n.Items {
		result = append(result, item)
	}
	return result
}

func (n *MapItem) GetValues() []interface{} { return []interface{}{n.Value} }

func (n *Array) GetValues() []interface{} {
	var result []interface{}
	for _, item := range n.Items {
		result = append(result, item)
	}
	return result
}

func (n *ArrayItem) GetValues() []interface{} { return []interface{}{n.Value} }
func (s *Scalar) GetValues() []interface{}    { return []interface{}{s.Value} }

func (n *DocumentSet) GetMetas() []*Meta { return n.Metas }
func (n *Document) GetMetas() []*Meta    { return n.Metas }
func (n *Map) GetMetas() []*Meta         { return n.Metas }
func (n *MapItem) GetMetas() []*Meta     { return n.Metas }
func (n *Array) GetMetas() []*Meta       { return n.Metas }
func (n *ArrayItem) GetMetas() []*Meta   { return n.Metas }

func (n *DocumentSet) addMeta(meta *Meta) { n.Metas = append(n.Metas, meta) }
func (n *Document) addMeta(meta *Meta)    { n.Metas = append(n.Metas, meta) }
func (n *Map) addMeta(meta *Meta) {
	panic(fmt.Sprintf("Attempted to attach metadata (%s) to Map (%v); maps cannot carry metadata", meta.Data, n))
}
func (n *MapItem) addMeta(meta *Meta) { n.Metas = append(n.Metas, meta) }
func (n *Array) addMeta(meta *Meta) {
	panic(fmt.Sprintf("Attempted to attach metadata (%s) to Array (%v); arrays cannot carry metadata", meta.Data, n))
}
func (n *ArrayItem) addMeta(meta *Meta) { n.Metas = append(n.Metas, meta) }

func (n *DocumentSet) GetAnnotations() interface{} { return n.annotations }
func (n *Document) GetAnnotations() interface{}    { return n.annotations }
func (n *Map) GetAnnotations() interface{}         { return n.annotations }
func (n *MapItem) GetAnnotations() interface{}     { return n.annotations }
func (n *Array) GetAnnotations() interface{}       { return n.annotations }
func (n *ArrayItem) GetAnnotations() interface{}   { return n.annotations }

func (n *DocumentSet) SetAnnotations(anns interface{}) { n.annotations = anns }
func (n *Document) SetAnnotations(anns interface{})    { n.annotations = anns }
func (n *Map) SetAnnotations(anns interface{})         { n.annotations = anns }
func (n *MapItem) SetAnnotations(anns interface{})     { n.annotations = anns }
func (n *Array) SetAnnotations(anns interface{})       { n.annotations = anns }
func (n *ArrayItem) SetAnnotations(anns interface{})   { n.annotations = anns }

type TypeCheck struct {
	Violations []string
}

func (tc *TypeCheck) HasViolations() bool {
	return len(tc.Violations) > 0
}

func (n *Document) Check() (chk TypeCheck) {
	switch typedContents := n.Value.(type) {
	case Node:
		chk = typedContents.Check()
	}

	return
}
func (n *Map) Check() (chk TypeCheck) {
	check := n.Type.CheckType(n, "")
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
		return
	}

	for _, item := range n.Items {
		check = item.Check()
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}
func (n *MapItem) Check() (chk TypeCheck) {
	mapItemViolation := fmt.Sprintf("Map item '%s' at %s", n.Key, n.Position.AsCompactString())

	check := n.Type.CheckType(n, mapItemViolation)
	if check.HasViolations() {
		chk.Violations = check.Violations
		return
	}

	// If the current value of the item is null
	// there is no extra validation needed Type wise
	if n.Value == nil {
		return
	}

	check = checkCollectionItem(n.Value, n.Type.GetValueType(), mapItemViolation)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return
}
func (n *Array) Check() (chk TypeCheck) {
	for _, item := range n.Items {
		check := item.Check()
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}
func (n *ArrayItem) Check() (chk TypeCheck) {
	arrayItemViolation := fmt.Sprintf("Array item at %s", n.Position.AsCompactString())

	// TODO: This check only ensures that the n is of ArrayItem type
	//       which we know because if it was not we would not assign
	//       the type to it.
	//       Given this maybe we can completely remove this check
	//       Lets not forget that the check of the type of the item
	//       is done by checkCollectionItem
	chk = n.Type.CheckType(n, arrayItemViolation)
	if chk.HasViolations() {
		return
	}

	check := checkCollectionItem(n.Value, n.Type.GetValueType(), arrayItemViolation)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return chk
}

func checkCollectionItem(value interface{}, valueType Type, violationErrorMessage string) (chk TypeCheck) {
	switch typedValue := value.(type) {
	case *Map:
		check := typedValue.Check()
		chk.Violations = append(chk.Violations, check.Violations...)

	case *Array:
		check := typedValue.Check()
		chk.Violations = append(chk.Violations, check.Violations...)
	default:
		chk = valueType.CheckType(&Scalar{Value: value}, violationErrorMessage)
	}
	return chk
}

func (n *DocumentSet) Check() TypeCheck { return TypeCheck{} }

// Below methods disallow marshaling of nodes directly
var _ []yaml.Marshaler = []yaml.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

func (n *DocumentSet) MarshalYAML() (interface{}, error) { panic("Unexpected marshaling of docset") }
func (n *Document) MarshalYAML() (interface{}, error)    { panic("Unexpected marshaling of doc") }
func (n *Map) MarshalYAML() (interface{}, error)         { panic("Unexpected marshaling of map") }
func (n *MapItem) MarshalYAML() (interface{}, error)     { panic("Unexpected marshaling of mapitem") }
func (n *Array) MarshalYAML() (interface{}, error)       { panic("Unexpected marshaling of array") }
func (n *ArrayItem) MarshalYAML() (interface{}, error)   { panic("Unexpected marshaling of arrayitem") }

// Below methods disallow marshaling of nodes directly
var _ []json.Marshaler = []json.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

func (n *DocumentSet) MarshalJSON() ([]byte, error) { panic("Unexpected marshaling of docset") }
func (n *Document) MarshalJSON() ([]byte, error)    { panic("Unexpected marshaling of doc") }
func (n *Map) MarshalJSON() ([]byte, error)         { panic("Unexpected marshaling of map") }
func (n *MapItem) MarshalJSON() ([]byte, error)     { panic("Unexpected marshaling of mapitem") }
func (n *Array) MarshalJSON() ([]byte, error)       { panic("Unexpected marshaling of array") }
func (n *ArrayItem) MarshalJSON() ([]byte, error)   { panic("Unexpected marshaling of arrayitem") }

func (n *DocumentSet) _private() {}
func (n *Document) _private()    {}
func (n *Map) _private()         {}
func (n *MapItem) _private()     {}
func (n *Array) _private()       {}
func (n *ArrayItem) _private()   {}
