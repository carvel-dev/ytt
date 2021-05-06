// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/orderedmap"
)

type StarlarkStruct struct {
	data *orderedmap.Map // [string]starlark.Value; most common usage: HasAttrs
}

func NewStarlarkStruct(goStringKeyToStarlarkValue *orderedmap.Map) *StarlarkStruct {
	return &StarlarkStruct{data: goStringKeyToStarlarkValue}
}

var _ starlark.Value = (*StarlarkStruct)(nil)
var _ starlark.HasAttrs = (*StarlarkStruct)(nil)
var _ starlark.IterableMapping = (*StarlarkStruct)(nil)
var _ starlark.Sequence = (*StarlarkStruct)(nil)

func (s *StarlarkStruct) String() string        { return "struct(...)" }
func (s *StarlarkStruct) Type() string          { return "struct" }
func (s *StarlarkStruct) Freeze()               {} // TODO
func (s *StarlarkStruct) Truth() starlark.Bool  { return s.data.Len() > 0 }
func (s *StarlarkStruct) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: struct") }
func (s *StarlarkStruct) Len() int              { return s.data.Len() }

// returns (nil, nil) if attribute not present
func (s *StarlarkStruct) Attr(name string) (starlark.Value, error) {
	val, found := s.data.Get(name)
	if found {
		return val.(starlark.Value), nil
	}
	return nil, nil
}

// callers must not modify the result.
func (s *StarlarkStruct) AttrNames() []string {
	var keys []string
	s.data.Iterate(func(key, _ interface{}) {
		keys = append(keys, key.(string))
	})
	return keys
}

func (s *StarlarkStruct) Get(key starlark.Value) (val starlark.Value, found bool, err error) {
	obj, err := NewStarlarkValue(key).AsGoValue()
	if err != nil {
		return nil, false, err
	}
	if attr, ok := obj.(string); ok {
		val, found := s.data.Get(attr)
		if found {
			return val.(starlark.Value), true, nil
		}
		return starlark.None, false, nil
	}
	return nil, false, fmt.Errorf("expected key `%s` to be a string but is a %T", key, key)
}

func (s *StarlarkStruct) Iterate() starlark.Iterator {
	return &StarlarkStructIterator{
		keys: s.data.Keys(),
	}
}

func (s *StarlarkStruct) Items() (items []starlark.Tuple) {
	s.data.Iterate(func(key, val interface{}) {
		items = append(items, starlark.Tuple{
			NewGoValue(key).AsStarlarkValue(),
			val.(starlark.Value),
		})
	})
	return
}

type StarlarkStructIterator struct {
	keys []interface{}
	idx  int
}

var _ starlark.Iterator = &StarlarkStructIterator{}

func (s *StarlarkStructIterator) Next(p *starlark.Value) bool {
	if s.idx < len(s.keys) {
		*p = NewGoValue(s.keys[s.idx]).AsStarlarkValue()
		s.idx++
		return true
	}
	return false
}

func (s *StarlarkStructIterator) Done() { /* intentionally blank. */ }
