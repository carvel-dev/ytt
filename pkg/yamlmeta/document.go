package yamlmeta

func (d *Document) IsEmpty() bool {
	if d.Value == nil {
		return true
	}
	// TODO remove doc empty checks for map and array
	if typedMap, isMap := d.Value.(*Map); isMap {
		return len(typedMap.Items) == 0
	}
	if typedArray, isArray := d.Value.(*Array); isArray {
		return len(typedArray.Items) == 0
	}
	return false
}

func (d *Document) AsInterface(opts InterfaceConvertOpts) interface{} {
	return convertToLowYAML(d.Value, opts)
}
