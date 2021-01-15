// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func (n *Document) IsEmpty() bool {
	if n.Value == nil {
		return true
	}
	// TODO remove doc empty checks for map and array
	if typedMap, isMap := n.Value.(*Map); isMap {
		return len(typedMap.Items) == 0
	}
	if typedArray, isArray := n.Value.(*Array); isArray {
		return len(typedArray.Items) == 0
	}
	return false
}

func (n *Document) AsYAMLBytes() ([]byte, error) {
	return yaml.Marshal(convertToLowYAML(convertToGo(n.Value)))
}

func (n *Document) AsInterface() interface{} {
	return convertToGo(n.Value)
}
