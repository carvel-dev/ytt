package yttlibrary

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	StructAPI = starlark.StringDict{
		"struct": &starlarkstruct.Module{
			Name: "struct",
			Members: starlark.StringDict{
				"make":          starlark.NewBuiltin("struct.make", core.ErrWrapper(starlarkstruct.Make)),
				"make_and_bind": starlark.NewBuiltin("struct.make_and_bind", core.ErrWrapper(structModule{}.MakeAndBind)),
				"bind":          starlark.NewBuiltin("struct.bind", core.ErrWrapper(structModule{}.Bind)),

				"encode": starlark.NewBuiltin("struct.encode", core.ErrWrapper(structModule{}.Encode)),
				"decode": starlark.NewBuiltin("struct.decode", core.ErrWrapper(structModule{}.Decode)),
			},
		},
	}
)

type structModule struct{}

func (b structModule) MakeAndBind(thread *starlark.Thread, f *starlark.Builtin,
	bindArgs starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if bindArgs.Len() != 1 {
		return starlark.None, fmt.Errorf("expected at exactly one argument")
	}

	for i, kwarg := range kwargs {
		if _, ok := kwarg[1].(starlark.Callable); ok {
			boundFunc, err := b.Bind(thread, nil, starlark.Tuple{kwarg[1], bindArgs.Index(0)}, nil)
			if err != nil {
				return starlark.None, fmt.Errorf("binding %s: %s", kwarg[0], err)
			}
			kwarg[1] = boundFunc
			kwargs[i] = kwarg
		}
	}

	return starlarkstruct.Make(thread, nil, starlark.Tuple{}, kwargs)
}

func (b structModule) Bind(thread *starlark.Thread, f *starlark.Builtin,
	bindArgs starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {

	if bindArgs.Len() < 2 {
		return starlark.None, fmt.Errorf("expected at least two arguments")
	}

	resultFunc := func(thread *starlark.Thread, f *starlark.Builtin,
		args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		newArgs := append(starlark.Tuple{}, bindArgs[1:]...)
		newArgs = append(newArgs, args...)

		return starlark.Call(thread, bindArgs.Index(0), newArgs, kwargs)
	}

	return starlark.NewBuiltin("struct.bind_result", core.ErrWrapper(resultFunc)), nil
}

func (b structModule) Encode(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	return core.NewGoValueWithOpts(val, core.GoValueOpts{MapIsStruct: true}).AsStarlarkValue(), nil
}

func (b structModule) Decode(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	return core.NewGoValue(val).AsStarlarkValue(), nil
}
