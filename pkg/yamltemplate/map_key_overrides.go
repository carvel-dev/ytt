// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationMapKeyOverride template.AnnotationName = "yaml/map-key-override"
)

type MapItemOverride struct {
	implicit bool
}

func (d MapItemOverride) Apply(
	typedMap *yamlmeta.Map, newItem *yamlmeta.MapItem, strict bool) error {

	itemIndex := map[interface{}]int{}

	for idx, item := range typedMap.Items {
		itemIndex[item.Key] = idx
	}

	if prevIdx, ok := itemIndex[newItem.Key]; ok {
		if d.implicit || template.NewAnnotations(newItem).Has(AnnotationMapKeyOverride) || !strict {
			typedMap.Items = append(typedMap.Items[:prevIdx], typedMap.Items[prevIdx+1:]...)
			return nil
		}

		return fmt.Errorf("expected key '%s' to not be specified again "+
			"(unless '%s' annotation is added)", newItem.Key, AnnotationMapKeyOverride)
	}

	return nil
}
