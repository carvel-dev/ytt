// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"encoding/json"
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

// TypeName returns the user-friendly name of the type of `val`
func TypeName(val interface{}) string {
	switch val.(type) {
	case *DocumentSet:
		return "document set"
	case *Document:
		return "document"
	case *Map:
		return "map"
	case *MapItem:
		return "map item"
	case *Array:
		return "array"
	case *ArrayItem:
		return "array item"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "float"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", val)
	}
}

func (ds *DocumentSet) GetPosition() *filepos.Position { return ds.Position }
func (d *Document) GetPosition() *filepos.Position     { return d.Position }
func (m *Map) GetPosition() *filepos.Position          { return m.Position }
func (mi *MapItem) GetPosition() *filepos.Position     { return mi.Position }
func (a *Array) GetPosition() *filepos.Position        { return a.Position }
func (ai *ArrayItem) GetPosition() *filepos.Position   { return ai.Position }

func (ds *DocumentSet) SetPosition(position *filepos.Position) { ds.Position = position }
func (d *Document) SetPosition(position *filepos.Position)     { d.Position = position }
func (m *Map) SetPosition(position *filepos.Position)          { m.Position = position }
func (mi *MapItem) SetPosition(position *filepos.Position)     { mi.Position = position }
func (a *Array) SetPosition(position *filepos.Position)        { a.Position = position }
func (ai *ArrayItem) SetPosition(position *filepos.Position)   { ai.Position = position }

func (ds *DocumentSet) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a %s", TypeName(ds))
}

func (d *Document) SetValue(val interface{}) error {
	d.ResetValue()
	return d.AddValue(val)
}

func (m *Map) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a %s", TypeName(m))
}

func (mi *MapItem) SetValue(val interface{}) error {
	mi.ResetValue()
	return mi.AddValue(val)
}

func (a *Array) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on an %s", TypeName(a))
}

func (ai *ArrayItem) SetValue(val interface{}) error {
	ai.ResetValue()
	return ai.AddValue(val)
}

func isValidValue(val interface{}) bool {
	switch val.(type) {
	case *Map, *orderedmap.Map,
		*Array, []interface{},
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool,
		string,
		nil:
		return true
	default:
		return false
	}
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
	return fmt.Errorf("document sets can only contain documents; this is a %s", TypeName(val))
}

func (d *Document) AddValue(val interface{}) error {
	if !isValidValue(val) {
		return fmt.Errorf("documents can only contain arrays, maps, or scalars; this is a %s", TypeName(val))
	}
	d.Value = val
	return nil
}

func (m *Map) AddValue(val interface{}) error {
	if item, ok := val.(*MapItem); ok {
		m.Items = append(m.Items, item)
		return nil
	}
	return fmt.Errorf("maps can only contain map items; this is a %s", TypeName(val))
}

func (mi *MapItem) AddValue(val interface{}) error {
	if !isValidValue(val) {
		return fmt.Errorf("map items can only contain arrays, maps, or scalars; this is a %s", TypeName(val))
	}
	mi.Value = val
	return nil
}

func (a *Array) AddValue(val interface{}) error {
	if item, ok := val.(*ArrayItem); ok {
		a.Items = append(a.Items, item)
		return nil
	}
	return fmt.Errorf("arrays can only contain array items; this is a %s", TypeName(val))
}

func (ai *ArrayItem) AddValue(val interface{}) error {
	if !isValidValue(val) {
		return fmt.Errorf("array items can only contain maps, arrays, or scalars; this is a %s", TypeName(val))
	}
	ai.Value = val
	return nil
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

func (ds *DocumentSet) GetComments() []*Comment { return ds.Comments }
func (d *Document) GetComments() []*Comment     { return d.Comments }
func (m *Map) GetComments() []*Comment          { return m.Comments }
func (mi *MapItem) GetComments() []*Comment     { return mi.Comments }
func (a *Array) GetComments() []*Comment        { return a.Comments }
func (ai *ArrayItem) GetComments() []*Comment   { return ai.Comments }

// SetComments replaces comments with "c"
func (ds *DocumentSet) SetComments(c []*Comment) { ds.Comments = c }

// SetComments replaces comments with "c"
func (d *Document) SetComments(c []*Comment) { d.Comments = c }

// SetComments replaces comments with "c"
func (m *Map) SetComments(c []*Comment) { m.Comments = c }

// SetComments replaces comments with "c"
func (mi *MapItem) SetComments(c []*Comment) { mi.Comments = c }

// SetComments replaces comments with "c"
func (a *Array) SetComments(c []*Comment) { a.Comments = c }

// SetComments replaces comments with "c"
func (ai *ArrayItem) SetComments(c []*Comment) { ai.Comments = c }

func (ds *DocumentSet) addComments(comment *Comment) { ds.Comments = append(ds.Comments, comment) }
func (d *Document) addComments(comment *Comment)     { d.Comments = append(d.Comments, comment) }
func (m *Map) addComments(comment *Comment) {
	panic(fmt.Sprintf("Attempted to attach (%#v) to (%#v)", comment, m))
}
func (mi *MapItem) addComments(comment *Comment) { mi.Comments = append(mi.Comments, comment) }
func (a *Array) addComments(comment *Comment) {
	panic(fmt.Sprintf("Attempted to attach (%#v) to (%#v)", comment, a))
}
func (ai *ArrayItem) addComments(comment *Comment) { ai.Comments = append(ai.Comments, comment) }

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

// Below methods disallow marshaling of nodes directly
var _ []yaml.Marshaler = []yaml.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (ds *DocumentSet) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", ds))
}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (d *Document) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", d))
}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (m *Map) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", m))
}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (mi *MapItem) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", mi))
}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (a *Array) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", a))
}

