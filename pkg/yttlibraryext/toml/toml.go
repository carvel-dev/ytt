// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package toml

import (
	"bytes"
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yttlibrary"
	"github.com/BurntSushi/toml"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

var (
	tomlMod = &starlarkstruct.Module{
		Name: "toml",
		Members: starlark.StringDict{
			"encode": starlark.NewBuiltin("toml.encode", core.ErrWrapper(tomlModule{}.Encode)),
			"decode": starlark.NewBuiltin("toml.decode", core.ErrWrapper(tomlModule{}.Decode)),
		},
	}

	// TOMLAPI contains the definition of the @ytt:toml module
	TOMLAPI = starlark.StringDict{"toml": tomlMod}
)

func init() {
	yttlibrary.RegisterExt(tomlMod)
}

type tomlModule struct{}

// Encode is a core.StarlarkFunc that renders the provided input into a TOML formatted string
func (b tomlModule) Encode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
func (b tomlModule) Decode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
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
