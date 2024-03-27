// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"carvel.dev/ytt/pkg/yamlmeta"
)

const validationsMeta = "validations"

// Add appends validations to node's validations metadata, later retrieved via Get().
func Add(node yamlmeta.Node, validations []NodeValidation) {
	metas := node.GetMeta(validationsMeta)
	if metas != nil {
		validations = append(metas.([]NodeValidation), validations...)
	}
	Set(node, validations)
}

// Set attaches validations to node's metadata, later retrieved via Get().
func Set(node yamlmeta.Node, meta []NodeValidation) {
	node.SetMeta(validationsMeta, meta)
}

// Get retrieves validations from node metadata, set previously via Set().
func Get(node yamlmeta.Node) []NodeValidation {
	metas := node.GetMeta(validationsMeta)
	if metas == nil {
		return nil
	}
	return metas.([]NodeValidation)
}
