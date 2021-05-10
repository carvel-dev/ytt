// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"
	"reflect"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/syntax"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	starlarkFragmentType = "yamlfragment"
)

type StarlarkFragment struct {
	data interface{}
}

var _ starlark.Value = &StarlarkFragment{}
var _ starlark.Comparable = (*StarlarkFragment)(nil)
var _ starlark.Sequence = (*StarlarkFragment)(nil)
var _ starlark.IterableMapping = (*StarlarkFragment)(nil)
var _ starlark.HasSetKey = (*StarlarkFragment)(nil)
var _ starlark.HasSetIndex = (*StarlarkFragment)(nil)
var _ starlark.Sliceable = (*StarlarkFragment)(nil)

var _ tplcore.StarlarkValueToGoValueConversion = &StarlarkFragment{}
var _ tplcore.GoValueToStarlarkValueConversion = &StarlarkFragment{}

func NewStarlarkFragment(data interface{}) *StarlarkFragment {
	return &StarlarkFragment{data}
}

func (s *StarlarkFragment) String() string {
	return fmt.Sprintf("%s(%T)", starlarkFragmentType, s.data)
}
func (s *StarlarkFragment) Type() string         { return starlarkFragmentType }
func (s *StarlarkFragment) Freeze()              {} // TODO
func (s *StarlarkFragment) Truth() starlark.Bool { return starlark.Bool(s.Len() > 0) }
func (s *StarlarkFragment) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: %s", starlarkFragmentType)
}

func (s *StarlarkFragment) AsGoValue() (interface{}, error) { return s.data, nil }
func (s *StarlarkFragment) AsStarlarkValue() starlark.Value { return s }

func (s *StarlarkFragment) CompareSameType(op syntax.Token, y starlark.Value, depth int) (bool, error) {
	return false, fmt.Errorf("%s.CompareSameType: Not implemented", starlarkFragmentType) // TODO
}

func (s *StarlarkFragment) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	wantedKey, err := tplcore.NewStarlarkValue(k).AsGoValue()
	if err != nil {
		return starlark.None, false, err
	}

	switch typedData := s.data.(type) {
	case nil:
		// do nothing

	case *yamlmeta.Map:
		for _, item := range typedData.Items {
			if reflect.DeepEqual(item.Key, wantedKey) {
				return NewGoValueWithYAML(item.Value).AsStarlarkValue(), true, nil
			}
		}

	case *yamlmeta.Array:
		wantedInt, ok := wantedKey.(int64)
		if !ok {
			return starlark.None, false, fmt.Errorf(
				"%s.Get: Expected array index to be an int64, but was %T", starlarkFragmentType, wantedKey)
		}

		for i, item := range typedData.Items {
			if wantedInt == int64(i) {
				return NewGoValueWithYAML(item.Value).AsStarlarkValue(), true, nil
			}
		}

	case *yamlmeta.DocumentSet:
		wantedInt, ok := wantedKey.(int64)
		if !ok {
			return starlark.None, false, fmt.Errorf(
				"%s.Get: Expected document set index to be an int64, but was %T", wantedKey, starlarkFragmentType)
		}

		for i, item := range typedData.Items {
			if wantedInt == int64(i) {
				return NewGoValueWithYAML(item.Value).AsStarlarkValue(), true, nil
			}
		}

	default:
		panic(fmt.Sprintf("%s.Get: Expected value to be a map, array or docset, but was %T",
			starlarkFragmentType, s.data))
	}

	return starlark.None, false, nil
}

func (s *StarlarkFragment) SetKey(k, v starlark.Value) error {
	return fmt.Errorf("%s.SetKey: Not implemented", starlarkFragmentType) // TODO
}

func (s *StarlarkFragment) Index(i int) starlark.Value {
	switch typedData := s.data.(type) {
	case *yamlmeta.Array:
		return NewGoValueWithYAML(typedData.Items[i].Value).AsStarlarkValue()
	case *yamlmeta.DocumentSet:
		return NewGoValueWithYAML(typedData.Items[i].Value).AsStarlarkValue()
	default:
		panic(fmt.Sprintf("%s.Index: Expected value to be a array or docset, but was %T", starlarkFragmentType, s.data))
	}
}

