// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/workspace/ref"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

type LibraryModule struct {
	libraryCtx              LibraryExecutionContext
	libraryExecutionFactory *LibraryExecutionFactory
	libraryValues           []*DataValues
	librarySchemas          []*schema.DocumentSchemaEnvelope
}

func NewLibraryModule(libraryCtx LibraryExecutionContext,
	libraryExecutionFactory *LibraryExecutionFactory,
	libraryValues []*DataValues, librarySchemas []*schema.DocumentSchemaEnvelope) LibraryModule {

	return LibraryModule{libraryCtx, libraryExecutionFactory, libraryValues, librarySchemas}
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

func (b LibraryModule) Get(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	libAlias, tplLoaderOptsOverrides, err := b.getOpts(kwargs)
	if err != nil {
		return starlark.None, err
	}

	if strings.HasPrefix(libPath, "@") {
		return starlark.None, fmt.Errorf(
			"Expected library '%s' to be specified without '@'", libPath)
	}

	foundLib, err := b.libraryCtx.Current.FindAccessibleLibrary(libPath)
	if err != nil {
		return starlark.None, err
	}

	// copy over library values
	dataValuess := append([]*DataValues{}, b.libraryValues...)
	libraryCtx := LibraryExecutionContext{Current: foundLib, Root: foundLib}

	return (&libraryValue{libPath, libAlias, dataValuess, b.librarySchemas, libraryCtx,
		b.libraryExecutionFactory.WithTemplateLoaderOptsOverrides(tplLoaderOptsOverrides),
	}).AsStarlarkValue(), nil
}

func (b LibraryModule) getOpts(kwargs []starlark.Tuple) (string, TemplateLoaderOptsOverrides, error) {
	var alias string
	var overrides TemplateLoaderOptsOverrides

	for _, kwarg := range kwargs {
		name, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return "", overrides, err
		}

		switch name {
		case "alias":
			val, err := core.NewStarlarkValue(kwarg[1]).AsString()
			if err != nil {
				return "", overrides, err
			}
			alias = val

		case "ignore_unknown_comments":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, err
			}
			overrides.IgnoreUnknownComments = &result

		case "implicit_map_key_overrides":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, err
			}
			overrides.ImplicitMapKeyOverrides = &result

		case "strict":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, err
			}
			overrides.StrictYAML = &result

		default:
			return "", overrides, fmt.Errorf("Unexpected kwarg '%s'", name)
		}
	}

	return alias, overrides, nil
}

type libraryValue struct {
	path        string
	alias       string
	dataValuess []*DataValues
	schemas     []*schema.DocumentSchemaEnvelope

	libraryCtx              LibraryExecutionContext
	libraryExecutionFactory *LibraryExecutionFactory
}

func (l *libraryValue) AsStarlarkValue() starlark.Value {
	desc := ref.LibraryRef{Path: l.path, Alias: l.alias}.AsString()
	evalErrMsg := fmt.Sprintf("Evaluating library '%s'", desc)
	exportErrMsg := fmt.Sprintf("Exporting from library '%s'", desc)

	// TODO technically not a module; switch to struct?
	return &starlarkstruct.Module{
		Name: "library",
		Members: starlark.StringDict{
			"with_data_values":        starlark.NewBuiltin("library.with_data_values", core.ErrWrapper(l.WithDataValues)),
			"with_data_values_schema": starlark.NewBuiltin("library.with_data_values_schema", core.ErrWrapper(l.WithDataValuesSchema)),
			"eval":                    starlark.NewBuiltin("library.eval", core.ErrWrapper(core.ErrDescWrapper(evalErrMsg, l.Eval))),
			"export":                  starlark.NewBuiltin("library.export", core.ErrWrapper(core.ErrDescWrapper(exportErrMsg, l.Export))),
			"data_values":             starlark.NewBuiltin("library.data_values", core.ErrWrapper(core.ErrDescWrapper(exportErrMsg, l.DataValues))),
		},
	}
}

