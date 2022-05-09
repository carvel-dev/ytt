// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

const validations = "validations"

// AddRules appends validation Rules to node's validations metadata, later retrieved via GetRules().
func AddRules(node yamlmeta.Node, rules []Rule) {
	metas := node.GetMeta(validations)
	if metas != nil {
		rules = append(metas.([]Rule), rules...)
	}
	SetRules(node, rules)
}

// SetRules attaches validation Rules to node's metadata, later retrieved via GetRules().
func SetRules(node yamlmeta.Node, rules []Rule) {
	node.SetMeta(validations, rules)
}

// GetRules retrieves validation Rules from node metadata, set previously via SetRules().
func GetRules(node yamlmeta.Node) []Rule {
	metas := node.GetMeta(validations)
	if metas == nil {
		return nil
	}
	return metas.([]Rule)
}
