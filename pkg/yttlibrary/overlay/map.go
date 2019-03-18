package overlay

import (
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func (o OverlayOp) mergeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem) error {
	ann, err := NewMapItemMatchAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdx, found, err := ann.Index(leftMap)
	if err != nil {
		return err
	}

	if !found {
		// No need to traverse further
		leftMap.Items = append(leftMap.Items, newItem)
		return nil
	}

	replace, err := o.apply(leftMap.Items[leftIdx].Value, newItem.Value)
	if err != nil {
		return err
	}
	if replace {
		leftMap.Items[leftIdx].Value = newItem.Value
	}

	return nil
}

func (o OverlayOp) removeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem) error {
	ann, err := NewMapItemMatchAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdx, found, err := ann.Index(leftMap)
	if err != nil {
		return err
	}

	if found {
		leftMap.Items = append(leftMap.Items[:leftIdx], leftMap.Items[leftIdx+1:]...)
	}

	return nil
}

func (o OverlayOp) replaceMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem) error {
	ann, err := NewMapItemMatchAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdx, found, err := ann.Index(leftMap)
	if err != nil {
		return err
	}

	if found {
		leftMap.Items[leftIdx] = newItem
	}

	return nil
}
