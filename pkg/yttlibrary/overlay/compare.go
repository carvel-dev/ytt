// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/yamlmeta"
)

type Comparison struct{}

func (b Comparison) Compare(left, right interface{}) (bool, string) {
	switch typedRight := right.(type) {
	case *yamlmeta.DocumentSet:
		panic("Unexpected docset")

	case *yamlmeta.Document:
		typedLeft, isDoc := left.(*yamlmeta.Document)
		if !isDoc {
			return false, fmt.Sprintf("Expected doc, but was %T", left)
		}

		return b.Compare(typedLeft.Value, typedRight.Value)

	case *yamlmeta.Map:
		typedLeft, isMap := left.(*yamlmeta.Map)
		if !isMap {
			return false, fmt.Sprintf("Expected map, but was %T", left)
		}

		for _, rightItem := range typedRight.Items {
			matched := false
			for _, leftItem := range typedLeft.Items {
				if reflect.DeepEqual(leftItem.Key, rightItem.Key) {
					result, explain := b.Compare(leftItem, rightItem)
					if !result {
						return false, explain
					}
					matched = true
				}
			}
			if !matched {
				return false, "Expected at least one map item to match by key"
			}
		}

		return true, ""

	case *yamlmeta.MapItem:
		typedLeft, isMapItem := left.(*yamlmeta.MapItem)
		if !isMapItem {
			return false, fmt.Sprintf("Expected mapitem, but was %T", left)
		}

		return b.Compare(typedLeft.Value, typedRight.Value)

	case *yamlmeta.Array:
		typedLeft, isArray := left.(*yamlmeta.Array)
		if !isArray {
			return false, fmt.Sprintf("Expected array, but was %T", left)
		}

		for i, item := range typedRight.Items {
			if i >= len(typedLeft.Items) {
				return false, "Expected to have matching number of array items"
			}
			result, explain := b.Compare(typedLeft.Items[i].Value, item.Value)
			if !result {
				return false, explain
			}
		}

		return true, ""

	case *yamlmeta.ArrayItem:
		typedLeft, isArrayItem := left.(*yamlmeta.ArrayItem)
		if !isArrayItem {
			return false, fmt.Sprintf("Expected arrayitem, but was %T", left)
		}

		return b.Compare(typedLeft.Value, typedRight.Value)

	default:
		return b.CompareLeafs(left, right)
	}
}

func (b Comparison) CompareLeafs(left, right interface{}) (bool, string) {
	if reflect.DeepEqual(left, right) {
		return true, ""
	}

	if result, _ := b.compareAsInt64s(left, right); result {
		return true, ""
	}

	return false, fmt.Sprintf("Expected leaf values to match %T %T", left, right)
}

func (b Comparison) compareAsInt64s(left, right interface{}) (bool, string) {
	leftVal, ok := b.upcastToInt64(left)
	if !ok {
		return false, "Left obj is upcastable to int64"
	}

	rightVal, ok := b.upcastToInt64(right)
	if !ok {
		return false, "Right obj is upcastable to int64"
	}

	return leftVal == rightVal, "Left and right numbers are not equal"
}

func (b Comparison) upcastToInt64(val interface{}) (int64, bool) {
	switch typedVal := val.(type) {
	case int:
		return int64(typedVal), true
	case int16:
		return int64(typedVal), true
	case int32:
		return int64(typedVal), true
	case int64:
		return int64(typedVal), true
	case int8:
		return int64(typedVal), true
	default:
		return 0, false
	}
}
