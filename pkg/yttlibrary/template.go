// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
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
