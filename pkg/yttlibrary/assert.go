// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/starlark-go/syntax"
)

// NewAssertModule constructs a new instance of AssertModule, respecting the "validations" experiment flag.
func NewAssertModule() AssertModule {
	return AssertModule{}
}

// AssertModule contains the definition of the @ytt:assert module.
type AssertModule struct{}

// AsModule produces the corresponding Starlark module definition suitable for use in running a Starlark program.
func (m AssertModule) AsModule() starlark.StringDict {
	members := starlark.StringDict{}
	members["equals"] = starlark.NewBuiltin("assert.equals", core.ErrWrapper(m.Equals))
	members["fail"] = starlark.NewBuiltin("assert.fail", core.ErrWrapper(m.Fail))
	members["try_to"] = starlark.NewBuiltin("assert.try_to", core.ErrWrapper(m.TryTo))
	members["min"] = starlark.NewBuiltin("assert.min", core.ErrWrapper(m.Min))
	members["min_len"] = starlark.NewBuiltin("assert.min_len", core.ErrWrapper(m.MinLen))
	members["max"] = starlark.NewBuiltin("assert.max", core.ErrWrapper(m.Max))
	members["max_len"] = starlark.NewBuiltin("assert.max_len", core.ErrWrapper(m.MaxLen))
	members["not_null"] = starlark.NewBuiltin("assert.not_null", core.ErrWrapper(m.NotNull))
	members["one_not_null"] = starlark.NewBuiltin("assert.one_not_null", core.ErrWrapper(m.OneNotNull))
	members["one_of"] = starlark.NewBuiltin("assert.one_of", core.ErrWrapper(m.OneOf))
	return starlark.StringDict{
		"assert": &starlarkstruct.Module{
			Name:    "assert",
			Members: members,
		},
	}
}

// Equals compares two values for equality
func (m AssertModule) Equals(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 2)
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
func (m AssertModule) Fail(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.None, fmt.Errorf("fail: %s", val)
}

// TryTo is a core.StarlarkFunc that attempts to invoke the passed in starlark.Callable, converting any error into
// an error message.
func (m AssertModule) TryTo(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
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

// Type reports the name of this type in the Starlark type system.
func (a *Assertion) Type() string { return "@ytt:assert.assertion" }

// CheckFunc returns the function that — given a value — makes this assertion on it.
func (a *Assertion) CheckFunc() starlark.Callable {
	return a.check
}

// ConversionHint helps the user get unstuck if they accidentally left an Assertion as a value in a YAML being
// encoded.
func (a *Assertion) ConversionHint() string {
	return a.Type() + " does not encode (did you mean to call check()?)"
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
func NewAssertionFromStarlarkFunc(funcName string, checkFunc core.StarlarkFunc) *Assertion {
	a := &Assertion{check: starlark.NewBuiltin(funcName, checkFunc)}

	m := orderedmap.NewMap()
	m.Set("check", a.check)
	a.StarlarkStruct = core.NewStarlarkStruct(m)

	return a
}

// NewAssertMaxLen produces an Assertion that a given sequence is at most "maximum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMaxLen(maximum starlark.Int) *Assertion {
	return NewAssertionFromSource(
		"assert.max_len",
		`lambda sequence: True if len(sequence) <= maximum else fail ("length = {}".format(len(sequence)))`,
		starlark.StringDict{"maximum": maximum},
	)
}

// MaxLen is a core.StarlarkFunc wrapping NewAssertMaxLen()
func (m AssertModule) MaxLen(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	max, err := starlark.NumberToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("expected value to be an number, but was %s", args[0].Type())
	}
	return NewAssertMaxLen(max), nil
}

// NewAssertMinLen produces an Assertion that a given sequence is at least "minimum" in length.
//
// see also: https://github.com/google/starlark-go/blob/master/doc/spec.md#len
func NewAssertMinLen(minimum starlark.Int) *Assertion {
	return NewAssertionFromSource(
		"assert.min_len",
		`lambda sequence: True if len(sequence) >= minimum else fail ("length = {}".format(len(sequence)))`,
		starlark.StringDict{"minimum": minimum},
	)
}

// MinLen is a core.StarlarkFunc wrapping NewAssertMinLen()
func (m AssertModule) MinLen(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	min, err := starlark.NumberToInt(args[0])
	if err != nil {
		return nil, fmt.Errorf("expected value to be an number, but was %s", args[0].Type())
	}
	return NewAssertMinLen(min), nil
}

