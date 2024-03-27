// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

var (
	// YAMLAPI contains the definition of the @ytt:yaml module
	YAMLAPI = starlark.StringDict{
		"yaml": &starlarkstruct.Module{
			Name: "yaml",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("yaml.encode", core.ErrWrapper(yamlModule{}.starlarkEncode)),
				"decode": starlark.NewBuiltin("yaml.decode", core.ErrWrapper(yamlModule{}.starlarkDecode)),
			},
		},
	}
)

type yamlModule struct{}

// starlarkEncode adapts a call from Starlark to yamlModule.Encode()
func (b yamlModule) starlarkEncode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	encoded, err := b.Encode(val)
	if err != nil {
		return starlark.None, nil
	}

	return starlark.String(encoded), nil
}

// Encode renders the provided input as a YAML-formatted string
func (b yamlModule) Encode(goValue interface{}) (string, error) {
	var docSet *yamlmeta.DocumentSet

	switch typedVal := goValue.(type) {
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
		return "", err
	}

	return string(valBs), nil
}

// Decode is a core.StarlarkFunc that parses the provided input from YAML format into dicts, lists, and scalars
func (b yamlModule) starlarkDecode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	valEncoded, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	value, err := b.Decode(valEncoded)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (b yamlModule) Decode(yamlString string) (starlark.Value, error) {
	var decoded interface{}

	err := yamlmeta.PlainUnmarshal([]byte(yamlString), &decoded)
	if err != nil {
		return starlark.None, err
	}

	value := core.NewGoValue(decoded).AsStarlarkValue()
	return value, nil
}
