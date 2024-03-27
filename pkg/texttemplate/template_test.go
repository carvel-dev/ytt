// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package texttemplate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/texttemplate"
	"github.com/k14s/starlark-go/starlark"
)

var (
	selectedFileTestPath = kvArg("TestTextTemplate.filetest")
	showTemplateCode     = kvArg("TestTextTemplate.code")
	showErrs             = kvArg("TestTextTemplate.errs")
)

func TestTextTemplate(t *testing.T) {
	files, err := os.ReadDir("filetests")
	if err != nil {
		t.Fatal(err)
	}

	if len(selectedFileTestPath) > 0 {
		fmt.Printf("only running %s test(s)\n", selectedFileTestPath)
	}

	var errs []error

	for _, file := range files {
		filePath := filepath.Join("filetests", file.Name())

		if len(selectedFileTestPath) > 0 && !strings.HasPrefix(file.Name(), selectedFileTestPath) {
			continue
		}

		testDesc := fmt.Sprintf("checking %s ...\n", file.Name())
		fmt.Printf("%s", testDesc)

		contents, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		const (
			testSep   = "\n+++\n"
			errPrefix = "\nERR:"
		)

		pieces := strings.SplitN(string(contents), testSep, 2)
		if len(pieces) != 2 {
			t.Fatalf("expected file %s to include +++ separator", filePath)
		}

		resultStr, testErr := evalTemplate(t, pieces[0])
		expectedStr := pieces[1]

		if strings.HasPrefix(expectedStr, errPrefix) {
			if testErr == nil {
				err = fmt.Errorf("expected eval error, but did not receive it")
			} else {
				resultStr := testErr.UserErr().Error()
				resultStr = regexp.MustCompile("__ytt_tpl\\d+_").ReplaceAllString(resultStr, "__ytt_tplXXX_")
				err = expectEquals(t, resultStr, strings.TrimPrefix(expectedStr, errPrefix))
			}
		} else {
			if testErr == nil {
				err = expectEquals(t, resultStr, expectedStr)
			} else {
				err = testErr.TestErr()
			}
		}

		if err != nil {
			fmt.Printf("   FAIL\n")
			if showErrs == "t" {
				sep := strings.Repeat(".", 80)
				fmt.Printf("%s\n%s%s\n", sep, err, sep)
			}
			errs = append(errs, fmt.Errorf("%s: %s", testDesc, err))
		} else {
			fmt.Printf("   .\n")
		}
	}

	if len(errs) > 0 {
		t.Errorf("%s", errs[0].Error())
	}

	if len(selectedFileTestPath) > 0 {
		t.Errorf("skipped tests")
	}
}

type testErr struct {
	realErr error // error returned to the user
	testErr error // error wrapped with helpful test context
}

func (e testErr) UserErr() error { return e.realErr }
func (e testErr) TestErr() error { return e.testErr }

func evalTemplate(t *testing.T, data string) (string, *testErr) {
	textRoot, err := texttemplate.NewParser().Parse([]byte(data), "stdin")
	if err != nil {
		return "", &testErr{err, fmt.Errorf("template parse error: %v", err)}
	}

	compiledTemplate, err := texttemplate.NewTemplate("stdin").Compile(textRoot)
	if err != nil {
		return "", &testErr{err, fmt.Errorf("template build error: %v", err)}
	}

	if showTemplateCode == "t" {
		fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())
	}

	loader := singleTemplateLoader{compiledTemplate: compiledTemplate}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		return "", &testErr{err, fmt.Errorf("eval error: %v\ncode:\n%s", err, compiledTemplate.DebugCodeAsString())}
	}

	resultStr := newVal.(*texttemplate.NodeRoot).AsString()

	return resultStr, nil
}

type singleTemplateLoader struct {
	compiledTemplate *template.CompiledTemplate
	template.NoopCompiledTemplateLoader
}

func (l singleTemplateLoader) FindCompiledTemplate(_ string) *template.CompiledTemplate {
	return l.compiledTemplate
}

func expectEquals(t *testing.T, resultStr, expectedStr string) error {
	if resultStr != expectedStr {
		return fmt.Errorf("not equal\n\n### result %d chars:\n>>>%s<<<\n###expected %d chars:\n>>>%s<<<", len(resultStr), resultStr, len(expectedStr), expectedStr)
	}
	return nil
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
