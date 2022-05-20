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

// assertMaximumLength produces a higher-order Starlark function that asserts that a given sequence is at most
// "maximum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func assertMaximumLength(maximum int) (*starlark.Function, error) {
	src := `lambda sequence: fail("length of {} is more than {}".format(len(sequence), maximum)) if len(sequence) > maximum else None`
	expr, err := syntax.ParseExpr("@ytt:assert.max_len()", src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, starlark.StringDict{"maximum": core.NewGoValue(maximum).AsStarlarkValue()})
	if err != nil {
		return nil, fmt.Errorf("Failed to invoke @ytt:assert.max_len(%v) :%s", maximum, err)
	}
	return evalExpr.(*starlark.Function), nil
}

// NewAssertMaxLength produces a higher-order Starlark function that asserts that a given sequence is at most "maximum"
// in length.
func NewAssertMaxLength(maximum int) *starlark.Function {
	maxLengthFunc, err := assertMaximumLength(maximum)
	if err != nil {
		// TODO: consider whether to return "err" instead of panicing
		panic(fmt.Sprintf("failed to build assert.maximum(): %s", err))
	}
	return maxLengthFunc
}

// MaxLength is a core.StarlarkFunc that asserts that a given sequence is at most a given maximum length.
func (b assertModule) MaxLength(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	val := args[0]
	v, err := starlark.NumberToInt(val)
	if err != nil {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", val.Type())
	}
	num, _ := v.Int64()
	intNum := int(num)
	maxLengthFunc, err := assertMaximumLength(intNum)
	if err != nil {
		return starlark.None, err
	}
	if len(args) == 1 {
		return maxLengthFunc, nil
	}

	result, err := starlark.Call(thread, maxLengthFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// assertMinimumLength produces a higher-order Starlark function that asserts that a given sequence is at least
// "minimum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func assertMinimumLength(minimum int) (*starlark.Function, error) {
	src := `lambda sequence: fail("length of {} is less than {}".format(len(sequence), minimum)) if len(sequence) < minimum else None`
	expr, err := syntax.ParseExpr("@ytt:assert.min_len()", src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, starlark.StringDict{"minimum": core.NewGoValue(minimum).AsStarlarkValue()})
	if err != nil {
		return nil, fmt.Errorf("Failed to invoke @ytt:assert.min_len(%v) :%s", minimum, err)
	}
	return evalExpr.(*starlark.Function), nil
}

// NewAssertMinLength produces a higher-order Starlark function that asserts that a given sequence is at least "minimum"
// in length.
func NewAssertMinLength(minimum int) *starlark.Function {
	minLengthFunc, err := assertMinimumLength(minimum)
	if err != nil {
		// TODO: consider whether to return "err" instead of panicing
		// - minimum is technically supplied by the user
		// - under what conditions does assertMinimum() produce an error?
		// - do any of those conditions occur *because* of the user input?
		panic(fmt.Sprintf("failed to build assert.minimum(): %s", err))
	}
	return minLengthFunc
}

// MinLength is a core.StarlarkFunc that asserts that a given sequence is at least a given minimum length.
func (b assertModule) MinLength(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	val := args[0]
	v, err := starlark.NumberToInt(val)
	if err != nil {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", val.Type())
	}
	num, _ := v.Int64()
	intNum := int(num)
	minLengthFunc, err := assertMinimumLength(intNum)
	if err != nil {
		return starlark.None, err
	}
	if len(args) == 1 {
		return minLengthFunc, nil
	}

	result, err := starlark.Call(thread, minLengthFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// assertMinimum produces a higher-order Starlark function that asserts that a given value is at least "minimum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func assertMinimum(minimum starlark.Value) (*starlark.Function, error) {
	src := `lambda value: fail("{} is less than {}".format(value, minimum)) if value < minimum else None`
	expr, err := syntax.ParseExpr("@ytt:assert.min()", src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, starlark.StringDict{"minimum": minimum})
	if err != nil {
		return nil, fmt.Errorf("Failed to invoke @ytt:assert.min(%v) :%s", minimum, err)
	}
	return evalExpr.(*starlark.Function), nil
}

// NewAssertMin produces a higher-order Starlark function that asserts that a given value is at least "minimum"
func NewAssertMin(minimum starlark.Value) *starlark.Function {
	minimumFunc, err := assertMinimum(minimum)
	if err != nil {
		// TODO: consider whether to return "err" instead of panicking
		// - minimum is technically supplied by the user
		// - under what conditions does assertMinimum() produce an error?
		// - do any of those conditions occur *because* of the user input?
		panic(fmt.Sprintf("failed to build assert.minimum(): %s", err))
	}
	return minimumFunc
}

// Min is a core.StarlarkFunc that asserts that a given value is at least a given minimum.
func (b assertModule) Min(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	minFunc, err := assertMinimum(args[0])
	if err != nil {
		return starlark.None, err
	}
	if len(args) == 1 {
		return minFunc, nil
	}

	result, err := starlark.Call(thread, minFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

// assertMaximum produces a higher-order Starlark function that asserts that a given value is less than or equal to "maximum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func assertMaximum(maximum starlark.Value) (*starlark.Function, error) {
	src := `lambda value: fail("{} is more than {}".format(value, maximum)) if value > maximum else None`
	expr, err := syntax.ParseExpr("@ytt:assert.max()", src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, starlark.StringDict{"maximum": maximum})
	if err != nil {
		return nil, fmt.Errorf("Failed to invoke @ytt:assert.max(%v) :%s", maximum, err)
	}
	return evalExpr.(*starlark.Function), nil
}

// NewAssertMax produces a higher-order Starlark function that asserts that a given value is less than or equal to "maximum"
func NewAssertMax(maximum starlark.Value) *starlark.Function {
	maximumFunc, err := assertMaximum(maximum)
	if err != nil {
		// TODO: consider whether to return "err" instead of panicking
		// - maximum is technically supplied by the user
		// - under what conditions does assertMaximum() produce an error?
		// - do any of those conditions occur *because* of the user input?
		panic(fmt.Sprintf("failed to build assert.max(): %s", err))
	}
	return maximumFunc
}

// Max is a core.StarlarkFunc that asserts that a given value is less than or equal to a given maximum.
func (b assertModule) Max(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.None, fmt.Errorf("expected at least one argument.")
	}
	if len(args) > 2 {
		return starlark.None, fmt.Errorf("expected at no more than two arguments.")
	}

	maxFunc, err := assertMaximum(args[0])
	if err != nil {
		return starlark.None, err
	}
	if len(args) == 1 {
		return maxFunc, nil
	}

	result, err := starlark.Call(thread, maxFunc, starlark.Tuple{args[1]}, []starlark.Tuple{})
	if err != nil {
		return starlark.None, err
	}
	return result, nil
}

func newNotNullStarlarkFunc() core.StarlarkFunc {
	return func(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 1 {
			return starlark.None, fmt.Errorf("expected exactly one argument")
		}
		_, ok := args[0].(starlark.NoneType)
		if ok {
			return starlark.None, fmt.Errorf("value was null")
		}

		return starlark.None, nil
	}
}
func NewAssertNotNull() starlark.Callable {
	return starlark.NewBuiltin("assert.not_null", core.ErrWrapper(newNotNullStarlarkFunc()))
}
func (b assertModule) NotNull(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}
