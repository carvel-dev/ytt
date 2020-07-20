package yttlibrary

import (
	"fmt"
	"regexp"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
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
