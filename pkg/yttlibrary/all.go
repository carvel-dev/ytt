// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"fmt"

	"carvel.dev/ytt/pkg/cmd/ui"
	tplcore "carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yttlibrary/overlay"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

var registeredExts []*starlarkstruct.Module

// RegisterExt adds "mod" to the standard set of ytt library modules as an extension.
// An "extension" is a Starlark module that has external Go dependencies
// (as opposed to other Starlark modules in the ytt library that either
// have no dependencies or depend on the Go standard lib).
// This enables those using ytt as a Go module to opt-in (rather than be forced)
// to accept such dependencies. Only Carvel-maintained extensions can be registered;
// this reserves the `@ytt:` namespace. Integrators who want to write their own extensions
// should construct their own library.
func RegisterExt(mod *starlarkstruct.Module) {
	switch mod.Name {
	case "toml":
		registeredExts = append(registeredExts, mod)
	default:
		panic("ytt library namespace can only be extended with ytt modules")
	}
}

type API struct {
	modules map[string]starlark.StringDict
}

// NewAPI builds an API instance to be used in executing a template.
func NewAPI(
	replaceNodeFunc tplcore.StarlarkFunc,
	dataMod DataModule,
	libraryMod starlark.StringDict,
	ui ui.UI) API {

	std := map[string]starlark.StringDict{
		"assert": NewAssertModule().AsModule(),
		"math":   NewMathModule(ui).AsModule(),
		"regexp": RegexpAPI,

		// Hashes
		"md5":    MD5API,
		"sha256": SHA256API,

		// Serializations
		"base64": Base64API,
		"json":   JSONAPI,
		"yaml":   YAMLAPI,
		"url":    URLAPI,
		"ip":     IPAPI,

		// Templating
		"template": NewTemplateModule(replaceNodeFunc).AsModule(),
		"data":     dataMod.AsModule(),

		// Object building
		"struct":  StructAPI,
		"module":  ModuleAPI,
		"overlay": overlay.API,

		// Versioning
		"version": VersionAPI,

		"library": libraryMod,
	}

	for _, ext := range registeredExts {
		// Double check that we are not overriding predefined library
		if _, found := std[ext.Name]; found {
			panic("Internal inconsistency: shadowing ytt library with an extension module")
		}
		std[ext.Name] = starlark.StringDict{ext.Name: ext}
	}

	return API{std}
}

func (a API) FindModule(module string) (starlark.StringDict, error) {
	if module, found := a.modules[module]; found {
		return module, nil
	}
	return nil, fmt.Errorf("builtin ytt library does not have module '%s' "+
		"(hint: is it available in newer version of ytt?)", module)
}
