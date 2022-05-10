// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

const validationsMeta = "validations"

// AddRules appends validation Rules to node's validations metadata, later retrieved via GetRules().
func Add(node yamlmeta.Node, validations []NodeValidation) {
	metas := node.GetMeta(validationsMeta)
	if metas != nil {
		validations = append(metas.([]NodeValidation), validations...)
	}
	Set(node, validations)
}

// SetRules attaches validations to node's metadata, later retrieved via GetRules().
func Set(node yamlmeta.Node, meta []NodeValidation) {
	node.SetMeta(validationsMeta, meta)
}

// GetRules retrieves validations from node metadata, set previously via SetRules().
func Get(node yamlmeta.Node) []NodeValidation {
	metas := node.GetMeta(validationsMeta)
	if metas == nil {
		return nil
	}
	return metas.([]NodeValidation)
}
