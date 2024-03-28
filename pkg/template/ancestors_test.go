// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	. "carvel.dev/ytt/pkg/template"
)

func TestAncestorsDeep(t *testing.T) {
	parents := map[NodeTag]NodeTag{
		NewNodeTag(14): NewNodeTag(13),
		NewNodeTag(13): NewNodeTag(12),
		NewNodeTag(12): NewNodeTag(11),
		NewNodeTag(11): NewNodeTag(10),
		NewNodeTag(10): NodeTagRoot,
	}

	comm := NewAncestors(parents).FindCommonParentTag(NewNodeTag(13), NewNodeTag(14))
	if !comm.Equals(NewNodeTag(13)) {
		t.Fatalf("expected 13, returned %d", comm)
	}
}

func TestAncestorsSame(t *testing.T) {
	parents := map[NodeTag]NodeTag{
		NewNodeTag(14): NewNodeTag(13),
		NewNodeTag(13): NewNodeTag(12),
		NewNodeTag(12): NewNodeTag(11),
		NewNodeTag(11): NewNodeTag(10),
		NewNodeTag(10): NodeTagRoot,
	}

	comm := NewAncestors(parents).FindCommonParentTag(NewNodeTag(13), NewNodeTag(13))
	if !comm.Equals(NewNodeTag(12)) {
		t.Fatalf("expected 12, returned %d", comm)
	}
}

func TestAncestorsShallow(t *testing.T) {
	parents := map[NodeTag]NodeTag{
		NewNodeTag(14): NewNodeTag(13),
		NewNodeTag(13): NewNodeTag(12),
		NewNodeTag(12): NewNodeTag(11),
		NewNodeTag(11): NewNodeTag(10),
		NewNodeTag(10): NodeTagRoot,
	}

	comm := NewAncestors(parents).FindCommonParentTag(NewNodeTag(14), NewNodeTag(13))
	if !comm.Equals(NewNodeTag(12)) {
		t.Fatalf("expected 12, returned %d", comm)
	}
}
