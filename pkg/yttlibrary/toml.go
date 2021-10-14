// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var (
	// TOMLAPI contains the definition of the @ytt:toml module
	TOMLAPI = starlark.StringDict{
		"toml": &starlarkstruct.Module{
			Name: "toml",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("toml.encode", core.ErrWrapper(tomlModule{}.Encode)),
				"decode": starlark.NewBuiltin("toml.decode", core.ErrWrapper(tomlModule{}.Decode)),
			},
		},
	}
)

type tomlModule struct{}

// Encode is a core.StarlarkFunc that renders the provided input into a TOML formatted string
func (b tomlModule) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	allowedKWArgs := map[string]struct{}{
		"indent": {},
	}
	if err := core.CheckArgNames(kwargs, allowedKWArgs); err != nil {
		return starlark.None, err
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	val = orderedmap.Conversion{yamlmeta.NewGoFromAST(val)}.AsUnorderedStringMaps()

	indent, err := core.Int64Arg(kwargs, "indent")
	if err != nil {
		return starlark.None, err
	}

	if indent < 0 || indent > 8 {
		// mitigate https://cwe.mitre.org/data/definitions/409.html
		return starlark.None, fmt.Errorf("indent value must be between 0 and 8")
	}

	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if indent > 0 {
		encoder.Indent = strings.Repeat(" ", int(indent))
	}

	err = encoder.Encode(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(buffer.String()), nil
}

// Decode is a core.StarlarkFunc that parses the provided input from TOML format into dicts, lists, and scalars
func (b tomlModule) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	valEncoded, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var valDecoded interface{}

	err = toml.Unmarshal([]byte(valEncoded), &valDecoded)
	if err != nil {
		return starlark.None, err
	}

	valDecoded = orderedmap.Conversion{valDecoded}.FromUnorderedMaps()

	return core.NewGoValue(valDecoded).AsStarlarkValue(), nil
}
