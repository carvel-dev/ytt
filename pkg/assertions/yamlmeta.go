// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assertions

import (
	"fmt"

	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

const validations = "validations"

// AddValidations appends validation Rules to node's validations metadata, later retrieved via GetValidations().
func AddValidations(node yamlmeta.Node, rules []Rule) {
	metas := node.GetMeta(validations)
	if metas != nil {
		currRules, ok := metas.([]Rule)
		if !ok {
			panic(fmt.Sprintf("Unexpected validations in meta : %s %v", node.GetPosition().AsCompactString(), metas))
		}
		rules = append(currRules, rules...)
	}
	SetValidations(node, rules)
}

// SetValidations attaches validation Rules to node's metadata, later retrieved via GetValidations().
func SetValidations(node yamlmeta.Node, rules []Rule) {
	node.SetMeta(validations, rules)
}

// GetValidations retrieves validation Rules from node metadata, set previously via SetValidations().
func GetValidations(node yamlmeta.Node) []Rule {
	metas := node.GetMeta(validations)
	if rules, ok := metas.([]Rule); ok {
		return rules
	}
	return nil
}
