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
				"eval": starlark.NewBuiltin("lib.eval", core.ErrWrapper(libModule{loader}.Eval)),
			},
		},
	}
}

type libModule struct {
	loader eval.Loader
}

func (m libModule) Eval(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var argsvalues interface{}

	if args.Len() < 1 || args.Len() > 2 {
		return starlark.None, fmt.Errorf("expected one or two arguments")
	}
	if args.Len() >= 2 {
		argsvalues = yamlmeta.NewASTFromInterface(core.NewStarlarkValue(args.Index(1)).AsInterface())
	}

	libPath, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var res *eval.Result

	res, err = m.loader.LoadInternalLibrary(libPath, argsvalues)
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

		for _, f := range res.Files {
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
