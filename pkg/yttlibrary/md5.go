// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"crypto/fips140"
	"crypto/md5"
	"fmt"
	"sync/atomic"

	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

// MD5Module contains the definition of the @ytt:md5 module.
//
// Deprecated: MD5 is a cryptographically broken hash algorithm. This
// module is retained only for backward compatibility with existing
// templates; use the @ytt:sha256 module instead.
type MD5Module struct {
	ui ui.UI
}

// hasWarnedMD5Deprecated indicates whether the deprecation notice for
// this module has been displayed. This flag ensures we display that
// warning only once; more than once and its noise. It's an atomic.Bool
// since Starlark builtins may be invoked concurrently across goroutines
// when ytt is embedded as a library.
var hasWarnedMD5Deprecated atomic.Bool

// NewMD5Module constructs a new instance of MD5Module with the
// configured UI (to enable displaying a warning).
func NewMD5Module(uiArg ui.UI) MD5Module {
	return MD5Module{ui: uiArg}
}

// AsModule produces the corresponding Starlark module definition
// suitable for use in running a Starlark program.
func (b MD5Module) AsModule() starlark.StringDict {
	sumFunc := core.ErrWrapper(b.warnOnCall(b.sum))
	return starlark.StringDict{
		"md5": &starlarkstruct.Module{
			Name: "md5",
			Members: starlark.StringDict{
				"sum": starlark.NewBuiltin("md5.sum", sumFunc),
			},
		},
	}
}

// warnOnCall ensures that if the wrapped function is called, the
// user is warned that md5 is deprecated.
func (b MD5Module) warnOnCall(wrappedFunc core.StarlarkFunc) core.StarlarkFunc {
	return func(
		thread *starlark.Thread,
		f *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		if hasWarnedMD5Deprecated.CompareAndSwap(false, true) {
			b.ui.Warnf("\nWarning: @ytt:md5 module is deprecated " +
				"because MD5 is a cryptographically weak hash " +
				"algorithm; use @ytt:sha256 instead.\n")
		}
		return wrappedFunc(thread, f, args, kwargs)
	}
}

func (MD5Module) sum(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	_ []starlark.Tuple,
) (starlark.Value, error) {
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
