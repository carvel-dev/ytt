package overlay

import (
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func (o OverlayOp) mergeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newItem, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
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

	replace, err := o.apply(leftMap.Items[leftIdx].Value, newItem.Value, matchChildDefaults)
	if err != nil {
		return err
	}
	if replace {
		leftMap.Items[leftIdx].Value = newItem.Value
	}

	return nil
}

func (o OverlayOp) removeMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
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

func (o OverlayOp) replaceMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	replaceAnn, err := NewReplaceAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdx, found, err := ann.Index(leftMap)
	if err != nil {
		return err
	}

	if found {
		newVal, err := replaceAnn.Value(leftMap.Items[leftIdx])
		if err != nil {
			return err
		}

		leftMap.Items[leftIdx] = newItem.DeepCopy()
		leftMap.Items[leftIdx].SetValue(newVal)
	}

	return nil
}

func (o OverlayOp) assertMapItem(leftMap *yamlmeta.Map, newItem *yamlmeta.MapItem,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newItem, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewMapItemMatchAnnotation(newItem, parentMatchChildDefaults, o.Thread)
	if err != nil {
		return err
	}

	testAnn, err := NewAssertAnnotation(newItem, o.Thread)
	if err != nil {
		return err
	}

	leftIdx, found, err := ann.Index(leftMap)
	if err != nil {
		return err
	}

	if found {
		err := testAnn.Check(leftMap.Items[leftIdx])
		if err != nil {
			return err
		}

		_, err = o.apply(leftMap.Items[leftIdx].Value, newItem.Value, matchChildDefaults)
		if err != nil {
			return err
		}
	}

	return nil
}
