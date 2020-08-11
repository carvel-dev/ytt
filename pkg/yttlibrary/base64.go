// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"encoding/base64"
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	Base64API = starlark.StringDict{
		"base64": &starlarkstruct.Module{
			Name: "base64",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("base64.encode", core.ErrWrapper(base64Module{}.Encode)),
				"decode": starlark.NewBuiltin("base64.decode", core.ErrWrapper(base64Module{}.Decode)),
			},
		},
	}
)

type base64Module struct{}

func (b base64Module) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	encoding, err := b.buildEncoding(kwargs)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(encoding.EncodeToString([]byte(val))), nil
}

func (b base64Module) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	encoding, err := b.buildEncoding(kwargs)
	if err != nil {
		return starlark.None, err
	}

	valDecoded, err := encoding.DecodeString(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(string(valDecoded)), nil
}

func (b base64Module) buildEncoding(kwargs []starlark.Tuple) (*base64.Encoding, error) {
	var encoding *base64.Encoding = base64.StdEncoding

	isURL, err := core.BoolArg(kwargs, "url")
	if err != nil {
		return nil, err
	}
	if isURL {
		encoding = base64.URLEncoding
	}

	isRaw, err := core.BoolArg(kwargs, "raw")
	if err != nil {
		return nil, err
	}
	if isRaw {
		encoding = encoding.WithPadding(base64.NoPadding)
	}

	isStrict, err := core.BoolArg(kwargs, "strict")
	if err != nil {
		return nil, err
	}
	if isStrict {
		encoding = encoding.Strict()
	}

	return encoding, nil
}
