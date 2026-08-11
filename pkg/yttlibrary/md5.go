// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"crypto/fips140"
	"crypto/md5"
	"fmt"

	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
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

	// MD5 is not a FIPS 140-3 approved algorithm. This function is a general
	// value-hashing convenience (e.g. for cache keys or dedup), not used for
	// authentication or integrity verification, so it is safe to compute
	// even under strict FIPS 140-3-only enforcement (GODEBUG=fips140=only).
	// Without this, building with the native Go FIPS 140-3 module and
	// running with fips140=only would panic on any call to md5.sum().
	var sum [md5.Size]byte
	fips140.WithoutEnforcement(func() {
		sum = md5.Sum([]byte(val))
	})

	return starlark.String(fmt.Sprintf("%x", sum)), nil
}
