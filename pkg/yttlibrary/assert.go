// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/starlark-go/syntax"
	"github.com/vmware-tanzu/carvel-ytt/pkg/experiments"
	"github.com/vmware-tanzu/carvel-ytt/pkg/orderedmap"
	"github.com/vmware-tanzu/carvel-ytt/pkg/spell"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
)

// NewAssertModule constructs a new instance of AssertModule, respecting the "validations" experiment flag.
func NewAssertModule() AssertModule {
	return AssertModule{}
}

// AssertModule contains the definition of the @ytt:assert module.
type AssertModule struct{}

// AsModule produces the corresponding Starlark module definition suitable for use in running a Starlark program.
func (m AssertModule) AsModule() starlark.StringDict {
	members := starlark.StringDict{
		"equals": starlark.NewBuiltin("assert.equals", core.ErrWrapper(m.Equals)),
		"fail":   starlark.NewBuiltin("assert.fail", core.ErrWrapper(m.Fail)),
		"try_to": starlark.NewBuiltin("assert.try_to", core.ErrWrapper(m.TryTo)),
	}
	if experiments.IsValidationsEnabled() {
		members = starlark.StringDict{
			"equals":       starlark.NewBuiltin("assert.equals", core.ErrWrapper(m.Equals)),
			"fail":         starlark.NewBuiltin("assert.fail", core.ErrWrapper(m.Fail)),
			"try_to":       starlark.NewBuiltin("assert.try_to", core.ErrWrapper(m.TryTo)),
			"min":          starlark.NewBuiltin("assert.min", core.ErrWrapper(m.Min)),
			"min_len":      starlark.NewBuiltin("assert.min_len", core.ErrWrapper(m.MinLen)),
			"max":          starlark.NewBuiltin("assert.max", core.ErrWrapper(m.Max)),
			"max_len":      starlark.NewBuiltin("assert.max_len", core.ErrWrapper(m.MaxLen)),
			"not_null":     starlark.NewBuiltin("assert.not_null", core.ErrWrapper(m.NotNull)),
			"one_not_null": starlark.NewBuiltin("assert.one_not_null", core.ErrWrapper(m.OneNotNull)),
		}
	}
	return starlark.StringDict{
		"assert": &starlarkstruct.Module{
			Name:    "assert",
			Members: members,
		},
	}
}

// Equals compares two values for equality
func (m AssertModule) Equals(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

	expectedString, err := m.asString(expected)
	if err != nil {
		return starlark.None, err
	}

	actualString, err := m.asString(actual)
	if err != nil {
		return starlark.None, err
	}

	if expectedString != actualString {
		return starlark.None, fmt.Errorf("Not equal:\n"+
			"(expected type: %s)\n%s\n\n(was type: %s)\n%s", expected.Type(), expectedString, actual.Type(), actualString)
	}

	return starlark.None, nil
}

