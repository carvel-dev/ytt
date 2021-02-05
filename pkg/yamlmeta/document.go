// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func (d *Document) IsEmpty() bool {
	if d.Value == nil {
		return true
	}
	// TODO remove doc empty checks for map and array
	if typedMap, isMap := d.Value.(*Map); isMap {
		return len(typedMap.Items) == 0
	}
	if typedArray, isArray := d.Value.(*Array); isArray {
		return len(typedArray.Items) == 0
	}
	return false
}

func (d *Document) AsYAMLBytes() ([]byte, error) {
	return yaml.Marshal(convertToLowYAML(convertToGo(d.Value)))
}

func (d *Document) AsInterface() interface{} {
	return convertToGo(d.Value)
}
