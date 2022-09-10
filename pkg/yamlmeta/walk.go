// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

// Visitor performs an operation on the given Node while traversing the AST.
// Typically defines the action taken during a Walk().
type Visitor interface {
	Visit(Node) error
}

// Walk traverses the tree starting at `n`, recursively, depth-first, invoking `v` on each node.
// if `v` returns non-nil error, the traversal is aborted.
func Walk(n Node, v Visitor) error {
	err := v.Visit(n)
	if err != nil {
		return err
	}

	for _, c := range n.GetValues() {
		if cn, ok := c.(Node); ok {
			err := Walk(cn, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// VisitorWithParent performs an operation on the given Node while traversing the AST, including a reference to "node"'s
//   parent node.
//
// Typically defines the action taken during a WalkWithParent().
type VisitorWithParent interface {
	VisitWithParent(Node, Node) error
}

// WalkWithParent traverses the tree starting at `n`, recursively, depth-first, invoking `v` on each node and including
//   a reference to "node"s parent node as well.
// if `v` returns non-nil error, the traversal is aborted.
func WalkWithParent(node Node, parent Node, v VisitorWithParent) error {
	err := v.VisitWithParent(node, parent)
	if err != nil {
		return err
	}

	for _, child := range node.GetValues() {
		if childNode, ok := child.(Node); ok {
			err = WalkWithParent(childNode, node, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
