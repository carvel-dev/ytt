// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	URLAPI = starlark.StringDict{
		"url": &starlarkstruct.Module{
			Name: "url",
			Members: starlark.StringDict{
				"path_segment_encode": starlark.NewBuiltin("url.path_segment_encode", core.ErrWrapper(urlModule{}.PathSegmentEncode)),
				"path_segment_decode": starlark.NewBuiltin("url.path_segment_decode", core.ErrWrapper(urlModule{}.PathSegmentDecode)),

				"query_param_value_encode": starlark.NewBuiltin("url.query_param_value_encode", core.ErrWrapper(urlModule{}.QueryParamValueEncode)),
				"query_param_value_decode": starlark.NewBuiltin("url.query_param_value_decode", core.ErrWrapper(urlModule{}.QueryParamValueDecode)),

				"query_params_encode": starlark.NewBuiltin("url.query_params_encode", core.ErrWrapper(urlModule{}.QueryParamsEncode)),
				"query_params_decode": starlark.NewBuiltin("url.query_params_decode", core.ErrWrapper(urlModule{}.QueryParamsDecode)),

				"parse": starlark.NewBuiltin("url.parse", core.ErrWrapper(urlModule{}.ParseURL)),
			},
		},
	}
)

type urlModule struct{}

// urlValue stores a parsed URL
type urlValue struct {
	url *url.URL
	*core.StarlarkStruct
}

// urlUser stores the user information
type urlUser struct {
	user *url.Userinfo
	*core.StarlarkStruct
}

func (b urlModule) PathSegmentEncode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(url.PathEscape(val)), nil
}

func (b urlModule) PathSegmentDecode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	val, err = url.PathUnescape(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(val), nil
}

func (b urlModule) QueryParamValueEncode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(url.QueryEscape(val)), nil
}

func (b urlModule) QueryParamValueDecode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	val, err = url.QueryUnescape(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(val), nil
}

func (b urlModule) QueryParamsEncode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val := core.NewStarlarkValue(args.Index(0)).AsGoValue()

	typedVal, ok := val.(*orderedmap.Map)
	if !ok {
		return starlark.None, fmt.Errorf("expected argument to be a map, but was %T", val)
	}

	urlVals := url.Values{}

	err := typedVal.IterateErr(func(key, val interface{}) error {
		keyStr, ok := key.(string)
		if !ok {
			return fmt.Errorf("expected map key to be string, but was %T", key)
		}

		valArray, ok := val.([]interface{})
		if !ok {
			return fmt.Errorf("expected map value to be array, but was %T", val)
		}

		if len(valArray) == 0 {
			urlVals[keyStr] = []string{}
		} else {
			for _, valItem := range valArray {
				valItemStr, ok := valItem.(string)
				if !ok {
					return fmt.Errorf("expected array value to be string, but was %T", valItem)
				}
				urlVals[keyStr] = append(urlVals[keyStr], valItemStr)
			}
		}

		return nil
	})
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(urlVals.Encode()), nil
}

func (b urlModule) QueryParamsDecode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	encodedVal, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	urlVals, err := url.ParseQuery(encodedVal)
	if err != nil {
		return starlark.None, err
	}

	result := orderedmap.NewMap()

	for _, key := range b.sortedKeys(urlVals) {
		val := []interface{}{}
		for _, v := range urlVals[key] {
			val = append(val, v)
		}
		result.Set(key, val)
	}

	return core.NewGoValue(result).AsStarlarkValue(), nil
}

func (b urlModule) sortedKeys(vals url.Values) []string {
	var result []string
	for k := range vals {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func (b urlModule) ParseURL(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	urlStr, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return starlark.None, err
	}

	return (&urlValue{parsedURL, nil}).AsStarlarkValue(), nil
}

func (uv *urlValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("user", uv.user())
	m.Set("without_user", starlark.NewBuiltin("url.without_user", core.ErrWrapper(uv.withoutUser)))
	m.Set("string", starlark.NewBuiltin("url.string", core.ErrWrapper(uv.string)))
	uv.StarlarkStruct = core.NewStarlarkStruct(m)
	return uv
}

func (uv *urlValue) ConversionHint() string {
	return "urlValue: cannot coerce to a string; use .string()"
}

func (uu *urlUser) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(uu.user.String()), nil
}

func (uv *urlValue) user() starlark.Value {
	if uv.url.User == nil {
		return starlark.None
	}

	uu := &urlUser{uv.url.User, nil}
	m := orderedmap.NewMap()
	m.Set("name", starlark.String(uu.user.Username()))
	m.Set("password", uu.password())
	m.Set("string", starlark.NewBuiltin("string", core.ErrWrapper(uu.string)))
	uu.StarlarkStruct = core.NewStarlarkStruct(m)
	return uu
}

func (uu *urlUser) ConversionHint() string {
	return "urlUser: cannot coerce to a string; use .string()"
}

func (uu *urlUser) password() starlark.Value {
	passwd, passwdSet := uu.user.Password()
	if !passwdSet {
		return starlark.None
	}
	return starlark.String(passwd)
}

func (uv *urlValue) withoutUser(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	urlVar := *uv.url
	urlVar.User = nil
	return (&urlValue{&urlVar, nil}).AsStarlarkValue(), nil
}

func (uv *urlValue) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(uv.url.String()), nil
}
