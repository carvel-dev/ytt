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
			},
		},
	}
)

type urlModule struct{}

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
	for k, _ := range vals {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
