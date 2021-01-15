// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package orderedmap

import (
	"fmt"
	"sort"
)

type Conversion struct {
	Object interface{}
}

func (c Conversion) AsUnorderedStringMaps() interface{} {
	return c.asUnorderedStringMaps(c.Object)
}

func (c Conversion) asUnorderedStringMaps(object interface{}) interface{} {
	switch typedObj := object.(type) {
	case map[interface{}]interface{}:
		panic("Expected *orderedmap.Map instead of map[interface{}]interface{} in asUnorderedStringMaps")

	case map[string]interface{}:
		panic("Expected *orderedmap.Map instead of map[string]interface{} in asUnorderedStringMaps")

	case *Map:
		result := map[string]interface{}{}
		typedObj.Iterate(func(k, v interface{}) {
			if strK, ok := k.(string); ok {
				result[strK] = c.asUnorderedStringMaps(v)
			} else {
				panic("Expected key to be string")
			}
		})
		return result

	case []interface{}:
		for i, item := range typedObj {
			typedObj[i] = c.asUnorderedStringMaps(item)
		}
		return typedObj

	default:
		return typedObj
	}
}

func (c Conversion) FromUnorderedMaps() interface{} {
	return c.fromUnorderedMaps(c.Object)
}

func (c Conversion) fromUnorderedMaps(object interface{}) interface{} {
	switch typedObj := object.(type) {
	case map[interface{}]interface{}:
		result := NewMap()
		for _, key := range c.sortedMapKeys(c.mapKeysFromInterfaceMap(typedObj)) {
			result.Set(key, c.fromUnorderedMaps(typedObj[key]))
		}
		return result

	case map[string]interface{}:
		result := NewMap()
		for _, key := range c.sortedMapKeys(c.mapKeysFromStringMap(typedObj)) {
			result.Set(key, c.fromUnorderedMaps(typedObj[key.(string)]))
		}
		return result

	case *Map:
		panic("Expected map[interface{}]interface{} instead of *unordered.Map in fromUnorderedMaps")

	case []interface{}:
		for i, item := range typedObj {
			typedObj[i] = c.fromUnorderedMaps(item)
		}
		return typedObj

	default:
		return typedObj
	}
}

func (Conversion) mapKeysFromInterfaceMap(m map[interface{}]interface{}) []interface{} {
	var keys []interface{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (Conversion) mapKeysFromStringMap(m map[string]interface{}) []interface{} {
	var keys []interface{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (Conversion) sortedMapKeys(keys []interface{}) []interface{} {
	sort.Slice(keys, func(i, j int) bool {
		iStr := fmt.Sprintf("%s", keys[i])
		jStr := fmt.Sprintf("%s", keys[j])
		return iStr < jStr
	})
	return keys
}
