// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
)

// BoolArg return a boolean value from starlark.Tupe based on a given key, defaults to defaultValue.
func BoolArg(kwargs []starlark.Tuple, keyToFind string, defaultValue bool) (bool, error) {
	for _, arg := range kwargs {
		key, err := NewStarlarkValue(arg.Index(0)).AsString()
		if err != nil {
			return false, err
		}
		if key == keyToFind {
			return NewStarlarkValue(arg.Index(1)).AsBool()
		}
	}
	return defaultValue, nil
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
