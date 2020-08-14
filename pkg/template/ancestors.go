// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
)

type Ancestors struct {
	nodeNumToParentNum map[NodeTag]NodeTag
}

func NewAncestors(nodeNumToParentNum map[NodeTag]NodeTag) Ancestors {
	return Ancestors{nodeNumToParentNum}
}

func (e Ancestors) FindParentTag(tag NodeTag) NodeTag {
	parentTag, ok := e.nodeNumToParentNum[tag]
	if !ok {
		panic(fmt.Sprintf("expected to find parent tag for %s", tag))
	}
	return parentTag
}

func (e Ancestors) FindCommonParentTag(currTag, newTag NodeTag) NodeTag {
	currAncestors := e.ancestors([]NodeTag{currTag}, currTag)
	newAncestors := e.ancestors([]NodeTag{}, newTag)
	commonAncestor := NodeTagRoot

	for i := 0; i < max(len(currAncestors), len(newAncestors)); i++ {
		if i == len(currAncestors) || i == len(newAncestors) {
			break
		}
		if currAncestors[i].Equals(newAncestors[i]) {
			commonAncestor = currAncestors[i]
		} else {
			break
		}
	}

	if false { // for debugging
		fmt.Printf("---\n")
		fmt.Printf("inst: %d\n", newTag)
		fmt.Printf("curr: %#v\n", currAncestors)
		fmt.Printf("new : %#v\n", newAncestors)
		fmt.Printf("comm: %d\n", commonAncestor)
	}

	return commonAncestor
}

func (e Ancestors) ancestors(result []NodeTag, tag NodeTag) []NodeTag {
	for {
		parentTag, ok := e.nodeNumToParentNum[tag]
		if !ok {
			panic(fmt.Sprintf("expected to find parent tag for %s", tag))
		}
		result = append([]NodeTag{parentTag}, result...)
		if parentTag.Equals(NodeTagRoot) {
			break
		}
		tag = parentTag
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
