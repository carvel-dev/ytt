// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var (
	YAMLAPI = starlark.StringDict{
		"yaml": &starlarkstruct.Module{
			Name: "yaml",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("yaml.encode", core.ErrWrapper(yamlModule{}.Encode)),
				"decode": starlark.NewBuiltin("yaml.decode", core.ErrWrapper(yamlModule{}.Decode)),
			},
		},
	}
)

type yamlModule struct{}

func (b yamlModule) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	var docSet *yamlmeta.DocumentSet

	switch typedVal := val.(type) {
	case *yamlmeta.DocumentSet:
		docSet = typedVal
	case *yamlmeta.Document:
		// Documents should be part of DocumentSet by the time it makes it here
		panic("Unexpected document")
	default:
		docSet = &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{{Value: typedVal}}}
	}

	valBs, err := docSet.AsBytes()
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(string(valBs)), nil
}

func (b yamlModule) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	valEncoded, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var valDecoded interface{}

	err = yamlmeta.PlainUnmarshal([]byte(valEncoded), &valDecoded)
	if err != nil {
		return starlark.None, err
	}

	return core.NewGoValue(valDecoded).AsStarlarkValue(), nil
}
