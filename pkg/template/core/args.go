package core

import (
	"go.starlark.net/starlark"
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
