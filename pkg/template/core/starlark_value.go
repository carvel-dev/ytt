// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/orderedmap"
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
	return e.asInterface(e.val)
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

func (e StarlarkValue) asInterface(val starlark.Value) (interface{}, error) {
	if obj, ok := val.(UnconvertableStarlarkValue); ok {
		return nil, fmt.Errorf("Unable to convert value: %s", obj.ConversionHint())
	}
	if obj, ok := val.(StarlarkValueToGoValueConversion); ok {
		return obj.AsGoValue()
	}

	switch typedVal := val.(type) {
	case nil, starlark.NoneType:
		return nil, nil // TODO is it nil or is it None

	case starlark.Bool:
		return bool(typedVal), nil

	case starlark.String:
		return string(typedVal), nil

	case starlark.Int:
		i1, ok := typedVal.Int64()
		if ok {
			return i1, nil
		}
		i2, ok := typedVal.Uint64()
		if ok {
			return i2, nil
		}
		panic("not sure how to get int") // TODO

	case starlark.Float:
		return float64(typedVal), nil

	case *starlark.Dict:
		return e.dictAsInterface(typedVal)

	case *StarlarkStruct:
		return e.structAsInterface(typedVal)

	case *starlark.List:
		return e.itearableAsInterface(typedVal)

	case starlark.Tuple:
		return e.itearableAsInterface(typedVal)

	case *starlark.Set:
		return e.itearableAsInterface(typedVal)

	default:
		panic(fmt.Sprintf("unknown type %T for conversion to go value", val))
	}
}

func (e StarlarkValue) dictAsInterface(val *starlark.Dict) (interface{}, error) {
	result := orderedmap.NewMap()
	for _, item := range val.Items() {
		if item.Len() != 2 {
			panic("dict item is not KV")
		}
		key, err := e.asInterface(item.Index(0))
		if err != nil {
			return nil, err
		}
		value, err := e.asInterface(item.Index(1))
		if err != nil {
			return nil, err
		}
		result.Set(key, value)
	}
	return result, nil
}

func (e StarlarkValue) structAsInterface(val *StarlarkStruct) (interface{}, error) {
	// TODO accessing privates
	result := orderedmap.NewMap()
	err := val.data.IterateErr(func(k, v interface{}) error {
		value, err := e.asInterface(v.(starlark.Value))
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

func (e StarlarkValue) itearableAsInterface(iterable starlark.Iterable) (interface{}, error) {
	iter := iterable.Iterate()
	defer iter.Done()

	var result []interface{}
	var x starlark.Value
	for iter.Next(&x) {
		elem, err := e.asInterface(x)
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}
	return result, nil
}
