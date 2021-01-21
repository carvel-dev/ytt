// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/filepos"
)

type Type interface {
	AssignTypeTo(typeable Typeable) TypeCheck
	GetValueType() Type
	CheckType(node TypeWithValues) TypeCheck
}

type TypeWithValues interface {
	GetValues() []interface{}
	GetPosition() *filepos.Position
	ValueTypeAsString() string
}

type Typeable interface {
	// TODO: extract methods common to Node and Typeable to a shared interface?
	GetPosition() *filepos.Position

	SetType(Type)
}

var _ Typeable = (*Document)(nil)
var _ Typeable = (*Map)(nil)
var _ Typeable = (*MapItem)(nil)
var _ Typeable = (*Array)(nil)
var _ Typeable = (*ArrayItem)(nil)

func (d *Document) SetType(t Type)   { d.Type = t }
func (m *Map) SetType(t Type)        { m.Type = t }
func (mi *MapItem) SetType(t Type)   { mi.Type = t }
func (a *Array) SetType(t Type)      { a.Type = t }
func (ai *ArrayItem) SetType(t Type) { ai.Type = t }
