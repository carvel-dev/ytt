// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"

	"carvel.dev/ytt/pkg/yamlmeta"
)

// checkDataValueKeysAreStrings returns an error if any map in the data values has a key that is not a string.
//
// Templates see data values as structs, whose attribute names have to be strings.
func checkDataValueKeysAreStrings(doc *yamlmeta.Document) error {
	return yamlmeta.Walk(doc, nonStringKeyChecker{})
}

type nonStringKeyChecker struct{}

// Visit returns an error if `node` is a map item whose key is not a string.
func (nonStringKeyChecker) Visit(node yamlmeta.Node) error {
	item, ok := node.(*yamlmeta.MapItem)
	if !ok {
		return nil
	}
	if _, isString := item.Key.(string); isString {
		return nil
	}

	msg := "Data value key"
	if item.Position.IsKnown() {
		msg += " at " + item.Position.AsCompactString()
	}
	msg += fmt.Sprintf(" must be a string, but was %s (%v)", yamlmeta.TypeName(item.Key), item.Key)
	if _, isBool := item.Key.(bool); isBool {
		msg += " (hint: in YAML, unquoted on, off, yes and no are booleans; quote the key to make it a string)"
	}
	return fmt.Errorf("%s", msg)
}
