// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"github.com/k14s/starlark-go/syntax"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
)

var (
	AssertAPI = starlark.StringDict{
		"assert": &starlarkstruct.Module{
			Name: "assert",
			Members: starlark.StringDict{
				"equals":   starlark.NewBuiltin("assert.equals", core.ErrWrapper(assertModule{}.Equals)),
				"fail":     starlark.NewBuiltin("assert.fail", core.ErrWrapper(assertModule{}.Fail)),
				"try_to":   starlark.NewBuiltin("assert.try_to", core.ErrWrapper(assertModule{}.TryTo)),
				"min":      starlark.NewBuiltin("assert.min", core.ErrWrapper(assertModule{}.Min)),
				"min_len":  starlark.NewBuiltin("assert.min_len", core.ErrWrapper(assertModule{}.MinLength)),
				"max":      starlark.NewBuiltin("assert.max", core.ErrWrapper(assertModule{}.Max)),
				"max_len":  starlark.NewBuiltin("assert.max_len", core.ErrWrapper(assertModule{}.MaxLength)),
				"not_null": starlark.NewBuiltin("assert.not_null", core.ErrWrapper(assertModule{}.NotNull)),
			},
		},
	}
)

type assertModule struct{}

// Equals compares two values for equality
func (b assertModule) Equals(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return starlark.None, fmt.Errorf("expected two arguments")
	}

	expected := args.Index(0)
	if _, notOk := expected.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", expected)
	}

	actual := args.Index(1)
	if _, notOk := actual.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", actual)
	}

	expectedString, err := b.asString(expected)
	if err != nil {
		return starlark.None, err
	}

	actualString, err := b.asString(actual)
	if err != nil {
		return starlark.None, err
	}

	if expectedString != actualString {
		return starlark.None, fmt.Errorf("Not equal:\n"+
			"(expected type: %s)\n%s\n\n(was type: %s)\n%s", expected.Type(), expectedString, actual.Type(), actualString)
	}

	return starlark.None, nil
}

func (b assertModule) asString(value starlark.Value) (string, error) {
	starlarkValue, err := core.NewStarlarkValue(value).AsGoValue()
	if err != nil {
		return "", err
	}
	yamlString, err := yamlModule{}.Encode(starlarkValue)
	if err != nil {
		return "", err
	}
	return yamlString, nil
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

func newAssertFunc(funcName, src string, env starlark.StringDict) *starlark.Function {
	expr, err := syntax.ParseExpr(funcName, src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, env)
	if err != nil {
		panic(fmt.Sprintf("Failed to evaluate internal expression (%s) given env=%s", src, env))
	}
	return evalExpr.(*starlark.Function)
}

// NewAssertMaxLen produces a higher-order Starlark function that asserts that a given sequence is at most
// "maximum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMaxLen(maximum starlark.Value) *starlark.Function {
	return newAssertFunc(
		"assert.max_len",
		`lambda sequence: fail("length of {} is more than {}".format(len(sequence), maximum)) if len(sequence) > maximum else None`,
		starlark.StringDict{"maximum": maximum},
	)
}

// MaxLength is a core.StarlarkFunc that asserts that a given sequence is at most a given maximum length.
func (b assertModule) MaxLength(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	max := args[0]
	if !(max.Type() == "int" || max.Type() == "float") {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", max.Type())
	}
	maxLenFunc := NewAssertMaxLen(args[0])
	if len(args) == 1 {
		return maxLenFunc, nil
	}

	result, err := starlark.Call(thread, maxLenFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// NewAssertMinLen produces a higher-order Starlark function that asserts that a given sequence is at least
// "minimum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMinLen(minimum starlark.Value) *starlark.Function {
	return newAssertFunc(
		"assert.min_len",
		`lambda sequence: fail("length of {} is less than {}".format(len(sequence), minimum)) if len(sequence) < minimum else None`,
		starlark.StringDict{"minimum": minimum},
	)
}

// MinLength is a core.StarlarkFunc that asserts that a given sequence is at least a given minimum length.
func (b assertModule) MinLength(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	min := args[0]
	if !(min.Type() == "int" || min.Type() == "float") {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", min.Type())
	}
	minLengthFunc := NewAssertMinLen(min)
	if len(args) == 1 {
		return minLengthFunc, nil
	}

	result, err := starlark.Call(thread, minLengthFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// NewAssertMin produces a higher-order Starlark function that asserts that a given value is at least "minimum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMin(minimum starlark.Value) *starlark.Function {
	return newAssertFunc(
		"assert.min",
		`lambda value: fail("{} is less than {}".format(value, minimum)) if yaml.decode(yaml.encode(value)) < yaml.decode(yaml.encode(minimum)) else None`,
		starlark.StringDict{"minimum": minimum, "yaml": YAMLAPI["yaml"]},
	)
}

// Min is a core.StarlarkFunc that asserts that a given value is at least a given minimum.
func (b assertModule) Min(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	minFunc := NewAssertMin(args[0])
	if len(args) == 1 {
		return minFunc, nil
	}

	result, err := starlark.Call(thread, minFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// NewAssertMax produces a higher-order Starlark function that asserts that a given value is less than or equal to "maximum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMax(maximum starlark.Value) *starlark.Function {
	return newAssertFunc(
		"assert.max",
		`lambda value: fail("{} is more than {}".format(value, maximum)) if yaml.decode(yaml.encode(value)) > yaml.decode(yaml.encode(maximum)) else None`,
		starlark.StringDict{"maximum": maximum, "yaml": YAMLAPI["yaml"]},
	)
}

// Max is a core.StarlarkFunc that asserts that a given value is less than or equal to a given maximum.
func (b assertModule) Max(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	maxFunc := NewAssertMax(args[0])
	if len(args) == 1 {
		return maxFunc, nil
	}

	result, err := starlark.Call(thread, maxFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

func NewAssertNotNull() *starlark.Function {
	return newAssertFunc(
		"assert.not_null",
		`lambda value: fail("value is null") if value == None else None`,
		starlark.StringDict{},
	)
}

func (b assertModule) NotNull(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	result, err := starlark.Call(thread, NewAssertNotNull(), args, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}
