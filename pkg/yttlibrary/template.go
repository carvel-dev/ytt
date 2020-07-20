package yttlibrary

import (
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
)

type templateModule struct {
	replaceNodeFunc core.StarlarkFunc
}

func NewTemplateModule(replaceNodeFunc core.StarlarkFunc) templateModule {
	return templateModule{replaceNodeFunc}
}

func (b templateModule) AsModule() starlark.StringDict {
	return starlark.StringDict{
		"template": &starlarkstruct.Module{
			Name: "template",
			Members: starlark.StringDict{
				"replace": starlark.NewBuiltin("template.replace", core.ErrWrapper(b.replaceNodeFunc)),
			},
		},
	}
}
