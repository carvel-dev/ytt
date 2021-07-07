// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	EvaluationCtxDialectName template.EvaluationCtxDialectName = "yaml"
)

type EvaluationCtx struct {
	implicitMapKeyOverrides bool
}

var _ template.EvaluationCtxDialect = EvaluationCtx{}

func (e EvaluationCtx) PrepareNode(
	parentNode template.EvaluationNode, val template.EvaluationNode) error {

	if typedMap, ok := parentNode.(*yamlmeta.Map); ok {
		if typedMapItem, ok := val.(*yamlmeta.MapItem); ok {
			return MapItemOverride{e.implicitMapKeyOverrides}.Apply(typedMap, typedMapItem, true)
		}
	}
	return nil
}

func (e EvaluationCtx) SetMapItemKey(node template.EvaluationNode, val interface{}) error {
	if item, ok := node.(*yamlmeta.MapItem); ok {
		item.Key = val
		return nil
	}

	panic(fmt.Sprintf("expected node '%T' to be MapItem", node))
}

func (e EvaluationCtx) Replace(
	parentNodes []template.EvaluationNode, val interface{}) error {

	switch typedCurrNode := parentNodes[len(parentNodes)-1].(type) {
	case *yamlmeta.Document:
		if len(parentNodes) < 2 {
			return fmt.Errorf("expected to find document set, but was not enough parents")
		}

		parentNode := parentNodes[len(parentNodes)-2]
		typedParentNode, ok := parentNode.(*yamlmeta.DocumentSet)
		if !ok {
			return fmt.Errorf("expected to find document set, but was %T", parentNode)
		}

		return e.replaceItemInDocSet(typedParentNode, typedCurrNode, val)

	case *yamlmeta.MapItem:
		if len(parentNodes) < 2 {
			return fmt.Errorf("expected to find map, but was not enough parents")
		}

		parentNode := parentNodes[len(parentNodes)-2]
		typedParentNode, ok := parentNode.(*yamlmeta.Map)
		if !ok {
			return fmt.Errorf("expected parent of map item to be map, but was %T", parentNode)
		}

		return e.replaceItemInMap(typedParentNode, typedCurrNode, val)

	case *yamlmeta.ArrayItem:
		if len(parentNodes) < 2 {
			return fmt.Errorf("expected to find array, but was not enough parents")
		}

		parentNode := parentNodes[len(parentNodes)-2]
		typedParentNode, ok := parentNode.(*yamlmeta.Array)
		if !ok {
			return fmt.Errorf("expected parent of array item to be array, but was %T", parentNode)
		}

		return e.replaceItemInArray(typedParentNode, typedCurrNode, val)

	default:
		return fmt.Errorf("expected to replace document value, map item or array item, but found %T", typedCurrNode)
	}
}

func (e EvaluationCtx) replaceItemInDocSet(dstDocSet *yamlmeta.DocumentSet, placeholderItem *yamlmeta.Document, val interface{}) error {
	insertItems, err := e.convertValToDocSetItems(val)
	if err != nil {
		return err
	}

	for i, item := range dstDocSet.Items {
		if item == placeholderItem {
			newItems := dstDocSet.Items[:i]
			newItems = append(newItems, insertItems...)
			newItems = append(newItems, dstDocSet.Items[i+1:]...)
			dstDocSet.Items = newItems
			return nil
		}
	}

	return fmt.Errorf("expected to find placeholder doc in docset")
}

func (e EvaluationCtx) convertValToDocSetItems(val interface{}) ([]*yamlmeta.Document, error) {
	result := []*yamlmeta.Document{}

	switch typedVal := val.(type) {
	case []interface{}:
		for _, item := range typedVal {
			result = append(result, &yamlmeta.Document{Value: item, Position: filepos.NewUnknownPosition()})
		}

	case *yamlmeta.DocumentSet:
		result = typedVal.Items

	default:
		return nil, fmt.Errorf("expected value to be docset, but was %T", val)
	}

	return result, nil
}

func (e EvaluationCtx) replaceItemInMap(
	dstMap *yamlmeta.Map, placeholderItem *yamlmeta.MapItem, val interface{}) error {

	insertItems, carryMeta, err := e.convertValToMapItems(val, placeholderItem.Position.DeepCopy())
	if err != nil {
		return err
	}

	// If map items does not carry metadata
	// we cannot check for override conflicts
	for _, newItem := range insertItems {
		err := MapItemOverride{e.implicitMapKeyOverrides}.Apply(dstMap, newItem, carryMeta)
		if err != nil {
			return err
		}
	}

	for i, item := range dstMap.Items {
		if item == placeholderItem {
			newItems := dstMap.Items[:i]
			newItems = append(newItems, insertItems...)
			newItems = append(newItems, dstMap.Items[i+1:]...)
			dstMap.Items = newItems
			return nil
		}
	}

	return fmt.Errorf("expected to find placeholder map item in map")
}

func (e EvaluationCtx) convertValToMapItems(val interface{}, position *filepos.Position) ([]*yamlmeta.MapItem, bool, error) {
	switch typedVal := val.(type) {
	case *orderedmap.Map:
		result := []*yamlmeta.MapItem{}
		typedVal.Iterate(func(k, v interface{}) {
			item := &yamlmeta.MapItem{Key: k, Value: yamlmeta.NewASTFromInterfaceWithPosition(v, position), Position: position}
			result = append(result, item)
		})
		return result, false, nil

	case *yamlmeta.Map:
		return typedVal.Items, true, nil

	default:
		return nil, false, fmt.Errorf("expected value to be map, but was %T", val)
	}
}

func (e EvaluationCtx) replaceItemInArray(dstArray *yamlmeta.Array, placeholderItem *yamlmeta.ArrayItem, val interface{}) error {
	insertItems, err := e.convertValToArrayItems(val, placeholderItem.Position.DeepCopy())
	if err != nil {
		return err
	}

	for i, item := range dstArray.Items {
		if item == placeholderItem {
			newItems := dstArray.Items[:i]
			newItems = append(newItems, insertItems...)
			newItems = append(newItems, dstArray.Items[i+1:]...)
			dstArray.Items = newItems
			return nil
		}
	}

	return fmt.Errorf("expected to find placeholder array item in array")
}

func (e EvaluationCtx) convertValToArrayItems(val interface{}, position *filepos.Position) ([]*yamlmeta.ArrayItem, error) {
	result := []*yamlmeta.ArrayItem{}

	switch typedVal := val.(type) {
	case []interface{}:
		for _, item := range typedVal {
			result = append(result, &yamlmeta.ArrayItem{Value: yamlmeta.NewASTFromInterfaceWithPosition(item, position), Position: position})
		}

	case *yamlmeta.Array:
		result = typedVal.Items

	default:
		return nil, fmt.Errorf("expected value to be array, but was %T", val)
	}

	return result, nil
}

func (e EvaluationCtx) ShouldWrapRootValue(nodeVal interface{}) bool {
	switch nodeVal.(type) {
	case *yamlmeta.Document, *yamlmeta.MapItem, *yamlmeta.ArrayItem:
		return true
	default:
		return false
	}
}

func (e EvaluationCtx) WrapRootValue(val interface{}) interface{} {
	return &StarlarkFragment{val}
}
