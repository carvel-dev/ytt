// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

func (ds *DocumentSet) DeepCopyAsNode() Node { return ds.DeepCopy() }
func (d *Document) DeepCopyAsNode() Node     { return d.DeepCopy() }
func (m *Map) DeepCopyAsNode() Node          { return m.DeepCopy() }
func (mi *MapItem) DeepCopyAsNode() Node     { return mi.DeepCopy() }
func (a *Array) DeepCopyAsNode() Node        { return a.DeepCopy() }
func (ai *ArrayItem) DeepCopyAsNode() Node   { return ai.DeepCopy() }

func (ds *DocumentSet) DeepCopyAsInterface() interface{} { return ds.DeepCopy() }
func (d *Document) DeepCopyAsInterface() interface{}     { return d.DeepCopy() }
func (m *Map) DeepCopyAsInterface() interface{}          { return m.DeepCopy() }
func (mi *MapItem) DeepCopyAsInterface() interface{}     { return mi.DeepCopy() }
func (a *Array) DeepCopyAsInterface() interface{}        { return a.DeepCopy() }
func (ai *ArrayItem) DeepCopyAsInterface() interface{}   { return ai.DeepCopy() }

func (ds *DocumentSet) DeepCopy() *DocumentSet {
	var newItems []*Document
	for _, item := range ds.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &DocumentSet{
		Metas:    []*Meta(MetaSlice(ds.Metas).DeepCopy()),
		AllMetas: []*Meta(MetaSlice(ds.AllMetas).DeepCopy()),

		Items:    newItems,
		Position: ds.Position,

		annotations: annotationsDeepCopy(ds.annotations),
	}
}

func (d *Document) DeepCopy() *Document {
	return &Document{
		Metas:    []*Meta(MetaSlice(d.Metas).DeepCopy()),
		Value:    nodeDeepCopy(d.Value),
		Position: d.Position,

		annotations: annotationsDeepCopy(d.annotations),
		injected:    d.injected,
	}
}

func (m *Map) DeepCopy() *Map {
	var newItems []*MapItem
	for _, item := range m.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &Map{
		Metas:    []*Meta(MetaSlice(m.Metas).DeepCopy()),
		Items:    newItems,
		Position: m.Position,

		annotations: annotationsDeepCopy(m.annotations),
	}
}

func (mi *MapItem) DeepCopy() *MapItem {
	return &MapItem{
		Metas:    []*Meta(MetaSlice(mi.Metas).DeepCopy()),
		Key:      mi.Key,
		Value:    nodeDeepCopy(mi.Value),
		Position: mi.Position,

		annotations: annotationsDeepCopy(mi.annotations),
	}
}

func (a *Array) DeepCopy() *Array {
	var newItems []*ArrayItem
	for _, item := range a.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &Array{
		Metas:    []*Meta(MetaSlice(a.Metas).DeepCopy()),
		Items:    newItems,
		Position: a.Position,

		annotations: annotationsDeepCopy(a.annotations),
	}
}

func (ai *ArrayItem) DeepCopy() *ArrayItem {
	return &ArrayItem{
		Metas:    []*Meta(MetaSlice(ai.Metas).DeepCopy()),
		Value:    nodeDeepCopy(ai.Value),
		Position: ai.Position,

		annotations: annotationsDeepCopy(ai.annotations),
	}
}

func (n *Meta) DeepCopy() *Meta { return &(*n) }

type MetaSlice []*Meta

func (s MetaSlice) DeepCopy() MetaSlice {
	var result []*Meta
	for _, meta := range s {
		result = append(result, meta.DeepCopy())
	}
	return result
}

func nodeDeepCopy(val interface{}) interface{} {
	if node, ok := val.(Node); ok {
		return node.DeepCopyAsInterface()
	}
	return val
}

func annotationsDeepCopy(anns interface{}) interface{} {
	if anns == nil {
		return nil
	}
	return anns.(interface{ DeepCopyAsInterface() interface{} }).DeepCopyAsInterface()
}
