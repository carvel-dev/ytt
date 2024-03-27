// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"strconv"

	"carvel.dev/ytt/pkg/template/core"
	"github.com/k14s/starlark-go/starlark"
)

var (
	NodeTagRoot = NodeTag{-100}
)

// Nodes contain source information that is used when building and compiling a template.
//
// Nodes track a source's structure by:
//   - assigning a unique integer id, NodeTag, for each 'Node'.
//   - storing breadcrumbs to find a Node's Parent.
//   - keeping track each Node's non-template comments via annotations.
type Nodes struct {
	id               int
	tagToNode        map[NodeTag]EvaluationNode
	childToParentTag map[NodeTag]NodeTag
	annotations      map[NodeTag]NodeAnnotations
}

// NewNodes initializes Nodes with empty maps to store info about the source template.
func NewNodes() *Nodes {
	return &Nodes{
		tagToNode:        map[NodeTag]EvaluationNode{},
		childToParentTag: map[NodeTag]NodeTag{},
		annotations:      map[NodeTag]NodeAnnotations{},
	}
}

func (n *Nodes) Ancestors() Ancestors { return NewAncestors(n.childToParentTag) }

// AddRootNode creates a new unique NodeTag to keep track of an EvaluationNode that has no parent.
func (n *Nodes) AddRootNode(node EvaluationNode) NodeTag {
	n.id++
	tag := NodeTag{n.id}
	n.tagToNode[tag] = node
	n.childToParentTag[tag] = NodeTagRoot
	return tag
}

// AddNode creates a new unique NodeTag to keep track of the EvaluationNode and its parent.
func (n *Nodes) AddNode(node EvaluationNode, parentTag NodeTag) NodeTag {
	n.id++
	tag := NodeTag{n.id}
	n.tagToNode[tag] = node
	n.childToParentTag[tag] = parentTag
	return tag
}

// FindNode uses a NodeTag to retrieve an EvaluationNode from the annotations map in Nodes.
func (n *Nodes) FindNode(tag NodeTag) (EvaluationNode, bool) {
	node, ok := n.tagToNode[tag]
	return node, ok
}

// AddAnnotation creates an entry in the annotations map of Nodes.
// The entry is NodeAnnotation with only the annotation's position,
// indexed by NodeTag and AnnotationName.
func (n *Nodes) AddAnnotation(tag NodeTag, ann Annotation) {
	if _, found := n.annotations[tag]; !found {
		n.annotations[tag] = NodeAnnotations{}
	}
	n.annotations[tag][ann.Name] = NodeAnnotation{Position: ann.Position}
}

// FindAnnotation uses a NodeTag, and an AnnotationName to retrieve a NodeAnnotation from the annotations map in Nodes.
func (n *Nodes) FindAnnotation(tag NodeTag, annName AnnotationName) (NodeAnnotation, bool) {
	ann, ok := n.annotations[tag][annName]
	return ann, ok
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
