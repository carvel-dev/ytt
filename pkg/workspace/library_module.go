package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type LibraryModule struct {
	library                 *Library
	libraryExecutionFactory *LibraryExecutionFactory
}

func NewLibraryModule(library *Library, libraryExecutionFactory *LibraryExecutionFactory) LibraryModule {
	return LibraryModule{library, libraryExecutionFactory}
}

func (b LibraryModule) AsModule() starlark.StringDict {
	return starlark.StringDict{
		"library": &starlarkstruct.Module{
			Name: "library",
			Members: starlark.StringDict{
				"get": starlark.NewBuiltin("library.get", core.ErrWrapper(b.Get)),
			},
		},
	}
}

func (l LibraryModule) Get(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	foundLib, err := l.library.FindAccessibleLibrary(libPath)
	if err != nil {
		return starlark.None, err
	}

	return (&libraryValue{foundLib, l.libraryExecutionFactory}).AsStarlarkValue(), nil
}

type libraryValue struct {
	library                 *Library
	libraryExecutionFactory *LibraryExecutionFactory
}

func (l *libraryValue) AsStarlarkValue() starlark.Value {
	// TODO technically not a module; switch to struct?
	return &starlarkstruct.Module{
		Name: "library instance",
		Members: starlark.StringDict{
			"with_data_values": starlark.NewBuiltin("library.with_data_values", core.ErrWrapper(l.WithDataValues)),
			"result":           starlark.NewBuiltin("library.result", core.ErrWrapper(l.Result)),
			"defs":             starlark.NewBuiltin("library.defs", core.ErrWrapper(l.Defs)),
		},
	}
}

func (l *libraryValue) WithDataValues(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	panic("TODO")

	return nil, nil
}

func (l *libraryValue) Result(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	libraryLoader := l.libraryExecutionFactory.New(l.library)

	astValues, err := libraryLoader.Values(yamlmeta.NewASTFromInterface(orderedmap.NewMap()))
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryLoader.Eval(astValues)
	if err != nil {
		return starlark.None, err
	}

	return yamltemplate.NewStarlarkFragment(result.DocSet), nil
}

func (l *libraryValue) Defs(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	panic("TODO")

	return nil, nil
}
