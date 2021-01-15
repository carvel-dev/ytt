// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"strconv"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template/core"
)

var (
	NodeTagRoot = NodeTag{-100}
)

type Nodes struct {
	id               int
	tagToNode        map[NodeTag]EvaluationNode
	childToParentTag map[NodeTag]NodeTag
}

func NewNodes() *Nodes {
	return &Nodes{
		tagToNode:        map[NodeTag]EvaluationNode{},
		childToParentTag: map[NodeTag]NodeTag{},
	}
}

func (n *Nodes) Ancestors() Ancestors { return NewAncestors(n.childToParentTag) }

func (n *Nodes) AddRootNode(node EvaluationNode) NodeTag {
	n.id++
	tag := NodeTag{n.id}
	n.tagToNode[tag] = node
	n.childToParentTag[tag] = NodeTagRoot
	return tag
}

func (n *Nodes) AddNode(node EvaluationNode, parentTag NodeTag) NodeTag {
	n.id++
	tag := NodeTag{n.id}
	n.tagToNode[tag] = node
	n.childToParentTag[tag] = parentTag
	return tag
}

func (n *Nodes) FindNode(tag NodeTag) (EvaluationNode, bool) {
	node, ok := n.tagToNode[tag]
	return node, ok
}

type NodeTag struct {
	id int
}

func NewNodeTag(id int) NodeTag { return NodeTag{id} }

func NewNodeTagFromStarlarkValue(val starlark.Value) (NodeTag, error) {
	id, err := core.NewStarlarkValue(val).AsInt64()
	if err != nil {
		return NodeTag{}, err
	}
	return NodeTag{int(id)}, nil
}

func (t NodeTag) Equals(other NodeTag) bool { return t.id == other.id }
func (t NodeTag) String() string            { return "node tag " + strconv.Itoa(t.id) }
func (t NodeTag) AsString() string          { return strconv.Itoa(t.id) }
