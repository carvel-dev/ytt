package yttlibrary

import (
	"encoding/json"
	"fmt"

	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	JSONAPI = starlark.StringDict{
		"json": &starlarkstruct.Module{
			Name: "json",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("json.encode", core.ErrWrapper(jsonModule{}.Encode)),
				"decode": starlark.NewBuiltin("json.decode", core.ErrWrapper(jsonModule{}.Decode)),
			},
		},
	}
)

type jsonModule struct{}

func (b jsonModule) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val := core.NewStarlarkValue(args.Index(0)).AsInterface()
	val = yamlmeta.NewInterfaceFromAST(val)                           // convert yaml fragments into Go objects
	val = core.NewGoValue(val, false).AsValueWithCheckedMapKeys(true) // JSON only supports string keys

	valBs, err := json.Marshal(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(string(valBs)), nil
}

func (b jsonModule) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	valEncoded, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var valDecoded interface{}

	err = json.Unmarshal([]byte(valEncoded), &valDecoded)
	if err != nil {
		return starlark.None, err
	}

	valDecoded = core.NewGoValue(valDecoded, false).AsValueWithCheckedMapKeys(false)

	return core.NewGoValue(valDecoded, false).AsStarlarkValue(), nil
}
