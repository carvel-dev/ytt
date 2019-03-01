package yttlibrary

import (
	"fmt"
	"regexp"

	"github.com/get-ytt/ytt/pkg/template/core"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	RegexpAPI = starlark.StringDict{
		"regexp": &starlarkstruct.Module{
			Name: "regexp",
			Members: starlark.StringDict{
				"match": starlark.NewBuiltin("regexp.match", core.ErrWrapper(regexpModule{}.Match)),
			},
		},
	}
)

type regexpModule struct{}

func (b regexpModule) Match(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return starlark.None, fmt.Errorf("expected exactly two arguments")
	}

	pattern, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	target, err := core.NewStarlarkValue(args.Index(1)).AsString()
	if err != nil {
		return starlark.None, err
	}

	matched, err := regexp.MatchString(pattern, target)
	if err != nil {
		return starlark.None, err
	}

	return starlark.Bool(matched), nil
}
