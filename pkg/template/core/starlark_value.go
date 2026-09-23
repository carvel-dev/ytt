// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"errors"
	"fmt"

	"carvel.dev/ytt/pkg/orderedmap"
	"github.com/k14s/starlark-go/starlark"
)

type StarlarkValueToGoValueConversion interface {
	AsGoValue() (interface{}, error)
}

var _ StarlarkValueToGoValueConversion = &StarlarkValue{}

type UnconvertableStarlarkValue interface {
	ConversionHint() string
}

type StarlarkValue struct {
	val starlark.Value
}

func NewStarlarkValue(val starlark.Value) StarlarkValue {
	return StarlarkValue{val}
}

func (e StarlarkValue) AsGoValue() (interface{}, error) {
	return e.asInterface(e.val, nil)
}

func (e StarlarkValue) AsString() (string, error) {
	if typedVal, ok := e.val.(starlark.String); ok {
		return string(typedVal), nil
	}
	return "", fmt.Errorf("expected a string, but was %s", e.val.Type())
}

func (e StarlarkValue) AsBool() (bool, error) {
	if typedVal, ok := e.val.(starlark.Bool); ok {
		return bool(typedVal), nil
	}
	return false, fmt.Errorf("expected starlark.Bool, but was %T", e.val)
}

func (e StarlarkValue) AsInt64() (int64, error) {
	if typedVal, ok := e.val.(starlark.Int); ok {
		i1, ok := typedVal.Int64()
		if ok {
			return i1, nil
		}
		return 0, fmt.Errorf("expected int64 value")
	}
	return 0, fmt.Errorf("expected starlark.Int")
}

// AsFloat64 converts a Starlark number (either int or float) to the corresponding Go double-precision float.
func (e StarlarkValue) AsFloat64() (float64, error) {
	switch e := e.val.(type) {
	case starlark.Int:
		return float64(e.Float()), nil
	case starlark.Float:
		return float64(e), nil
	}
	return 0, fmt.Errorf("expected float value, but was %T", e.val)
}

// path holds the *starlark.Dict and *starlark.List values currently being
// converted, so that a value that (directly or indirectly) contains itself
// is reported as an error instead of recursing until the stack overflows.
// Dict and List are the only Starlark values that can be mutated after
// creation, so they are the only ones that can participate in a cycle.
func (e StarlarkValue) asInterface(
	val starlark.Value, path []starlark.Value,
) (any, error) {
	if obj, ok := val.(UnconvertableStarlarkValue); ok {
		return nil, fmt.Errorf("Unable to convert value: %s", obj.ConversionHint())
	}
	if obj, ok := val.(StarlarkValueToGoValueConversion); ok {
		return obj.AsGoValue()
	}
	if result, handled, err := scalarAsInterface(val); handled {
		return result, err
	}

	switch typedVal := val.(type) {
	case *starlark.Dict:
		return e.dictAsInterface(typedVal, path)

	case *StarlarkStruct:
		return e.structAsInterface(typedVal, path)

	case *starlark.List:
		return e.itearableAsInterface(typedVal, path)

	case starlark.Tuple:
		return e.itearableAsInterface(typedVal, path)

	case *starlark.Set:
		return e.itearableAsInterface(typedVal, path)

	default:
		panic(fmt.Sprintf("unknown type %T for conversion to go value", val))
	}
}

// scalarAsInterface converts a scalar Starlark value to its Go equivalent.
// handled is false if val is not one of the types this function handles.
func scalarAsInterface(
	val starlark.Value,
) (result any, handled bool, err error) {
	switch typedVal := val.(type) {
	case nil, starlark.NoneType:
		return nil, true, nil // TODO is it nil or is it None

	case starlark.Bool:
		return bool(typedVal), true, nil

	case starlark.String:
		return string(typedVal), true, nil

	case starlark.Int:
		i1, ok := typedVal.Int64()
		if ok {
			return i1, true, nil
		}
		i2, ok := typedVal.Uint64()
		if ok {
			return i2, true, nil
		}
		panic("not sure how to get int") // TODO

	case starlark.Float:
		return float64(typedVal), true, nil

	default:
		return nil, false, nil
	}
}

func (e StarlarkValue) dictAsInterface(
	val *starlark.Dict, path []starlark.Value,
) (any, error) {
	if pathContains(path, val) {
		return nil, errors.New("Unable to convert value: self-referencing dict")
	}
	path = append(path, val)

	result := orderedmap.NewMap()
	for _, item := range val.Items() {
		key, value, err := e.dictItemAsInterface(item, path)
		if err != nil {
			return nil, err
		}
		result.Set(key, value)
	}
	return result, nil
}

// dictItemKeyValueLen is the length of a starlark.Tuple representing a
// single dict item: exactly one key and one value.
const dictItemKeyValueLen = 2

func (e StarlarkValue) dictItemAsInterface(
	item starlark.Tuple, path []starlark.Value,
) (key, value any, err error) {
	if item.Len() != dictItemKeyValueLen {
		panic("dict item is not KV")
	}
	key, err = e.asInterface(item.Index(0), path)
	if err != nil {
		return nil, nil, err
	}
	value, err = e.asInterface(item.Index(1), path)
	if err != nil {
		return nil, nil, err
	}
	return key, value, nil
}

func (e StarlarkValue) structAsInterface(
	val *StarlarkStruct, path []starlark.Value,
) (any, error) {
	// TODO accessing privates
	result := orderedmap.NewMap()
	err := val.data.IterateErr(func(k, v interface{}) error {
		value, err := e.asInterface(v.(starlark.Value), path)
		if err == nil {
			result.Set(k, value)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e StarlarkValue) itearableAsInterface(
	iterable starlark.Iterable, path []starlark.Value,
) (any, error) {
	path, err := extendPathWithList(path, iterable)
	if err != nil {
		return nil, err
	}

	iter := iterable.Iterate()
	defer iter.Done()

	var result []interface{}
	var x starlark.Value
	for iter.Next(&x) {
		elem, err := e.asInterface(x, path)
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}
	return result, nil
}

// extendPathWithList appends list to path, erroring out if it is already
// present (i.e. list directly or indirectly contains itself). Non-list
// iterables are returned unchanged, since only *starlark.List and
// *starlark.Dict are mutable enough to form a cycle.
func extendPathWithList(
	path []starlark.Value, iterable starlark.Iterable,
) ([]starlark.Value, error) {
	list, ok := iterable.(*starlark.List)
	if !ok {
		return path, nil
	}
	if pathContains(path, list) {
		return nil, errors.New("Unable to convert value: self-referencing list")
	}
	return append(path, list), nil
}

// pathContains reports whether x is already present in path.
// Every value ever added to path is a *starlark.Dict or *starlark.List,
// both of which are comparable, so this equality check is safe.
func pathContains(path []starlark.Value, x starlark.Value) bool {
	for _, y := range path {
		if x == y {
			return true
		}
	}
	return false
}