func (l *libraryValue) WithDataValues(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	dataValues, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	valsYAML, err := NewDataValues(&yamlmeta.Document{
		Value:    yamlmeta.NewASTFromInterfaceWithNoPosition(dataValues),
		Position: filepos.NewUnknownPosition(),
	})
	if err != nil {
		return starlark.None, err
	}

	// copy over library values
	newDataValuess := append([]*DataValues{}, l.dataValuess...)
	newDataValuess = append(newDataValuess, valsYAML)

	libVal := &libraryValue{l.path, l.alias, newDataValuess, l.schemas, l.libraryCtx, l.libraryExecutionFactory}

	return libVal.AsStarlarkValue(), nil
}

func (l *libraryValue) WithDataValuesSchema(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	libSchema, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}

	newDocSchema, err := schema.NewDocumentSchemaEnvelope(&yamlmeta.Document{
		Value:    yamlmeta.NewASTFromInterface(libSchema),
		Position: filepos.NewUnknownPosition(),
	})
	if err != nil {
		return starlark.None, err
	}

	newLibSchemas := append([]*schema.DocumentSchemaEnvelope{}, l.schemas...)
	newLibSchemas = append(newLibSchemas, newDocSchema)

	libVal := &libraryValue{l.path, l.alias, l.dataValuess, newLibSchemas, l.libraryCtx, l.libraryExecutionFactory}

	return libVal.AsStarlarkValue(), nil
}

func (l *libraryValue) Eval(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no arguments")
	}

	libraryExecution := l.libraryExecutionFactory.New(l.libraryCtx)

	schema, librarySchemas, err := l.librarySchemas(libraryExecution)
	if err != nil {
		return starlark.None, err
	}
	astValues, libValues, err := l.libraryValues(libraryExecution, schema)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryExecution.Eval(astValues, libValues, librarySchemas)
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

	libraryExecution := l.libraryExecutionFactory.New(l.libraryCtx)

	schema, _, err := l.librarySchemas(libraryExecution)
	if err != nil {
		return starlark.None, err
	}
	astValues, _, err := l.libraryValues(libraryExecution, schema)
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

	libraryExecution := l.libraryExecutionFactory.New(l.libraryCtx)

	schema, librarySchemas, err := l.librarySchemas(libraryExecution)
	if err != nil {
		return starlark.None, err
	}
	astValues, libValues, err := l.libraryValues(libraryExecution, schema)
	if err != nil {
		return starlark.None, err
	}

	result, err := libraryExecution.Eval(astValues, libValues, librarySchemas)
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

func (l *libraryValue) librarySchemas(ll *LibraryExecution) (Schema, []*schema.DocumentSchemaEnvelope, error) {
	var schemasForCurrentLib, schemasForChildLib []*schema.DocumentSchemaEnvelope

	for _, docSchema := range l.schemas {
		matchingSchema, usedInCurrLibrary := docSchema.UsedInLibrary(ref.LibraryRef{Path: l.path, Alias: l.alias})
		if usedInCurrLibrary {
			schemasForCurrentLib = append(schemasForCurrentLib, matchingSchema)
		} else {
			schemasForChildLib = append(schemasForChildLib, matchingSchema)
		}
	}

	schema, librarySchemas, err := ll.Schemas(schemasForCurrentLib)
	if err != nil {
		return nil, nil, err
	}

	foundChildSchemas := append(librarySchemas, schemasForChildLib...)
	return schema, foundChildSchemas, nil
}

func (l *libraryValue) libraryValues(ll *LibraryExecution, schema Schema) (*DataValues, []*DataValues, error) {
	var dvss, afterLibModDVss, childDVss []*DataValues
	for _, dv := range l.dataValuess {
		matchingDVs := dv.UsedInLibrary(ref.LibraryRef{Path: l.path, Alias: l.alias})
		if matchingDVs != nil {
			if matchingDVs.IntendedForAnotherLibrary() {
				childDVss = append(childDVss, matchingDVs)
			} else {
				if matchingDVs.AfterLibMod {
					afterLibModDVss = append(afterLibModDVss, matchingDVs)
				} else {
					dvss = append(dvss, matchingDVs)
				}
			}
		}
	}

	dvs, foundChildDVss, err := ll.Values(append(dvss, afterLibModDVss...), schema)
	if err != nil {
		return nil, nil, err
	}

	// Order data values specified in a parent library, on top of
	// data values specified within a child library
	foundChildDVss = append(foundChildDVss, childDVss...)

	return dvs, foundChildDVss, nil
}
