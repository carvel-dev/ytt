// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assertions

import "github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"

// AddValidations appends validation Rules to node's validations metadata, later retrieved via GetValidations().
func AddValidations(node yamlmeta.Node, rules []Rule) {
	metas := node.GetMeta("validations")
	if currRules, ok := metas.([]Rule); ok {
		rules = append(currRules, rules...)
	}
	SetValidations(node, rules)
}

// SetValidations attaches validation Rules to node's metadata, later retrieved via GetValidations().
func SetValidations(node yamlmeta.Node, rules []Rule) {
	node.SetMeta("validations", rules)
}

// GetValidations retrieves validation Rules from node metadata, set previously via SetValidations().
func GetValidations(node yamlmeta.Node) []Rule {
	metas := node.GetMeta("validations")
	if rules, ok := metas.([]Rule); ok {
		return rules
	}
	return nil
}
