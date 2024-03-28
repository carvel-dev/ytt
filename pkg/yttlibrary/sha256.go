// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"crypto/sha256"
	"fmt"

	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

var (
	SHA256API = starlark.StringDict{
		"sha256": &starlarkstruct.Module{
			Name: "sha256",
			Members: starlark.StringDict{
				"sum": starlark.NewBuiltin("sha256.sum", core.ErrWrapper(sha256Module{}.Sum)),
			},
		},
	}
)

type sha256Module struct{}

func (b sha256Module) Sum(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(fmt.Sprintf("%x", sha256.Sum256([]byte(val)))), nil
}
