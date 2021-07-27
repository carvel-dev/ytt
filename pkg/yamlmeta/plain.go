// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

func PlainMarshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

func PlainUnmarshal(data []byte, out interface{}) error {
	docSet, err := NewParser(ParserOpts{WithoutComments: true}).ParseBytes(data, "")
	if err != nil {
		return err
	}

	if len(docSet.Items) != 1 {
		return fmt.Errorf("Expected to find exactly one YAML document")
	}

	newVal := docSet.Items[0].AsInterface()

	outVal := reflect.ValueOf(out)
	if newVal == nil {
		outVal.Elem().Set(reflect.Zero(outVal.Elem().Type()))
	} else {
		outVal.Elem().Set(reflect.ValueOf(newVal))
	}

	return nil
}
