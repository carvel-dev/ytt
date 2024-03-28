// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package validations_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/experiments"
	"carvel.dev/ytt/pkg/validations"
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

// TestValidations verifies that the validations mechanism correctly:
// 1. parses `@assert/validate` annotations (and because schema package delegates, `@schema/validation` too),
// 2. decides which rules run under various conditions
// 3. checks the rules
// 4. combines rule results into an appropriate validation outcome (list of error or passing)
//
// Example usage:
//
//	Run a specific test:
//	./hack/test-all.sh -v -run TestValidations/filetests/success.tpltest
//
//	Include template compilation results in the output:
//	./hack/test-all.sh -v -run TestValidations/filetests/success.tpltest TestValidations.code=true
//
// see also:
// - pkg/template/... for tests around validations used in schema and data values.
// - pkg/yamltemplate/filetests/assert/... for tests of the exact behaviors of named rules.
func TestValidations(t *testing.T) {
	ft := filetests.FileTests{}
	ft.PathToTests = "filetests"
	ft.ShowTemplateCode = showTemplateCode(kvArg("TestValidations.code"))
	ft.EvalFunc = EvalAndValidateTemplate(ft)

	ft.Run(t)
}

func EvalAndValidateTemplate(ft filetests.FileTests) filetests.EvaluateTemplate {
	return func(src string) (filetests.MarshalableResult, *filetests.TestErr) {
		result, testErr := ft.DefaultEvalTemplate(src)
		if testErr != nil {
			return nil, testErr
		}

		err := validations.ProcessAssertValidateAnns(result.(yamlmeta.Node))
		if err != nil {
			return nil, filetests.NewTestErr(err, fmt.Errorf("Failed to process @assert/validate annotations: %s", err))
		}

		chk, err := validations.Run(result.(yamlmeta.Node), "template-test")
		if err != nil {
			err := fmt.Errorf("\n%s", err)
			return nil, filetests.NewTestErr(err, fmt.Errorf("Unexpected error (did you include the \"ERR:\" marker in the output?):%v", err))
		}
		// TODO: proper error handling!
		if chk.HasInvalidations() {
			err := fmt.Errorf("\n%s", chk.ResultsAsString())
			return nil, filetests.NewTestErr(err, fmt.Errorf("Unexpected violations (did you include the \"ERR:\" marker in the output?):%v", err))
		}

		return result, testErr
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
