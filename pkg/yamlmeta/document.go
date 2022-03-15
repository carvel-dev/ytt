// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"bytes"

	yaml3 "gopkg.in/yaml.v3"
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
	out := bytes.Buffer{}

	encoder := yaml3.NewEncoder(&out)
	encoder.SetIndent(2)
	err := encoder.Encode(convertToYAML3Node(d))
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (d *Document) AsInterface() interface{} {
	return convertToGo(d.Value)
}
