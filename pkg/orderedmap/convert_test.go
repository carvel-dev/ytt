// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package orderedmap_test

import (
	"reflect"
	"testing"

	"github.com/vmware-tanzu/carvel-ytt/pkg/orderedmap"
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