func (m AssertModule) asString(value starlark.Value) (string, error) {
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

// Fail is a core.StarlarkFunc that forces a Starlark failure.
func (m AssertModule) Fail(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.None, fmt.Errorf("fail: %s", val)
}

// TryTo is a core.StarlarkFunc that attempts to invoke the passed in starlark.Callable, converting any error into
// an error message.
func (m AssertModule) TryTo(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

// Assertion encapsulates a rule (a predicate) that can be accessed in a Starlark expression (via the "check" attribute)
// or in Go (via CheckFunc()).
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

// NewAssertionFromSource creates an Assertion whose "check" attribute is the lambda expression defined in "checkSrc".
func NewAssertionFromSource(funcName, checkSrc string, env starlark.StringDict) *Assertion {
	expr, err := syntax.ParseExpr(funcName, checkSrc, syntax.BlockScanner)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse internal expression (%s) :%s", checkSrc, err))
	}
	thread := &starlark.Thread{Name: "ytt-internal"}

	evalExpr, err := starlark.EvalExpr(thread, expr, env)
	if err != nil {
		panic(fmt.Sprintf("Failed to evaluate internal expression (%s) given env=%s", checkSrc, env))
	}

	a := &Assertion{check: evalExpr.(*starlark.Function)}
	m := orderedmap.NewMap()
	m.Set("check", a.check)
	a.StarlarkStruct = core.NewStarlarkStruct(m)
	return a
}

// NewAssertionFromStarlarkFunc creates an Assertion whose "check" attribute is "checkFunc".
func NewAssertionFromStarlarkFunc(checkFunc core.StarlarkFunc) *Assertion {
	a := &Assertion{check: starlark.NewBuiltin("check", checkFunc)}

	m := orderedmap.NewMap()
	m.Set("check", a.check)
	a.StarlarkStruct = core.NewStarlarkStruct(m)

	return a
}

// NewAssertMaxLen produces an Assertion that a given sequence is at most "maximum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMaxLen(maximum starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.max_len",
		`lambda sequence: True if len(sequence) <= maximum else fail ("length of {} is more than {}".format(len(sequence), maximum))`,
		starlark.StringDict{"maximum": maximum},
	)
}

// MaxLen is a core.StarlarkFunc wrapping NewAssertMaxLen()
func (m AssertModule) MaxLen(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	max := args[0]
	if !(max.Type() == "int" || max.Type() == "float") {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", max.Type())
	}
	maxLenFunc := NewAssertMaxLen(args[0])
	return maxLenFunc, nil
}

// NewAssertMinLen produces an Assertion that a given sequence is at least "minimum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMinLen(minimum starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.min_len",
		`lambda sequence: True if len(sequence) >= minimum else fail ("length of {} is less than {}".format(len(sequence), minimum))`,
		starlark.StringDict{"minimum": minimum},
	)
}

// MinLen is a core.StarlarkFunc wrapping NewAssertMinLen()
func (m AssertModule) MinLen(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	min := args[0]
	if !(min.Type() == "int" || min.Type() == "float") {
		return starlark.None, fmt.Errorf("expected value to be an number, but was %s", min.Type())
	}
	minLengthFunc := NewAssertMinLen(min)
	return minLengthFunc, nil
}

// NewAssertMin produces an Assertion that a given value is at least "minimum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMin(minimum starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.min",
		`lambda value: True if yaml.decode(yaml.encode(value)) >= yaml.decode(yaml.encode(minimum)) else fail("{} is less than {}".format(value, minimum))`,
		starlark.StringDict{"minimum": minimum, "yaml": YAMLAPI["yaml"]},
	)
}

// Min is a core.StarlarkFunc wrapping NewAssertMin()
func (m AssertModule) Min(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	minFunc := NewAssertMin(args[0])
	return minFunc, nil
}

// NewAssertMax produces an Assertion that a given value is at most "maximum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMax(maximum starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.max",
		`lambda value: True if yaml.decode(yaml.encode(value)) <= yaml.decode(yaml.encode(maximum)) else fail("{} is more than {}".format(value, maximum))`,
		starlark.StringDict{"maximum": maximum, "yaml": YAMLAPI["yaml"]},
	)
}

// Max is a core.StarlarkFunc wrapping NewAssertMax()
func (m AssertModule) Max(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	maxFunc := NewAssertMax(args[0])
	return maxFunc, nil
}

// NewAssertNotNull produces an Assertion that a given value is not null.
func NewAssertNotNull() *Assertion {
	return NewAssertionFromSource(
		"assert.not_null",
		`lambda value: fail("value is null") if value == None else True`,
		starlark.StringDict{},
	)
}

// NotNull is a core.StarlarkFunc wrapping NewAssertNotNull()
func (m AssertModule) NotNull(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 1 {
		return starlark.None, fmt.Errorf("expected no more than one argument")
	}

	result := NewAssertNotNull()
	if args.Len() == 0 {
		return result, nil
	}

	// support shorthand syntax: assert.not_null(value)
	return starlark.Call(thread, result.CheckFunc(), args, []starlark.Tuple{})
}

