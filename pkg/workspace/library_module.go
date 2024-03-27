// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/workspace/ref"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yamltemplate"
	"carvel.dev/ytt/pkg/yttlibrary/overlay"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

// LibraryModule is the definition of the ytt-supplied Starlark module `@ytt:library`
//
// This module produces library instances (see Get()). The configuration is copied from the
// library execution to the new instance of a library module (see NewLibraryModule())
type LibraryModule struct {
	libraryCtx              LibraryExecutionContext
	libraryExecutionFactory *LibraryExecutionFactory
	libraryValues           []*datavalues.Envelope
	librarySchemas          []*datavalues.SchemaEnvelope
}

func NewLibraryModule(libraryCtx LibraryExecutionContext,
	libraryExecutionFactory *LibraryExecutionFactory,
	libraryValues []*datavalues.Envelope, librarySchemas []*datavalues.SchemaEnvelope) LibraryModule {

	return LibraryModule{libraryCtx, libraryExecutionFactory, libraryValues, librarySchemas}
}

// AsModule defines the contents of the "@ytt:library" module.
// Produces an instance of the "@ytt:library" module suitable to be consumed by starlark.Thread.Load()
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

// Get is a starlark.Builtin that, given the path to a private library, returns an instance of it as a libraryValue.
func (b LibraryModule) Get(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	libAlias, tplLoaderOptsOverrides, wantsToSkipDVValidations, err := b.getOpts(kwargs)
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
	dataValuess := append([]*datavalues.Envelope{}, b.libraryValues...)
	libraryCtx := LibraryExecutionContext{Current: foundLib, Root: foundLib}

	return (&libraryValue{libPath, libAlias, dataValuess, b.librarySchemas, libraryCtx,
		b.libraryExecutionFactory.
			WithTemplateLoaderOptsOverrides(tplLoaderOptsOverrides).
			ThatSkipsDataValuesValidations(wantsToSkipDVValidations),
	}).AsStarlarkValue(), nil
}

func (b LibraryModule) getOpts(kwargs []starlark.Tuple) (string, TemplateLoaderOptsOverrides, bool, error) {
	var alias string
	var overrides TemplateLoaderOptsOverrides
	var skipDataValuesValidation = false

	for _, kwarg := range kwargs {
		name, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return "", overrides, false, err
		}

		switch name {
		case "alias":
			val, err := core.NewStarlarkValue(kwarg[1]).AsString()
			if err != nil {
				return "", overrides, false, err
			}
			alias = val

		case "ignore_unknown_comments":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, false, err
			}
			overrides.IgnoreUnknownComments = &result

		case "implicit_map_key_overrides":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, false, err
			}
			overrides.ImplicitMapKeyOverrides = &result

		case "strict":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, false, err
			}
			overrides.StrictYAML = &result

		case "dangerous_data_values_disable_validation":
			result, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return "", overrides, false, err
			}
			skipDataValuesValidation = result

		default:
			return "", overrides, false, fmt.Errorf("Unexpected kwarg '%s'", name)
		}
	}

	return alias, overrides, skipDataValuesValidation, nil
}

// libraryValue is an instance of a private library.
// Instances are immutable.
type libraryValue struct {
	path        string
	alias       string
	dataValuess []*datavalues.Envelope
	schemas     []*datavalues.SchemaEnvelope

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

	allowedKWArgs := map[string]struct{}{
		"plain": {},
	}
	if err := core.CheckArgNames(kwargs, allowedKWArgs); err != nil {
		return starlark.None, err
	}

	usePlainMerge, err := core.BoolArg(kwargs, "plain", false)
	if err != nil {
		return starlark.None, err
	}

	dvDoc := &yamlmeta.Document{
		Value:    yamlmeta.NewASTFromInterfaceWithNoPosition(dataValues),
		Position: filepos.NewUnknownPosition(),
	}

	if usePlainMerge {
		err = overlay.AnnotateForPlainMerge(dvDoc)
		if err != nil {
			return starlark.None, err
		}
	}
	valsYAML, err := datavalues.NewEnvelope(dvDoc)
	if err != nil {
		return starlark.None, err
	}

	newDataValuess := append([]*datavalues.Envelope{}, l.dataValuess...)
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

	newDocSchema, err := datavalues.NewSchemaEnvelope(&yamlmeta.Document{
		Value:    yamlmeta.NewASTFromInterface(libSchema),
		Position: filepos.NewUnknownPosition(),
	})
	if err != nil {
		return starlark.None, err
	}

	newLibSchemas := append([]*datavalues.SchemaEnvelope{}, l.schemas...)
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

func (l *libraryValue) librarySchemas(ll *LibraryExecution) (*datavalues.Schema, []*datavalues.SchemaEnvelope, error) {
	var schemasForCurrentLib, schemasForChildLib []*datavalues.SchemaEnvelope

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

func (l *libraryValue) libraryValues(ll *LibraryExecution, schema *datavalues.Schema) (*datavalues.Envelope, []*datavalues.Envelope, error) {
	var dvss, afterLibModDVss, childDVss []*datavalues.Envelope
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
