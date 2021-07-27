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
		Comments:    []*Comment(CommentSlice(ds.Comments).DeepCopy()),
		AllComments: []*Comment(CommentSlice(ds.AllComments).DeepCopy()),

		Items:    newItems,
		Position: ds.Position,

		annotations: annotationsDeepCopy(ds.annotations),
	}
}

func (d *Document) DeepCopy() *Document {
	return &Document{
		Comments: []*Comment(CommentSlice(d.Comments).DeepCopy()),
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
		Comments: []*Comment(CommentSlice(m.Comments).DeepCopy()),
		Items:    newItems,
		Position: m.Position,

		annotations: annotationsDeepCopy(m.annotations),
	}
}

func (mi *MapItem) DeepCopy() *MapItem {
	return &MapItem{
		Comments: []*Comment(CommentSlice(mi.Comments).DeepCopy()),
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
		Comments: []*Comment(CommentSlice(a.Comments).DeepCopy()),
		Items:    newItems,
		Position: a.Position,

		annotations: annotationsDeepCopy(a.annotations),
	}
}

func (ai *ArrayItem) DeepCopy() *ArrayItem {
	return &ArrayItem{
		Comments: []*Comment(CommentSlice(ai.Comments).DeepCopy()),
		Value:    nodeDeepCopy(ai.Value),
		Position: ai.Position,

		annotations: annotationsDeepCopy(ai.annotations),
	}
}

func (n *Comment) DeepCopy() *Comment { return &(*n) }

type CommentSlice []*Comment

func (s CommentSlice) DeepCopy() CommentSlice {
	var result []*Comment
	for _, comment := range s {
		result = append(result, comment.DeepCopy())
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
