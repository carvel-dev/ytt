// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func NewGoValueWithYAML(val interface{}) tplcore.GoValue {
	convertFunc := func(valToConvert interface{}) (starlark.Value, bool) {
		switch valToConvert.(type) {
		case *yamlmeta.Map, *yamlmeta.Array, *yamlmeta.DocumentSet:
			return &StarlarkFragment{valToConvert}, true
		case *yamlmeta.MapItem, *yamlmeta.ArrayItem, *yamlmeta.Document:
			panic(fmt.Sprintf("NewComplexGoValue: Unexpected %T in conversion of fragment", valToConvert))
		default:
			return starlark.None, false
		}
	}
	return tplcore.NewGoValueWithOpts(val, tplcore.GoValueOpts{Convert: convertFunc})
}
