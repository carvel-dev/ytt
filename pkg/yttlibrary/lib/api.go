package lib

import (
	"fmt"

	"github.com/k14s/ytt/pkg/eval"
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
				"get":      starlark.NewBuiltin("lib.get", core.ErrWrapper(libModule{loader}.Get)),
				"external": starlark.NewBuiltin("lib.external", core.ErrWrapper(libModule{loader}.External)),
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

	return m.load(thread, f, args, kwargs, true)
}

func (m libModule) External(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return m.load(thread, f, args, kwargs, false)
}

func (m libModule) load(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple,
	internal bool) (starlark.Value, error) {

	var recursive bool = true

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	for _, kw := range kwargs {
		if kw.Len() != 2 {
			return starlark.None, fmt.Errorf("expected 2 values in kwarg tuples")
		}

		kwname, err := core.NewStarlarkValue(kw.Index(0)).AsString()
		if err != nil {
			return starlark.None, err
		}

		switch kwname {
		case "recursive":
			recursive, err = core.NewStarlarkValue(kw.Index(1)).AsBool()
			if err != nil {
				return starlark.None, err
			}
		}
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var lib eval.Library

	if internal {
		lib, err = m.loader.LoadInternalLibrary(libPath, nil)
	} else {
		lib, err = m.loader.LoadExternalLibrary([]string{libPath}, recursive, nil)
	}
	if err != nil {
		return starlark.None, err
	}

	lib2 := &library{lib, false}
	res := map[interface{}]interface{}{
		"eval": starlark.NewBuiltin("lib.eval", core.ErrWrapper(lib2.Eval)),
	}

	return core.NewGoValue(res, true).AsStarlarkValue(), nil
}

type library struct {
	eval.Library
	evaluated bool
}

func (lib *library) Eval(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if lib.evaluated {
		return starlark.None, fmt.Errorf("eval() cannot be called twice on loaded library")
	}

	var argsvalues interface{}

	if args.Len() == 1 {
		argsvalues = core.NewStarlarkValue(args.Index(0)).AsInterface()
	} else if args.Len() > 1 {
		return starlark.None, fmt.Errorf("eval() cannot be called with more than one argument")
	}

	lib.evaluated = true

	if argsvalues != nil {
		err := lib.FilterValues(eval.NewValuesFilter(yamlmeta.NewASTFromInterface(argsvalues)))
		if err != nil {
			return starlark.None, err
		}
	}

	res, err := lib.Library.Eval()
	if err != nil {
		return starlark.None, err
	}

	res2 := &evaluationResult{res}
	resObj := map[interface{}]interface{}{
		"output": starlark.NewBuiltin("lib.output", core.ErrWrapper(res2.Output)),
		"text":   starlark.NewBuiltin("lib.text", core.ErrWrapper(res2.Text)),
	}

	return core.NewGoValue(resObj, true).AsStarlarkValue(), nil
}

type evaluationResult struct {
	*eval.Result
}

func (res *evaluationResult) Output(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() == 1 {
		fname, err := core.NewStarlarkValue(args.Index(0)).AsString()
		if err != nil {
			return starlark.None, err
		}

		if val, ok := res.DocSets[fname]; ok {
			return yamltemplate.NewStarlarkFragment(val), nil
		} else {
			return starlark.None, nil
		}

	} else if args.Len() == 0 {
		return yamltemplate.NewStarlarkFragment(res.DocSet), nil

	} else {
		return starlark.None, fmt.Errorf("expected zero or one argument")
	}
}

func (res *evaluationResult) Text(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() == 1 {
		fname, err := core.NewStarlarkValue(args.Index(0)).AsString()
		if err != nil {
			return starlark.None, err
		}

		for _, f := range res.OutputFiles {
			if f.RelativePath() == fname {
				return starlark.String(string(f.Bytes())), nil
			}
		}

		return starlark.None, nil

	} else if args.Len() == 0 {
		combinedDocBytes, err := res.DocSet.AsBytes()
		if err != nil {
			return starlark.None, fmt.Errorf("Marshaling combined template result: %s", err)
		}

		return starlark.String(string(combinedDocBytes)), nil

	} else {
		return starlark.None, fmt.Errorf("expected zero or one argument")
	}
}
