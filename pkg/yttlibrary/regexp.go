// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"regexp"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	RegexpAPI = starlark.StringDict{
		"regexp": &starlarkstruct.Module{
			Name: "regexp",
			Members: starlark.StringDict{
				"match":   starlark.NewBuiltin("regexp.match", core.ErrWrapper(regexpModule{}.Match)),
				"replace": starlark.NewBuiltin("regexp.replace", core.ErrWrapper(regexpModule{}.Replace)),
			},
		},
	}
)

type regexpModule struct{}

func (b regexpModule) Match(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return starlark.None, fmt.Errorf("expected exactly two arguments")
	}

	pattern, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	target, err := core.NewStarlarkValue(args.Index(1)).AsString()
	if err != nil {
		return starlark.None, err
	}

	matched, err := regexp.MatchString(pattern, target)
	if err != nil {
		return starlark.None, err
	}

	return starlark.Bool(matched), nil
}

func (b regexpModule) Replace(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 3 {
		return starlark.None, fmt.Errorf("expected exactly 3 arguments")
	}

	pattern, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return starlark.None, err
	}

	source, err := core.NewStarlarkValue(args.Index(1)).AsString()
	if err != nil {
		return starlark.None, err
	}

	repl := args.Index(2)
	switch typedRepl := repl.(type) {
	case starlark.Callable:
		return b.replaceLambda(thread, re, source, typedRepl)
	default:
		return b.replaceString(re, source, repl)
	}
}

func (b regexpModule) replaceString(re *regexp.Regexp, source string, repl starlark.Value) (starlark.Value, error) {
	replStr, err := core.NewStarlarkValue(repl).AsString()
	if err != nil {
		return starlark.None, err
	}

	newString := re.ReplaceAllString(source, replStr)

	return starlark.String(newString), nil
}

func (b regexpModule) replaceLambda(thread *starlark.Thread, re *regexp.Regexp, source string, repl starlark.Callable) (starlark.Value, error) {
	var lastErr error
	newString := re.ReplaceAllStringFunc(source, func(match string) string {
		if lastErr != nil {
			// if we have multiple matches but an earlier replace caused an error, we want to return
			// quickly then propagate that error
			return ""
		}

		args := starlark.Tuple{starlark.String(match)}
		var result starlark.Value
		result, lastErr = starlark.Call(thread, repl, args, []starlark.Tuple{})
		if lastErr != nil {
			return ""
		}

		var newString string
		newString, lastErr = core.NewStarlarkValue(result).AsString()
		if lastErr != nil {
			return ""
		}
		return newString
	})

	if lastErr != nil {
		return nil, lastErr
	}

	return starlark.String(newString), nil
}
