package yttlibrary

import (
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	ModuleAPI = starlark.StringDict{
		"module": &starlarkstruct.Module{
			Name: "module",
			Members: starlark.StringDict{
				"make": starlark.NewBuiltin("module.make", core.ErrWrapper(starlarkstruct.MakeModule)),
			},
		},
	}
)
