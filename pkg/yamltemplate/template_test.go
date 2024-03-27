// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate_test

import (
	"os"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/experiments"
	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/yamlmeta"
	_ "carvel.dev/ytt/pkg/yttlibraryext"
	"carvel.dev/ytt/test/filetests"
)

// TestMain is invoked when any tests are run in this package, *instead of* those tests being run directly.
// This allows for setup to occur before *any* test is run.
func TestMain(m *testing.M) {
	experiments.ResetForTesting()
	os.Setenv(experiments.Env, "validations")

	exitVal := m.Run() // execute the specified tests

	os.Exit(exitVal) // required in order to properly report the error level when tests fail.
}

// Example usage:
//
//	Run a specific test:
//	./hack/test-all.sh -v -run TestYAMLTemplate/filetests/if.tpltest
//
//	Include template compilation results in the output:
//	./hack/test-all.sh -v -run TestYAMLTemplate/filetests/if.tpltest TestYAMLTemplate.code=true
func TestYAMLTemplate(t *testing.T) {
	fileTests := filetests.FileTests{
		PathToTests:      "filetests",
		ShowTemplateCode: showTemplateCode(kvArg("TestYAMLTemplate.code")),
		DataValues:       defaultInput(),
	}
	fileTests.Run(t)
}

func defaultInput() yamlmeta.Document {
	return yamlmeta.Document{
		Value: orderedmap.NewMapWithItems([]orderedmap.MapItem{
			{Key: "int", Value: 123},
			{Key: "intNeg", Value: -49},
			{Key: "float", Value: 123.123},
			{Key: "t", Value: true},
			{Key: "f", Value: false},
			{Key: "nullz", Value: nil},
			{Key: "string", Value: "string"},
			{Key: "map", Value: orderedmap.NewMapWithItems([]orderedmap.MapItem{{Key: "a", Value: 123}})},
			{Key: "list", Value: []interface{}{"a", 123, orderedmap.NewMapWithItems([]orderedmap.MapItem{{Key: "a", Value: 123}})}},
		}),
	}
}

func showTemplateCode(showTemplateCodeFlag string) bool {
	return strings.HasPrefix(strings.ToLower(showTemplateCodeFlag), "t")
}

func kvArg(name string) string {
	name += "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, name) {
			return strings.TrimPrefix(arg, name)
		}
	}
	return ""
}