// NewAssertOneNotNull produces an Assertion that a given value is a map having exactly one item with a non-null value.
func NewAssertOneNotNull(keys starlark.Sequence) *Assertion {
	return NewAssertionFromStarlarkFunc(AssertModule{}.oneNotNullCheck(keys))
}

// OneNotNull is a core.StarlarkFunc wrapping NewAssertOneNotNull()
func (m AssertModule) OneNotNull(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	if args.Len() == 0 {
		return NewAssertOneNotNull(nil), nil
	}
	if b, ok := args[0].(starlark.Bool); ok {
		if b.Truth() {
			return NewAssertOneNotNull(nil), nil
		}
		return nil, fmt.Errorf("one_not_null() cannot be False")
	}

	seq, ok := args[0].(starlark.Sequence)
	if !ok {
		return nil, fmt.Errorf("expected a sequence of keys, but was a \"%s\"", args[0].Type())
	}
	return NewAssertOneNotNull(seq), nil
}

func (m AssertModule) oneNotNullCheck(keys starlark.Sequence) core.StarlarkFunc {
	return func(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 1 {
			return starlark.None, fmt.Errorf("check: got %d arguments, want %d", args.Len(), 1)
		}
		val, err := m.yamlEncodeDecode(args[0])
		if err != nil {
			return nil, err
		}
		dict, ok := val.(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("check: value must be a map or dict, but was \"%s\"", val.Type())
		}

		var keysToCheck []starlark.Value

		if keys == nil {
			for _, key := range dict.Keys() {
				keysToCheck = append(keysToCheck, key)
			}
		} else {
			var key starlark.Value
			keys := keys.Iterate()
			for keys.Next(&key) {
				v, err := m.yamlEncodeDecode(key)
				if err != nil {
					return nil, err
				}
				keysToCheck = append(keysToCheck, v)
			}
		}

		var nulls, notNulls []string

		for _, key := range keysToCheck {
			value, found, err := dict.Get(key)
			if err != nil {
				return nil, fmt.Errorf("check: unexpected error while looking up key %s in dict %s", key, dict)
			}
			if !found {
				hint := m.maybeSuggestKey(key, *dict)
				return nil, fmt.Errorf("check: %s is not present in value%s", key, hint)
			}
			if value == starlark.None {
				nulls = append(nulls, key.String())
			} else {
				notNulls = append(notNulls, key.String())
			}
		}

		switch len(notNulls) {
		case 0:
			if len(dict.Keys()) == 0 {
				return nil, fmt.Errorf("check: value is empty")
			}
			return nil, fmt.Errorf("check: all values are null")
		case 1:
			return starlark.True, nil
		default:
			return nil, fmt.Errorf("check: multiple values are not null %s", notNulls)
		}
	}
}

func (m AssertModule) maybeSuggestKey(given starlark.Value, expected starlark.Dict) string {
	var keysAsStrings []string
	for _, k := range expected.Keys() {
		keysAsStrings = append(keysAsStrings, k.String())
	}
	mostSimilarKey := spell.Nearest(given.String(), keysAsStrings)
	var hint string
	if mostSimilarKey != "" {
		hint = fmt.Sprintf(" (did you mean %s?)", mostSimilarKey)
	}
	return hint
}

func (m AssertModule) yamlEncodeDecode(val starlark.Value) (starlark.Value, error) {
	yaml := yamlModule{}
	value, err := core.NewStarlarkValue(val).AsGoValue()
	if err != nil {
		return nil, err
	}
	encode, err := yaml.Encode(value)
	if err != nil {
		return nil, err
	}
	val, err = yaml.Decode(encode)
	if err != nil {
		return nil, err
	}
	return val, nil
}
