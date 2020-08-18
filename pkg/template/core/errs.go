// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"runtime/debug"
)

type StarlarkFunc func(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

func ErrWrapper(wrappedFunc StarlarkFunc) StarlarkFunc {
	return func(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (val starlark.Value, resultErr error) {
		// Catch any panics to give a better contextual information
		defer func() {
			if err := recover(); err != nil {
				if typedErr, ok := err.(error); ok {
					resultErr = fmt.Errorf("%s (backtrace: %s)", typedErr, debug.Stack())
				} else {
					resultErr = fmt.Errorf("(p) %s (backtrace: %s)", err, debug.Stack())
				}
			}
		}()

		val, err := wrappedFunc(thread, f, args, kwargs)
		if err != nil {
			return val, fmt.Errorf("%s: %s", f.Name(), err)
		}

		return val, nil
	}
}

func ErrDescWrapper(desc string, wrappedFunc StarlarkFunc) StarlarkFunc {
	return func(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		val, err := wrappedFunc(thread, f, args, kwargs)
		if err != nil {
			return val, fmt.Errorf("%s: %s", desc, err)
		}
		return val, nil
	}
}
