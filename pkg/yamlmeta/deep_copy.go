package yamlmeta

func (n *DocumentSet) DeepCopyAsNode() Node { return n.DeepCopy() }
func (n *Document) DeepCopyAsNode() Node    { return n.DeepCopy() }
func (n *Map) DeepCopyAsNode() Node         { return n.DeepCopy() }
func (n *MapItem) DeepCopyAsNode() Node     { return n.DeepCopy() }
func (n *Array) DeepCopyAsNode() Node       { return n.DeepCopy() }
func (n *ArrayItem) DeepCopyAsNode() Node   { return n.DeepCopy() }

func (n *DocumentSet) DeepCopyAsInterface() interface{} { return n.DeepCopy() }
func (n *Document) DeepCopyAsInterface() interface{}    { return n.DeepCopy() }
func (n *Map) DeepCopyAsInterface() interface{}         { return n.DeepCopy() }
func (n *MapItem) DeepCopyAsInterface() interface{}     { return n.DeepCopy() }
func (n *Array) DeepCopyAsInterface() interface{}       { return n.DeepCopy() }
func (n *ArrayItem) DeepCopyAsInterface() interface{}   { return n.DeepCopy() }

func (n *DocumentSet) DeepCopy() *DocumentSet {
	var newItems []*Document
	for _, item := range n.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &DocumentSet{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		AllMetas: []*Meta(MetaSlice(n.AllMetas).DeepCopy()),

		Items:    newItems,
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
	}
}

func (n *Document) DeepCopy() *Document {
	return &Document{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		Value:    nodeDeepCopy(n.Value),
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
		injected:    n.injected,
	}
}

func (n *Map) DeepCopy() *Map {
	var newItems []*MapItem
	for _, item := range n.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &Map{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		Items:    newItems,
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
	}
}

func (n *MapItem) DeepCopy() *MapItem {
	return &MapItem{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		Key:      n.Key,
		Value:    nodeDeepCopy(n.Value),
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
	}
}

func (n *Array) DeepCopy() *Array {
	var newItems []*ArrayItem
	for _, item := range n.Items {
		newItems = append(newItems, item.DeepCopy())
	}

	return &Array{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		Items:    newItems,
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
	}
}

func (n *ArrayItem) DeepCopy() *ArrayItem {
	return &ArrayItem{
		Metas:    []*Meta(MetaSlice(n.Metas).DeepCopy()),
		Value:    nodeDeepCopy(n.Value),
		Position: n.Position,

		annotations: annotationsDeepCopy(n.annotations),
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
