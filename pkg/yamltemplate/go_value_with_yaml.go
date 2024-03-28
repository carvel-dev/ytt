// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"

	tplcore "carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/starlark-go/starlark"
)

func NewGoValueWithYAML(val interface{}) tplcore.GoValue {
	convertFunc := func(valToConvert interface{}) (starlark.Value, bool) {
		switch valToConvert.(type) {
		case *yamlmeta.Map, *yamlmeta.Array, *yamlmeta.DocumentSet:
			return &StarlarkFragment{valToConvert}, true
		case *yamlmeta.MapItem, *yamlmeta.ArrayItem, *yamlmeta.Document:
			panic(fmt.Sprintf("Unexpected %#v in conversion of fragment", valToConvert))
		default:
			return starlark.None, false
		}
	}
	return tplcore.NewGoValueWithOpts(val, tplcore.GoValueOpts{Convert: convertFunc})
}
