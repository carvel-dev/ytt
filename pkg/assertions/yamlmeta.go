package assertions

import "github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"

func AddValidations(node yamlmeta.Node, rules []Rule) {
	metas := node.GetMeta("validations")
	if currRules, ok := metas.([]Rule); ok {
		rules = append(currRules, rules...)
	}
	SetValidations(node, rules)
}

func SetValidations(node yamlmeta.Node, rules []Rule) {
	node.SetMeta("validations", rules)
}

func GetValidations(node yamlmeta.Node) []Rule {
	metas := node.GetMeta("validations")
	if rules, ok := metas.([]Rule); ok {
		return rules
	}
	return nil
}
