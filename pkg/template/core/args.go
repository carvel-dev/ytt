// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
)

func BoolArg(kwargs []starlark.Tuple, keyToFind string) (bool, error) {
	for _, arg := range kwargs {
		key, err := NewStarlarkValue(arg.Index(0)).AsString()
		if err != nil {
			return false, err
		}
		if key == keyToFind {
			return NewStarlarkValue(arg.Index(1)).AsBool()
		}
	}
	return false, nil
}

func Int64Arg(kwargs []starlark.Tuple, keyToFind string) (int64, error) {
	for _, arg := range kwargs {
		key, err := NewStarlarkValue(arg.Index(0)).AsString()
		if err != nil {
			return 0, err
		}
		if key == keyToFind {
			return NewStarlarkValue(arg.Index(1)).AsInt64()
		}
	}
	return 0, nil
}

func CheckArgNames(kwargs []starlark.Tuple, validKeys map[string]struct{}) error {
	for _, arg := range kwargs {
		key, err := NewStarlarkValue(arg.Index(0)).AsString()
		if err != nil {
			return err
		}
		if _, ok := validKeys[key]; !ok {
			return fmt.Errorf("invalid argument name: %s", key)
		}
	}
	return nil
}
