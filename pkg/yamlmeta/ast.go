package yamlmeta

import (
	"github.com/k14s/ytt/pkg/filepos"
)

type Node interface {
	GetPosition() *filepos.Position

	GetValues() []interface{} // ie children
	SetValue(interface{}) error
	AddValue(interface{}) error
	ResetValue()

	GetMetas() []*Meta
	addMeta(*Meta)

	GetAnnotations() interface{}
	SetAnnotations(interface{})

	DeepCopyAsInterface() interface{}
	DeepCopyAsNode() Node

	_private()
}

var _ []Node = []Node{&DocumentSet{}, &Document{}, &Map{}, &MapItem{}, &Array{}, &ArrayItem{}}

type DocumentSet struct {
	Metas    []*Meta
	AllMetas []*Meta

	Items    []*Document
	Position *filepos.Position

	annotations   interface{}
	originalBytes *[]byte
}

type Document struct {
	Metas    []*Meta
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
	injected    bool // indicates that Document was not present in the parsed content
}

type Map struct {
	Metas    []*Meta
	Items    []*MapItem
	Position *filepos.Position

	annotations interface{}
}

type MapItem struct {
	Metas    []*Meta
	Key      interface{}
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
}

type Array struct {
	Metas    []*Meta
	Items    []*ArrayItem
	Position *filepos.Position

	annotations interface{}
}

type ArrayItem struct {
	Metas    []*Meta
	Value    interface{}
	Position *filepos.Position

	annotations interface{}
}

type Meta struct {
	Data     string
	Position *filepos.Position
}
