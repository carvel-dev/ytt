// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"crypto/md5"
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	MD5API = starlark.StringDict{
		"md5": &starlarkstruct.Module{
			Name: "md5",
			Members: starlark.StringDict{
				"sum": starlark.NewBuiltin("md5.sum", core.ErrWrapper(md5Module{}.Sum)),
			},
		},
	}
)

type md5Module struct{}

func (b md5Module) Sum(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(fmt.Sprintf("%x", md5.Sum([]byte(val)))), nil
}
