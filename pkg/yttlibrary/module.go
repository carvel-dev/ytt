// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
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
