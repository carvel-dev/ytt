// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"reflect"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var (
	AssertAPI = starlark.StringDict{
		"assert": &starlarkstruct.Module{
			Name: "assert",
			Members: starlark.StringDict{
				"equals": starlark.NewBuiltin("assert.equals", core.ErrWrapper(assertModule{}.Equals)),
				"fail":   starlark.NewBuiltin("assert.fail", core.ErrWrapper(assertModule{}.Fail)),
				"try_to": starlark.NewBuiltin("assert.try_to", core.ErrWrapper(assertModule{}.TryTo)),
			},
		},
	}
)

type assertModule struct{}

func (b assertModule) Equals(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return starlark.None, fmt.Errorf("expected two arguments")
	}

	expected := args.Index(0)
	actual := args.Index(1)
	if _, notOk := expected.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", expected)
	}
	if _, notOk := actual.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", actual)
	}

	expectedType := expected.Type()
	actualType := actual.Type()
	if expectedType != actualType {
		return starlark.None, fmt.Errorf("arguments are of different types. expected %s, but got %s", expectedType, actualType)
	}

	if expectedType == "yamlfragment" {
		expectedStarlarkValue, err := core.NewStarlarkValue(expected).AsGoValue()
		if err != nil {
			return starlark.None, err
		}
		actualStarlarkValue, err := core.NewStarlarkValue(actual).AsGoValue()
		if err != nil {
			return starlark.None, err
		}
		document := yamlmeta.Document{Value: expectedStarlarkValue}
		expectedYaml, err := document.AsYAMLBytes()
		if err != nil {
			return starlark.None, err
		}
		document = yamlmeta.Document{Value: actualStarlarkValue}
		actualYaml, err := document.AsYAMLBytes()
		if err != nil {
			return starlark.None, err
		}
		if string(expectedYaml) != string(actualYaml) {
			//errorMessage := string(actualYaml) + "\nis not equal to the expected yaml value\n" + string(expectedYaml)
			return starlark.None, fmt.Errorf("yamlfragments are not equal:\n"+
				"Expected:\n%s---\nActual:\n%s", string(expectedYaml), string(actualYaml))
		}
	} else {
		if !reflect.DeepEqual(expected, actual) {
			return starlark.None, fmt.Errorf("%s is not equal to the expected value %s", actual.String(), expected.String())
		}
	}

	return starlark.None, nil
}

func (b assertModule) Fail(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.None, fmt.Errorf("fail: %s", val)
}

func (b assertModule) TryTo(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	lambda := args.Index(0)
	if _, ok := lambda.(starlark.Callable); !ok {
		return starlark.None, fmt.Errorf("expected argument to be a function, but was %T", lambda)
	}

	retVal, err := starlark.Call(thread, lambda, nil, nil)
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
	}
	return starlark.Tuple{retVal, starlark.None}, nil
}
