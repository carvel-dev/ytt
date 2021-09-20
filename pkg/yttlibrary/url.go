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

// URLValue stores a parsed URL
type URLValue struct {
	url                  *url.URL
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
}

// URLUser stores the user information
type URLUser struct {
	user                 *url.Userinfo
	*core.StarlarkStruct // TODO: keep authorship of the interface by delegating instead of embedding
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

	val, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	typedVal, ok := val.(*orderedmap.Map)
	if !ok {
		return starlark.None, fmt.Errorf("expected argument to be a map, but was %T", val)
	}

	urlVals := url.Values{}

	err = typedVal.IterateErr(func(key, val interface{}) error {
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

	return (&URLValue{parsedURL, nil}).AsStarlarkValue(), nil
}

const urlValueTypeName = "url.value"

// Type reports the fully-qualified name of type of this custom struct.
func (uv *URLValue) Type() string { return "@ytt:" + urlValueTypeName }

func (uv *URLValue) AsStarlarkValue() starlark.Value {
	m := orderedmap.NewMap()
	m.Set("user", uv.User())
	m.Set("without_user", starlark.NewBuiltin(urlValueTypeName+".without_user", core.ErrWrapper(uv.WithoutUser)))
	m.Set("string", starlark.NewBuiltin(urlValueTypeName+".string", core.ErrWrapper(uv.string)))
	m.Set("hostname", starlark.NewBuiltin(urlValueTypeName+".hostname", core.ErrWrapper(uv.Hostname)))
	uv.StarlarkStruct = core.NewStarlarkStruct(m)
	return uv
}

func (uv *URLValue) ConversionHint() string {
	return uv.Type() + " does not automatically encode (hint: use .string())"
}

func (uv *URLValue) Hostname(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(uv.url.Hostname()), nil
}

func (uu *URLUser) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(uu.user.String()), nil
}

func (uv *URLValue) User() starlark.Value {
	if uv.url.User == nil {
		return starlark.None
	}

	uu := &URLUser{uv.url.User, nil}
	m := orderedmap.NewMap()
	m.Set("name", starlark.String(uu.user.Username()))
	m.Set("password", uu.password())
	m.Set("string", starlark.NewBuiltin("string", core.ErrWrapper(uu.string)))
	uu.StarlarkStruct = core.NewStarlarkStruct(m)
	return uu
}

const urlUserTypeName = "url.user"

// Type reports the fully-qualified name of type of this custom struct.
func (uu *URLUser) Type() string { return "@ytt:" + urlUserTypeName }

func (uu *URLUser) ConversionHint() string {
	return uu.Type() + " does not automatically encode (hint: use .string())"
}

func (uu *URLUser) password() starlark.Value {
	passwd, passwdSet := uu.user.Password()
	if !passwdSet {
		return starlark.None
	}
	return starlark.String(passwd)
}

func (uv *URLValue) WithoutUser(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	urlVar := *uv.url
	urlVar.User = nil
	return (&URLValue{&urlVar, nil}).AsStarlarkValue(), nil
}

func (uv *URLValue) string(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no argument")
	}
	return starlark.String(uv.url.String()), nil
}
