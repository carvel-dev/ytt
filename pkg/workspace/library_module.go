package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type filteredLibValues struct {
	Lib         []*DataValues
	AfterLibMod []*DataValues
	ChildLib    []*DataValues
}

type LibraryModule struct {
	libraryCtx              LibraryExecutionContext
	libraryExecutionFactory *LibraryExecutionFactory
	libraryValues           []*DataValues
}

func NewLibraryModule(libraryCtx LibraryExecutionContext,
	libraryExecutionFactory *LibraryExecutionFactory,
	libraryValues []*DataValues) LibraryModule {

	return LibraryModule{libraryCtx, libraryExecutionFactory, libraryValues}
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

	libTag, err := l.libTag(kwargs)
	if err != nil {
		return starlark.None, err
	}

	if strings.HasPrefix(libPath, "@") {
		return starlark.None, fmt.Errorf(
			"Expected library '%s' to be specified without '@'", libPath)
	}

	foundLib, err := l.libraryCtx.Current.FindAccessibleLibrary(libPath)
	if err != nil {
		return starlark.None, err
	}

	libraryCtx := LibraryExecutionContext{Current: foundLib, Root: foundLib}

	filteredValues, err := l.getFilteredValues(l.libraryValues, libPath, libTag)
	if err != nil {
		return starlark.None, err
	}

	return (&libraryValue{libPath, libTag, libraryCtx, filteredValues, l.libraryExecutionFactory}).AsStarlarkValue(), nil
}

func (l LibraryModule) libTag(kwargs []starlark.Tuple) (string, error) {
	for _, kwarg := range kwargs {
		name, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return "", err
		}

		val, err := core.NewStarlarkValue(kwarg[1]).AsString()
		if err != nil {
			return "", err
		}

		switch name {
		case "tag":
			return val, nil
		default:
			return "", fmt.Errorf("Unexpected kwarg %s in library module get", name)
		}
	}
	return "", nil
}

func (l LibraryModule) getFilteredValues(values []*DataValues,
	currentLibName, currentLibTag string) (filteredLibValues, error) {

	var filteredValues filteredLibValues

	for _, doc := range values {
		childDVs := doc.UsedInLibrary(LibPathPiece{LibName: currentLibName, Tag: currentLibTag})
		if childDVs != nil {
			doc.MarkUsed()
			if childDVs.HasLib() {
				filteredValues.ChildLib = append(filteredValues.ChildLib, childDVs)
			} else {
				if childDVs.AfterLibMod {
					filteredValues.AfterLibMod = append(filteredValues.AfterLibMod, childDVs)
				} else {
					filteredValues.Lib = append(filteredValues.Lib, childDVs)
				}
			}
		}
	}

	return filteredValues, nil
}

type libraryValue struct {
	desc                    string // used in error messages
	tag                     string
	libraryCtx              LibraryExecutionContext
	filteredValues          filteredLibValues
	libraryExecutionFactory *LibraryExecutionFactory
}

func (l *libraryValue) AsStarlarkValue() starlark.Value {
	evalErrMsg := fmt.Sprintf("Evaluating library '%s'", l.desc)
	exportErrMsg := fmt.Sprintf("Exporting from library '%s'", l.desc)

	// TODO technically not a module; switch to struct?
	return &starlarkstruct.Module{
		Name: "library",
		Members: starlark.StringDict{
			"with_data_values": starlark.NewBuiltin("library.with_data_values", core.ErrWrapper(l.WithDataValues)),
			"eval":             starlark.NewBuiltin("library.eval", core.ErrWrapper(core.ErrDescWrapper(evalErrMsg, l.Eval))),
			"export":           starlark.NewBuiltin("library.export", core.ErrWrapper(core.ErrDescWrapper(exportErrMsg, l.Export))),
			"data_values":      starlark.NewBuiltin("library.data_values", core.ErrWrapper(core.ErrDescWrapper(exportErrMsg, l.DataValues))),
		},
	}
}

func (l *libraryValue) WithDataValues(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	dataValues := core.NewStarlarkValue(args.Index(0)).AsGoValue()

	libVal := &libraryValue{l.desc, l.tag, l.libraryCtx, l.filteredValues, l.libraryExecutionFactory}

	valsYAML, err := NewDataValues(&yamlmeta.Document{
		Value:    yamlmeta.NewASTFromInterface(dataValues),
		Position: filepos.NewUnknownPosition(),
	})
	if err != nil {
		return starlark.None, err
	}

	libVal.filteredValues.Lib = append([]*DataValues{}, l.filteredValues.Lib...)
	libVal.filteredValues.Lib = append(libVal.filteredValues.Lib, valsYAML)

	return libVal.AsStarlarkValue(), nil
}

func (l *libraryValue) Eval(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no arguments")
	}

	libraryLoader := l.libraryExecutionFactory.New(l.libraryCtx)

	astValues, libValues, err := l.libraryValues(libraryLoader)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryLoader.Eval(astValues, libValues)
	if err != nil {
		return starlark.None, err
	}

	return yamltemplate.NewStarlarkFragment(result.DocSet), nil
}

func (l *libraryValue) DataValues(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no arguments")
	}

	libraryLoader := l.libraryExecutionFactory.New(l.libraryCtx)

	astValues, _, err := l.libraryValues(libraryLoader)
	if err != nil {
		return starlark.None, err
	}

	val := core.NewGoValueWithOpts(astValues.Doc.AsInterface(), core.GoValueOpts{MapIsStruct: true})
	return val.AsStarlarkValue(), nil
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

	libraryLoader := l.libraryExecutionFactory.New(l.libraryCtx)

	astValues, libValues, err := l.libraryValues(libraryLoader)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryLoader.Eval(astValues, libValues)
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

func (l *libraryValue) libraryValues(ll *LibraryLoader) (*DataValues, []*DataValues, error) {
	values := append([]*DataValues{}, l.filteredValues.Lib...)
	values = append(values, l.filteredValues.AfterLibMod...)

	finalValues, libValues, err := ll.Values(values)
	if err != nil {
		return nil, nil, err
	}

	libValues = append(libValues, l.filteredValues.ChildLib...)

	return finalValues, libValues, nil
}
