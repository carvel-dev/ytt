package yttlibrary

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type API struct {
	modules map[string]starlark.StringDict
}

func NewAPI(replaceNodeFunc tplcore.StarlarkFunc, values *yamlmeta.Document,
	loader template.CompiledTemplateLoader, libraryMod starlark.StringDict) API {

	return API{map[string]starlark.StringDict{
		"assert": AssertAPI,
		"regexp": RegexpAPI,

		// Hashes
		"md5":    MD5API,
		"sha256": SHA256API,

		// Serializations
		"base64": Base64API,
		"json":   JSONAPI,
		"yaml":   YAMLAPI,
		"url":    URLAPI,

		// Templating
		"template": NewTemplateModule(replaceNodeFunc).AsModule(),
		"data":     NewDataModule(values, loader).AsModule(),

		// Object building
		"struct":  StructAPI,
		"module":  ModuleAPI,
		"overlay": overlay.API,

		// Versioning
		"version": VersionAPI,

		"library": libraryMod,
	}}
}

func (a API) FindModule(module string) (starlark.StringDict, error) {
	if module, found := a.modules[module]; found {
		return module, nil
	}
	return nil, fmt.Errorf("builtin ytt library does not have module '%s'", module)
}