// NewAssertMin produces an Assertion that a given value is at least "minimum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMin(min starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.min",
		`lambda val: yaml.decode(yaml.encode(val)) >= yaml.decode(yaml.encode(min)) or fail("value < {}".format(yaml.decode(yaml.encode(min))))`,
		starlark.StringDict{"min": min, "yaml": YAMLAPI["yaml"]},
	)
}

// Min is a core.StarlarkFunc wrapping NewAssertMin()
func (m AssertModule) Min(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	minFunc := NewAssertMin(args[0])
	return minFunc, nil
}

// NewAssertMax produces an Assertion that a given value is at most "maximum".
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#comparisons
func NewAssertMax(max starlark.Value) *Assertion {
	return NewAssertionFromSource(
		"assert.max",
		`lambda val: yaml.decode(yaml.encode(val)) <= yaml.decode(yaml.encode(max)) or fail("value > {}".format(yaml.decode(yaml.encode(max))))`,
		starlark.StringDict{"max": max, "yaml": YAMLAPI["yaml"]},
	)
}

// Max is a core.StarlarkFunc wrapping NewAssertMax()
func (m AssertModule) Max(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want %d", args.Len(), 1)
	}

	maxFunc := NewAssertMax(args[0])
	return maxFunc, nil
}

// NewAssertNotNull produces an Assertion that a given value is not null.
func NewAssertNotNull() *Assertion {
	return NewAssertionFromSource(
		"assert.not_null",
		`lambda value: value != None or fail("value is null")`,
		starlark.StringDict{},
	)
}

// NotNull is a core.StarlarkFunc wrapping NewAssertNotNull()
func (m AssertModule) NotNull(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 1 {
		return starlark.None, fmt.Errorf("got %d arguments, want at most %d", args.Len(), 1)
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
	return NewAssertionFromStarlarkFunc("assert.one_not_null", AssertModule{}.oneNotNullCheck(keys))
}

// OneNotNull is a core.StarlarkFunc wrapping NewAssertOneNotNull()
func (m AssertModule) OneNotNull(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
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
		return nil, fmt.Errorf("expected a sequence of keys, but was a '%s'", args[0].Type())
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
			return nil, fmt.Errorf("check: value must be a map or dict, but was '%s'", val.Type())
		}

		var keysToCheck []starlark.Value

		if keys == nil {
			keysToCheck = append(keysToCheck, dict.Keys()...)
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

		var nulls, notNulls []starlark.Value

		for _, key := range keysToCheck {
			value, found, err := dict.Get(key)
			if err != nil {
				return nil, fmt.Errorf("check: unexpected error while looking up key %s in dict %s", key, dict)
			}
			if !found {
				// allow schema to catch this (see also https://github.com/carvel-dev/ytt/issues/722)
				nulls = append(nulls, key)
			}
			if value == starlark.None {
				nulls = append(nulls, key)
			} else {
				notNulls = append(notNulls, key)
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
			return nil, fmt.Errorf("check: %s are not null", starlark.NewList(notNulls).String())
		}
	}
}

// NewAssertOneOf produces an Assertion that a given value is one of a pre-defined set.
//
// see also:https://github.com/google/starlark-go/blob/master/doc/spec.md#membership-tests
func NewAssertOneOf(enum starlark.Sequence) *Assertion {
	return NewAssertionFromSource(
		"assert.one_of",
		`lambda val: yaml.decode(yaml.encode(val)) in yaml.decode(yaml.encode(enum)) or fail("not one of allowed values")`,
		starlark.StringDict{"enum": enum, "yaml": YAMLAPI["yaml"]},
	)
}

// OneOf is a core.StarlarkFunc wrapping NewAssertOneOf()
func (m AssertModule) OneOf(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() == 0 {
		return starlark.None, fmt.Errorf("got %d arguments, want at least %d", args.Len(), 1)
	}

	enum := args[0]
	if args.Len() > 1 {
		enum = args
	}

	seq, ok := enum.(starlark.Sequence)
	if !ok {
		return nil, fmt.Errorf("expected a sequence, but was a '%s'", enum.Type())
	}

	maxFunc := NewAssertOneOf(seq)
	return maxFunc, nil
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
