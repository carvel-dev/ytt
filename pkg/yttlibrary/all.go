package yttlibrary

import (
	"github.com/get-ytt/ytt/pkg/template"
	tplcore "github.com/get-ytt/ytt/pkg/template/core"
	"github.com/get-ytt/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type API map[string]starlark.StringDict

func NewAPI(replaceNodeFunc tplcore.StarlarkFunc, values interface{},
	loader template.CompiledTemplateLoader) API {

	return map[string]starlark.StringDict{
		"@ytt:assert": AssertAPI,
		"@ytt:regexp": RegexpAPI,

		// Hashes
		"@ytt:md5":    MD5API,
		"@ytt:sha256": SHA256API,

		// Serializations
		"@ytt:base64": Base64API,
		"@ytt:json":   JSONAPI,
		"@ytt:yaml":   YAMLAPI,

		// Templating
		"@ytt:template": NewTemplateModule(replaceNodeFunc).AsModule(),
		"@ytt:data":     NewDataModule(values, loader).AsModule(),

		// Object building
		"@ytt:struct":  StructAPI,
		"@ytt:module":  ModuleAPI,
		"@ytt:overlay": overlay.API,
	}
}
