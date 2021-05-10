// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package orderedmap

import (
	"encoding/json"
	"reflect"
)

type Map struct {
	items []MapItem
}

type MapItem struct {
	Key   interface{}
	Value interface{}
}

func NewMap() *Map {
	return &Map{}
}

func NewMapWithItems(items []MapItem) *Map {
	return &Map{items}
}

func (m *Map) Set(key, value interface{}) {
	for i, item := range m.items {
		if m.isKeyEq(item.Key, key) {
			item.Value = value
			m.items[i] = item
			return
		}
	}
	m.items = append(m.items, MapItem{key, value})
}

func (m *Map) Get(key interface{}) (interface{}, bool) {
	for _, item := range m.items {
		if m.isKeyEq(item.Key, key) {
			return item.Value, true
		}
	}
	return nil, false
}

func (m *Map) Delete(key interface{}) bool {
	for i, item := range m.items {
		if m.isKeyEq(item.Key, key) {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return true
		}
	}
	return false
}

func (m *Map) isKeyEq(key1, key2 interface{}) bool {
	return reflect.DeepEqual(key1, key2)
}

func (m *Map) Keys() (keys []interface{}) {
	m.Iterate(func(k, _ interface{}) {
		keys = append(keys, k)
	})
	return
}

func (m *Map) Iterate(iterFunc func(k, v interface{})) {
	for _, item := range m.items {
		iterFunc(item.Key, item.Value)
	}
}

func (m *Map) IterateErr(iterFunc func(k, v interface{}) error) error {
	for _, item := range m.items {
		err := iterFunc(item.Key, item.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Map) Len() int { return len(m.items) }

// Below methods disallow marshaling of Map directly
// TODO yaml library is not imported here
// var _ []yaml.Marshaler = []yaml.Marshaler{&Map{}}
var _ []json.Marshaler = []json.Marshaler{&Map{}}

func (*Map) MarshalYAML() (interface{}, error) { panic("Unexpected marshaling of *orderedmap.Map") }
func (*Map) MarshalJSON() ([]byte, error)      { panic("Unexpected marshaling of *orderedmap.Map") }
