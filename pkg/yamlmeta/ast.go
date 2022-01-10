// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/filepos"
)

type Node interface {
	GetPosition() *filepos.Position
	SetPosition(*filepos.Position)

	GetValues() []interface{} // ie children
	SetValue(interface{}) error
	AddValue(interface{}) error
	ResetValue()

	GetComments() []*Comment
	SetComments([]*Comment)
	addComments(*Comment)

	GetMeta(name string) interface{}
	SetMeta(name string, data interface{})
	GetAnnotations() interface{}
	SetAnnotations(interface{})

	DeepCopyAsInterface() interface{}
	DeepCopyAsNode() Node

	sealed() // limit the concrete types of Node to map directly only to types allowed in YAML spec.

}

var _ = []Node{&DocumentSet{}, &Map{}, &Array{}}

type DocumentSet struct {
	Comments    []*Comment
	AllComments []*Comment

	Items    []*Document
	Position *filepos.Position

	meta          map[string]interface{}
	annotations   interface{}
	originalBytes *[]byte
}

type Document struct {
	Comments []*Comment
	Value    interface{}
	Position *filepos.Position

	meta        map[string]interface{}
	annotations interface{}
	injected    bool // indicates that Document was not present in the parsed content
}

type Map struct {
	Comments []*Comment
	Items    []*MapItem
	Position *filepos.Position

	meta        map[string]interface{}
	annotations interface{}
}

type MapItem struct {
	Comments []*Comment
	Key      interface{}
	Value    interface{}
	Position *filepos.Position

	meta        map[string]interface{}
	annotations interface{}
}

type Array struct {
	Comments []*Comment
	Items    []*ArrayItem
	Position *filepos.Position

	meta        map[string]interface{}
	annotations interface{}
}

type ArrayItem struct {
	Comments []*Comment
	Value    interface{}
	Position *filepos.Position

	meta        map[string]interface{}
	annotations interface{}
}

type Comment struct {
	Data     string
	Position *filepos.Position
}
