package lib

import (
	"fmt"

	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func NewAPI(loader eval.Loader) starlark.StringDict {
	return starlark.StringDict{
		"lib": &starlarkstruct.Module{
			Name: "lib",
			Members: starlark.StringDict{
				"get": starlark.NewBuiltin("lib.eval", core.ErrWrapper(libModule{loader}.Get)),
			},
		},
	}
}

type libModule struct {
	loader eval.Loader
}

func (m libModule) Get(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected one argument")
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	l := &libraryToLoad{
		loader:  m.loader,
		libPath: libPath,
		values:  []eval.ValuesAst{},
	}

	return l.starlark(), nil
}

type libraryToLoad struct {
	loader  eval.Loader
	libPath string
	values  []eval.ValuesAst
	result  *eval.Result
	evalErr error
}

func (l *libraryToLoad) starlark() starlark.Value {
	obj := orderedmap.Conversion{map[interface{}]interface{}{
		"data_values": starlark.NewBuiltin("lib.data_values", core.ErrWrapper(l.DataValues)),
		"yaml":        starlark.NewBuiltin("lib.yaml", core.ErrWrapper(l.Yaml)),
		"text":        starlark.NewBuiltin("lib.text", core.ErrWrapper(l.Text)),
		"exports":     starlark.NewBuiltin("lib.exports", core.ErrWrapper(l.Exports)),
		"list":        starlark.NewBuiltin("lib.list", core.ErrWrapper(l.List)),
	}}.FromUnorderedMaps()

	return core.NewGoValue(obj, true).AsStarlarkValue()
}

func (l *libraryToLoad) ensureLoaded() error {
	if l.result == nil && l.evalErr == nil {
		l.result, l.evalErr = l.loader.LoadInternalLibrary(l.libPath, l.values...)
	}
	return l.evalErr
}

func (l *libraryToLoad) DataValues(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected one argument")
	}

	argsvalues := yamlmeta.NewASTFromInterface(core.NewStarlarkValue(args.Index(0)).AsInterface())

	newLib := &libraryToLoad{
		loader:  l.loader,
		libPath: l.libPath,
		values:  append(append([]eval.ValuesAst{}, l.values...), argsvalues),
	}

	return newLib.starlark(), nil
}

func (l *libraryToLoad) Yaml(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	err := l.ensureLoaded()
	if err != nil {
		return starlark.None, err
	}

	if args.Len() == 0 {
		return yamltemplate.NewStarlarkFragment(l.result.DocSet), nil

	} else if args.Len() == 1 {
		arg0 := core.NewStarlarkValue(args.Index(0)).AsInterface()

		list, ok := arg0.([]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("expected argument to be a list of strings")
		}

		var resDocSet *yamlmeta.DocumentSet

		for _, listitem := range list {
			fname, ok := listitem.(string)
			if !ok {
				return starlark.None, fmt.Errorf("expected argument list to contain strings")
			}

			docSet := l.result.DocSets[fname]
			if docSet == nil {
				return starlark.None, fmt.Errorf("No YAML result for '%v'", fname)
			}

			if resDocSet == nil {
				resDocSet = docSet
			} else {
				resDocSet = &yamlmeta.DocumentSet{
					Items: append(resDocSet.Items, docSet.Items...),
				}
			}
		}

		if resDocSet != nil {
			return yamltemplate.NewStarlarkFragment(resDocSet), nil
		} else {
			return starlark.None, nil
		}

	} else {
		return starlark.None, fmt.Errorf("expected zero or one list argument")
	}
}

func (l *libraryToLoad) Exports(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	err := l.ensureLoaded()
	if err != nil {
		return starlark.None, err
	}

	if args.Len() == 0 {
		combinedExports := starlark.NewDict(0)
		for _, exports := range l.result.Exports {
			if exports.Values == nil {
				continue
			}
			for k, v := range exports.Values {
				combinedExports.SetKey(starlark.String(k), v)
			}
		}

		return combinedExports, nil

	} else if args.Len() == 1 {
		arg0 := core.NewStarlarkValue(args.Index(0)).AsInterface()

		list, ok := arg0.([]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("expected argument to be a list of strings")
		}

		combinedExports := starlark.NewDict(0)

		for _, listitem := range list {
			fname, ok := listitem.(string)
			if !ok {
				return starlark.None, fmt.Errorf("expected argument list to contain strings")
			}

			found := false
			for _, exports := range l.result.Exports {
				if exports.RelativePath != fname {
					continue
				}
				found = true
				if exports.Values == nil {
					continue
				}
				for k, v := range exports.Values {
					combinedExports.SetKey(starlark.String(k), v)
				}
			}

			if !found {
				return starlark.None, fmt.Errorf("No exports for '%v'", fname)
			}
		}

		return combinedExports, nil

	} else {
		return starlark.None, fmt.Errorf("expected zero or one argument")
	}
}

func (l *libraryToLoad) Text(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	err := l.ensureLoaded()
	if err != nil {
		return starlark.None, err
	}

	if args.Len() == 0 {
		combinedDocBytes, err := l.result.DocSet.AsBytes()
		if err != nil {
			return starlark.None, fmt.Errorf("Marshaling combined template result: %s", err)
		}

		return starlark.String(string(combinedDocBytes)), nil

	} else if args.Len() == 1 {
		arg0 := core.NewStarlarkValue(args.Index(0)).AsInterface()

		list, ok := arg0.([]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("expected argument to be a list of strings")
		}

		var res string

		for _, listitem := range list {
			fname, ok := listitem.(string)
			if !ok {
				return starlark.None, fmt.Errorf("expected argument list to contain strings")
			}

			found := false
			for _, f := range l.result.Files {
				if f.RelativePath() == fname {
					res = res + string(f.Bytes())
					found = true
				}
			}

			if !found {
				return starlark.None, fmt.Errorf("No result for '%v'", fname)
			}
		}

		return starlark.String(res), nil

	} else {
		return starlark.None, fmt.Errorf("expected zero or one argument")
	}
}

func (l *libraryToLoad) List(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	err := l.ensureLoaded()
	if err != nil {
		return starlark.None, err
	}

	var list []starlark.Value
	for _, f := range l.result.Files {
		list = append(list, starlark.String(f.RelativePath()))
	}

	return starlark.NewList(list), nil
}
