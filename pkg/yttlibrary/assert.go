// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"github.com/k14s/starlark-go/syntax"
	"github.com/vmware-tanzu/carvel-ytt/pkg/orderedmap"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
)

var (
	AssertAPI = starlark.StringDict{
		"assert": &starlarkstruct.Module{
			Name: "assert",
			Members: starlark.StringDict{
				"equals":       starlark.NewBuiltin("assert.equals", core.ErrWrapper(assertModule{}.Equals)),
				"fail":         starlark.NewBuiltin("assert.fail", core.ErrWrapper(assertModule{}.Fail)),
				"try_to":       starlark.NewBuiltin("assert.try_to", core.ErrWrapper(assertModule{}.TryTo)),
				"one_not_null": starlark.NewBuiltin("assert.one_not_null", core.ErrWrapper(assertModule{}.OneNotNull)),
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

// Assertion encapsulates a specific assertion, ready to be checked against any number of values.
type Assertion struct {
	check starlark.Callable
	*core.StarlarkStruct
}

const assertionTypeName = "assert.assertion"

// Type reports the name of this type in the Starlark type system.
func (a *Assertion) Type() string { return "@ytt:" + assertionTypeName }

// CheckFunc returns the function that — given a value — makes this assertion on it.
func (a *Assertion) CheckFunc() starlark.Callable {
	return a.check
}

// NewAssertion creates a struct with one attribute: "check" that contains the assertion defined by "src".
func NewAssertion(funcName, src string, env starlark.StringDict) *Assertion {
	expr, err := syntax.ParseExpr(funcName, src, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", src, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, env)
	if err != nil {
		panic(fmt.Sprintf("Failed to evaluate internal expression (%s) given env=%s", src, env))
	}

	a := &Assertion{check: evalExpr.(*starlark.Function)}
	m := orderedmap.NewMap()
	m.Set("check", a.check)
	a.StarlarkStruct = core.NewStarlarkStruct(m)
	return a
}

// NewAssertOneNotNull produces an assertion object that asserts that a given map has one and only one not null map item.
func NewAssertOneNotNull() *Assertion {
	return NewAssertion(
		"assert.one_not_null",
		`lambda v: True if len([x for x in v if v[x] != None]) == 1 else fail("all values are null") if len([x for x in v if v[x] != None]) == 0 else fail("multiple values are not null")`,
		starlark.StringDict{},
	)
}

// OneNotNull is a core.StarlarkFunc that asserts that a given map has one and only one not null map item.
func (b assertModule) OneNotNull(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return NewAssertOneNotNull(), nil
}
