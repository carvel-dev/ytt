// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"
	"reflect"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

var (
	API = starlark.StringDict{
		"overlay": &starlarkstruct.Module{
			Name: "overlay",
			Members: starlark.StringDict{
				"apply":   starlark.NewBuiltin("overlay.apply", core.ErrWrapper(overlayModule{}.Apply)),
				"index":   starlark.NewBuiltin("overlay.index", core.ErrWrapper(overlayModule{}.Index)),
				"all":     starlark.NewBuiltin("overlay.all", core.ErrWrapper(overlayModule{}.All)),
				"map_key": overlayModule{}.MapKey(),
				"subset":  starlark.NewBuiltin("overlay.subset", core.ErrWrapper(overlayModule{}.Subset)),

				"and_op": starlark.NewBuiltin("overlay.and_op", core.ErrWrapper(overlayModule{}.AndOp)),
				"or_op":  starlark.NewBuiltin("overlay.or_op", core.ErrWrapper(overlayModule{}.OrOp)),
				"not_op": starlark.NewBuiltin("overlay.not_op", core.ErrWrapper(overlayModule{}.NotOp)),
			},
		},
	}
)

type overlayModule struct{}

func (b overlayModule) Apply(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() == 0 {
		return starlark.None, fmt.Errorf("expected exactly at least argument")
	}

	goValue, err := core.NewStarlarkValue(args).AsGoValue()
	if err != nil {
		return starlark.None, err
	}
	typedVals := goValue.([]interface{})
	var result interface{} = typedVals[0]

	for _, right := range typedVals[1:] {
		var err error
		result, err = Op{Left: result, Right: right, Thread: thread}.Apply() // left is modified
		if err != nil {
			return starlark.None, err
		}
	}

	return yamltemplate.NewStarlarkFragment(result), nil
}

func (b overlayModule) Index(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	expectedIdx64, err := core.NewStarlarkValue(args.Index(0)).AsInt64()
	if err != nil {
		return starlark.None, err
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		idx64, err := core.NewStarlarkValue(args.Index(0)).AsInt64()
		if err != nil {
			return starlark.None, err
		}

		if expectedIdx64 == idx64 {
			return starlark.Bool(true), nil
		}

		return starlark.Bool(false), nil
	}

	return starlark.NewBuiltin("overlay.index_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) All(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 3 {
		return starlark.None, fmt.Errorf("expected exactly 3 arguments")
	}

	return starlark.Bool(true), nil
}

func (b overlayModule) MapKey() *starlark.Builtin {
	return starlark.NewBuiltin("overlay.map_key", core.ErrWrapper(b.mapKey))
}

func (b overlayModule) mapKey(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	keyName, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		oldVal, err := core.NewStarlarkValue(args.Index(1)).AsGoValue()
		if err != nil {
			return starlark.None, err
		}
		newVal, err := core.NewStarlarkValue(args.Index(2)).AsGoValue()
		if err != nil {
			return starlark.None, err
		}

		result, err := b.compareByMapKey(keyName, oldVal, newVal)
		if err != nil {
			return nil, err
		}

		return starlark.Bool(result), nil
	}

	return starlark.NewBuiltin("overlay.map_key_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) compareByMapKey(keyName string, oldVal, newVal interface{}) (bool, error) {
	oldKeyVal, err := b.pullOutMapValue(keyName, oldVal)
	if err != nil {
		return false, err
	}

	newKeyVal, err := b.pullOutMapValue(keyName, newVal)
	if err != nil {
		return false, err
	}

	result, _ := Comparison{}.CompareLeafs(oldKeyVal, newKeyVal)
	return result, nil
}

func (b overlayModule) pullOutMapValue(keyName string, val interface{}) (interface{}, error) {
	typedMap, ok := val.(*yamlmeta.Map)
	if !ok {
		return starlark.None, fmt.Errorf("Expected value to be map, but was %T", val)
	}

	for _, item := range typedMap.Items {
		if reflect.DeepEqual(item.Key, keyName) {
			return item.Value, nil
		}
	}

	return starlark.None, fmt.Errorf("Expected to find mapitem with key '%s', but did not", keyName)
}

func (b overlayModule) Subset(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	expectedArg := args.Index(0)

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		leftVal, err := core.NewStarlarkValue(args.Index(1)).AsGoValue()
		if err != nil {
			return starlark.None, err
		}
		expectedVal, err := core.NewStarlarkValue(expectedArg).AsGoValue()
		if err != nil {
			return starlark.None, err
		}

		actualObj := yamlmeta.NewASTFromInterface(leftVal)
		expectedObj := yamlmeta.NewASTFromInterface(expectedVal)

		if _, ok := actualObj.(*yamlmeta.ArrayItem); ok {
			expectedObj = &yamlmeta.ArrayItem{Value: expectedObj}
		}
		if _, ok := actualObj.(*yamlmeta.Document); ok {
			expectedObj = &yamlmeta.Document{Value: expectedObj}
		}

		result, _ := Comparison{}.Compare(actualObj, expectedObj)
		return starlark.Bool(result), nil
	}

	return starlark.NewBuiltin("overlay.subset_matcher", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) AndOp(
	thread *starlark.Thread, f *starlark.Builtin,
	andArgs starlark.Tuple, andKwargs []starlark.Tuple) (starlark.Value, error) {

	if andArgs.Len() < 1 {
		return starlark.None, fmt.Errorf("expected at least one argument")
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		for _, andArg := range andArgs {
			result, err := starlark.Call(thread, andArg, args, kwargs)
			if err != nil {
				return nil, err
			}
			resultBool, err := core.NewStarlarkValue(result).AsBool()
			if err != nil {
				return nil, err
			}
			if !resultBool {
				return starlark.Bool(false), nil
			}
		}

		return starlark.Bool(true), nil
	}

	return starlark.NewBuiltin("overlay.and_op", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) OrOp(
	thread *starlark.Thread, f *starlark.Builtin,
	orArgs starlark.Tuple, orKwargs []starlark.Tuple) (starlark.Value, error) {

	if orArgs.Len() < 1 {
		return starlark.None, fmt.Errorf("expected at least one argument")
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		for _, orArg := range orArgs {
			result, err := starlark.Call(thread, orArg, args, kwargs)
			if err != nil {
				return nil, err
			}
			resultBool, err := core.NewStarlarkValue(result).AsBool()
			if err != nil {
				return nil, err
			}
			if resultBool {
				return starlark.Bool(true), nil
			}
		}

		return starlark.Bool(false), nil
	}

	return starlark.NewBuiltin("overlay.or_op", core.ErrWrapper(matchFunc)), nil
}

func (b overlayModule) NotOp(
	thread *starlark.Thread, f *starlark.Builtin,
	notArgs starlark.Tuple, notKwargs []starlark.Tuple) (starlark.Value, error) {

	if notArgs.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	matchFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		if args.Len() != 3 {
			return starlark.None, fmt.Errorf("expected exactly 3 arguments")
		}

		result, err := starlark.Call(thread, notArgs[0], args, kwargs)
		if err != nil {
			return nil, err
		}
		resultBool, err := core.NewStarlarkValue(result).AsBool()
		if err != nil {
			return nil, err
		}
		return starlark.Bool(!resultBool), nil
	}

	return starlark.NewBuiltin("overlay.or_op", core.ErrWrapper(matchFunc)), nil
}
