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