// MarshalYAML panics because Nodes cannot be marshalled directly.
func (ai *ArrayItem) MarshalYAML() (interface{}, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", ai))
}

// Below methods disallow marshaling of nodes directly
var _ []json.Marshaler = []json.Marshaler{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (ds *DocumentSet) MarshalJSON() ([]byte, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", ds))
}

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (d *Document) MarshalJSON() ([]byte, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", d))
}

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (m *Map) MarshalJSON() ([]byte, error) { panic(fmt.Sprintf("Unexpected marshaling of %T", m)) }

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (mi *MapItem) MarshalJSON() ([]byte, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", mi))
}

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (a *Array) MarshalJSON() ([]byte, error) { panic(fmt.Sprintf("Unexpected marshaling of %T", a)) }

// MarshalJSON panics because Nodes cannot be marshalled directly.
func (ai *ArrayItem) MarshalJSON() ([]byte, error) {
	panic(fmt.Sprintf("Unexpected marshaling of %T", ai))
}

func (ds *DocumentSet) sealed() {}
func (d *Document) sealed()     {}
func (m *Map) sealed()          {}
func (mi *MapItem) sealed()     {}
func (a *Array) sealed()        {}
func (ai *ArrayItem) sealed()   {}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (ds *DocumentSet) GetMeta(name string) interface{} {
	if ds.meta == nil {
		ds.meta = make(map[string]interface{})
	}
	return ds.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (ds *DocumentSet) SetMeta(name string, data interface{}) {
	if ds.meta == nil {
		ds.meta = make(map[string]interface{})
	}
	ds.meta[name] = data
}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (d *Document) GetMeta(name string) interface{} {
	if d.meta == nil {
		d.meta = make(map[string]interface{})
	}
	return d.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (d *Document) SetMeta(name string, data interface{}) {
	if d.meta == nil {
		d.meta = make(map[string]interface{})
	}
	d.meta[name] = data
}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (m *Map) GetMeta(name string) interface{} {
	if m.meta == nil {
		m.meta = make(map[string]interface{})
	}
	return m.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (m *Map) SetMeta(name string, data interface{}) {
	if m.meta == nil {
		m.meta = make(map[string]interface{})
	}
	m.meta[name] = data
}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (mi *MapItem) GetMeta(name string) interface{} {
	if mi.meta == nil {
		mi.meta = make(map[string]interface{})
	}
	return mi.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (mi *MapItem) SetMeta(name string, data interface{}) {
	if mi.meta == nil {
		mi.meta = make(map[string]interface{})
	}
	mi.meta[name] = data
}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (a *Array) GetMeta(name string) interface{} {
	if a.meta == nil {
		a.meta = make(map[string]interface{})
	}
	return a.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (a *Array) SetMeta(name string, data interface{}) {
	if a.meta == nil {
		a.meta = make(map[string]interface{})
	}
	a.meta[name] = data
}

// GetMeta returns the metadata named `name` that was previously attached via SetMeta()
func (ai *ArrayItem) GetMeta(name string) interface{} {
	if ai.meta == nil {
		ai.meta = make(map[string]interface{})
	}
	return ai.meta[name]
}

// SetMeta attaches metadata identified by `name` than can later be retrieved via GetMeta()
func (ai *ArrayItem) SetMeta(name string, data interface{}) {
	if ai.meta == nil {
		ai.meta = make(map[string]interface{})
	}
	ai.meta[name] = data
}
