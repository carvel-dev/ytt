package workspace

import (
	"fmt"
	"strings"

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

	return (&libraryValue{foundLib, nil, l.libraryExecutionFactory}).AsStarlarkValue(), nil
}

type libraryValue struct {
	library                 *Library
	dataValuess             []EvalValuesAst
	libraryExecutionFactory *LibraryExecutionFactory
}

func (l *libraryValue) AsStarlarkValue() starlark.Value {
	// TODO technically not a module; switch to struct?
	return &starlarkstruct.Module{
		Name: "library instance",
		Members: starlark.StringDict{
			"with_data_values": starlark.NewBuiltin("library.with_data_values", core.ErrWrapper(l.WithDataValues)),
			"result":           starlark.NewBuiltin("library.result", core.ErrWrapper(l.Result)),
			"export":           starlark.NewBuiltin("library.export", core.ErrWrapper(l.Export)),
		},
	}
}

func (l *libraryValue) WithDataValues(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	dataValues := core.NewStarlarkValue(args.Index(0)).AsInterface()

	libVal := &libraryValue{l.library, nil, l.libraryExecutionFactory}
	libVal.dataValuess = append([]EvalValuesAst{}, l.dataValuess...)
	libVal.dataValuess = append(libVal.dataValuess, yamlmeta.NewASTFromInterface(dataValues))

	return libVal.AsStarlarkValue(), nil
}

func (l *libraryValue) Result(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no arguments")
	}

	libraryLoader := l.libraryExecutionFactory.New(l.library)

	astValues, err := libraryLoader.Values(l.dataValuess)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryLoader.Eval(astValues)
	if err != nil {
		return starlark.None, err
	}

	return yamltemplate.NewStarlarkFragment(result.DocSet), nil
}

func (l *libraryValue) Export(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	symbolName, locationPath, err := l.exportArgs(args, kwargs)
	if err != nil {
		return starlark.None, err
	}

	if strings.HasPrefix(symbolName, "_") {
		return starlark.None, fmt.Errorf(
			"Symbols starting with '_' are private, and cannot be exported")
	}

	libraryLoader := l.libraryExecutionFactory.New(l.library)

	astValues, err := libraryLoader.Values(l.dataValuess)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryLoader.Eval(astValues)
	if err != nil {
		return starlark.None, err
	}

	foundExports := []EvalExport{}

	for _, exp := range result.Exports {
		if _, found := exp.Symbols[symbolName]; found {
			if len(locationPath) == 0 || locationPath == exp.Path {
				foundExports = append(foundExports, exp)
			}
		}
	}

	switch len(foundExports) {
	case 0:
		return starlark.None, fmt.Errorf(
			"Expected to find exported symbol '%s', but did not", symbolName)

	case 1:
		return foundExports[0].Symbols[symbolName], nil

	default:
		var paths []string
		for _, exp := range foundExports {
			paths = append(paths, exp.Path)
		}

		return starlark.None, fmt.Errorf("Expected to find exactly "+
			"one exported symbol '%s', but found multiple across files: %s",
			symbolName, strings.Join(paths, ", "))
	}
}

func (l *libraryValue) exportArgs(args starlark.Tuple, kwargs []starlark.Tuple) (string, string, error) {
	if args.Len() != 1 {
		return "", "", fmt.Errorf("expected exactly one argument")
	}

	symbolName, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return "", "", err
	}

	var locationPath string

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))

		switch kwargName {
		case "path":
			var err error
			locationPath, err = core.NewStarlarkValue(kwarg[1]).AsString()
			if err != nil {
				return "", "", err
			}

		default:
			return "", "", fmt.Errorf("Unexpected keyword argument '%s'", kwargName)
		}
	}

	return symbolName, locationPath, nil
}
