// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"github.com/k14s/ytt/pkg/filepos"
)

type Type interface {
	// Checks whether `value` is permitted within (applicable to map types only)
	CheckAllows(item *MapItem) TypeCheck
	AssignTypeTo(typeable Typeable) TypeCheck
	GetValueType() Type
	CheckType(node TypeWithValues, prependErrorMessage string) TypeCheck
}

type TypeWithValues interface {
	GetValues() []interface{}
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

func (n *Document) SetType(t Type)  { n.Type = t }
func (n *Map) SetType(t Type)       { n.Type = t }
func (n *MapItem) SetType(t Type)   { n.Type = t }
func (n *Array) SetType(t Type)     { n.Type = t }
func (n *ArrayItem) SetType(t Type) { n.Type = t }
