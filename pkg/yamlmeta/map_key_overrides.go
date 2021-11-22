// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/orderedmap"
)

type overrideMapKeys struct{}

// Visit if `node` is a Map, among its MapItem's that have duplicate keys, removes all but the last.
// This visitor always returns `nil`
func (r *overrideMapKeys) Visit(node Node) error {
	mapNode, isMap := node.(*Map)
	if !isMap {
		return nil
	}

	lastItems := orderedmap.NewMap()
	for _, item := range mapNode.Items {
		lastItems.Set(item.Key, item)
	}

	var newItems []*MapItem
	lastItems.Iterate(func(_, value interface{}) {
		newItems = append(newItems, value.(*MapItem))
	})
	mapNode.Items = newItems

	return nil
}
