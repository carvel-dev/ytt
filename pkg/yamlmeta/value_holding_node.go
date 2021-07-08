// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

func (d *Document) Val() interface{} {
	return d.Value
}
func (mi *MapItem) Val() interface{} {
	return mi.Value
}
func (ai *ArrayItem) Val() interface{} {
	return ai.Value
}