func (s *StarlarkFragment) SetIndex(index int, v starlark.Value) error {
	return fmt.Errorf("%s.SetIndex: Not implemented", starlarkFragmentType) // TODO
}

func (s *StarlarkFragment) Len() int {
	switch typedData := s.data.(type) {
	case nil:
		return 0
	case *yamlmeta.Map:
		return len(typedData.Items)
	case *yamlmeta.Array:
		return len(typedData.Items)
	case *yamlmeta.DocumentSet:
		return len(typedData.Items)
	default:
		panic(fmt.Sprintf("%s.Len: Expected value to be a map, array or docset, but was %T", starlarkFragmentType, s.data))
	}
}

// Items seems to be only used for splatting kwargs
func (s *StarlarkFragment) Items() []starlark.Tuple {
	switch typedData := s.data.(type) {
	case nil:
		return []starlark.Tuple{}

	case *yamlmeta.Map:
		var result []starlark.Tuple
		for _, item := range typedData.Items {
			result = append(result, starlark.Tuple{
				NewGoValueWithYAML(item.Key).AsStarlarkValue(),
				NewGoValueWithYAML(item.Value).AsStarlarkValue(),
			})
		}
		return result

	default:
		panic(fmt.Sprintf("%s.Items: Expected value to be a map, but was %T", starlarkFragmentType, s.data))
	}
}

func (s *StarlarkFragment) Slice(start, end, step int) starlark.Value {
	panic(fmt.Sprintf("%s.Slice: Not implemented", starlarkFragmentType)) // TODO
}

func (s *StarlarkFragment) Iterate() starlark.Iterator {
	switch typedData := s.data.(type) {
	case nil:
		return StarlarkFragmentNilIterator{}
	case *yamlmeta.Map:
		return &StarlarkFragmentKeysIterator{data: typedData}
	case *yamlmeta.Array:
		return &StarlarkFragmentValuesIterator{data: typedData}
	case *yamlmeta.DocumentSet:
		return &StarlarkFragmentValuesIterator{data: typedData}
	default:
		panic(fmt.Sprintf("%s.Iterate: Expected value to be a map, array or docset, but was %T", starlarkFragmentType, s.data))
	}
}

type StarlarkFragmentNilIterator struct{}

func (s StarlarkFragmentNilIterator) Next(p *starlark.Value) bool { return false }
func (s StarlarkFragmentNilIterator) Done()                       {}

type StarlarkFragmentKeysIterator struct {
	data *yamlmeta.Map
	idx  int
}

func (s *StarlarkFragmentKeysIterator) Next(p *starlark.Value) bool {
	if s.idx < len(s.data.Items) {
		var val starlark.Value = NewGoValueWithYAML(s.data.Items[s.idx].Key).AsStarlarkValue()
		*p = val
		s.idx++
		return true
	}
	return false
}

func (s *StarlarkFragmentKeysIterator) Done() {}

type StarlarkFragmentValuesIterator struct {
	data yamlmeta.Node
	idx  int
}

func (s *StarlarkFragmentValuesIterator) Next(p *starlark.Value) bool {
	if s.idx < len(s.data.GetValues()) {
		var val starlark.Value

		switch typedData := s.data.(type) {
		case *yamlmeta.Array:
			val = NewGoValueWithYAML(typedData.Items[s.idx].Value).AsStarlarkValue()
		case *yamlmeta.DocumentSet:
			val = NewGoValueWithYAML(typedData.Items[s.idx].Value).AsStarlarkValue()
		default:
			panic(fmt.Sprintf("%s.Next: Expected value to be a array or docset, but was %T", starlarkFragmentType, s.data))
		}
		*p = val
		s.idx++
		return true
	}
	return false
}

func (s *StarlarkFragmentValuesIterator) Done() {}
