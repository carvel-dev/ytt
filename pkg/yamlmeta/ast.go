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
	addComments(*Comment)

	GetAnnotations() interface{}
	SetAnnotations(interface{})

	DeepCopyAsInterface() interface{}
	DeepCopyAsNode() Node

	Check() TypeCheck

	sealed() // limit the concrete types of Node to map directly only to types allowed in YAML spec.
}

type ValueHoldingNode interface {
	Node
	Val() interface{}
}

var _ = []Node{&DocumentSet{}, &Map{}, &Array{}}
var _ = []ValueHoldingNode{&Document{}, &MapItem{}, &ArrayItem{}}

type DocumentSet struct {
	Comments    []*Comment
	AllComments []*Comment

	Items    []*Document
	Position *filepos.Position

	annotations   interface{}
	originalBytes *[]byte
}

type Document struct {
	Type     Type
	Comments []*Comment
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
	injected    bool // indicates that Document was not present in the parsed content
}

type Map struct {
	Type     Type
	Comments []*Comment
	Items    []*MapItem
	Position *filepos.Position

	annotations interface{}
}

type MapItem struct {
	Type     Type
	Comments []*Comment
	Key      interface{}
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
}

type Array struct {
	Type     Type
	Comments []*Comment
	Items    []*ArrayItem
	Position *filepos.Position

	annotations interface{}
}

type ArrayItem struct {
	Type     Type
	Comments []*Comment
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
}

type Scalar struct {
	Position *filepos.Position
	Value    interface{}
}

type Comment struct {
	Data     string
	Position *filepos.Position
}
