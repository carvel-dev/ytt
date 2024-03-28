// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0
package orderedmap_test

import (
	"reflect"
	"testing"

	"carvel.dev/ytt/pkg/orderedmap"
)

func TestFromUnorderedMaps(t *testing.T) {
	inputA := map[string]interface{}{
		"key": []interface{}{map[string]interface{}{"nestedKey": "nestedValue"}},
	}
	inputB := map[string]interface{}{
		"key": []interface{}{map[string]interface{}{"nestedKey": "nestedValue"}},
	}

	orderedmap.Conversion{Object: inputA}.FromUnorderedMaps()

	if !reflect.DeepEqual(inputA, inputB) {
		t.Errorf("Nested object was modified. Got: %v, Expected: %v", inputA, inputB)
	}
}
