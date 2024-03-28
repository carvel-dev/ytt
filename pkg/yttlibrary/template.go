// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

type TemplateModule struct {
	replaceNodeFunc core.StarlarkFunc
}

func NewTemplateModule(replaceNodeFunc core.StarlarkFunc) TemplateModule {
	return TemplateModule{replaceNodeFunc}
}

func (b TemplateModule) AsModule() starlark.StringDict {
	return starlark.StringDict{
		"template": &starlarkstruct.Module{
			Name: "template",
			Members: starlark.StringDict{
				"replace": starlark.NewBuiltin("template.replace", core.ErrWrapper(b.replaceNodeFunc)),
			},
		},
	}
}
