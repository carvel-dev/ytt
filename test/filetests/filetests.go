// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package filetests houses a test harness for evaluating templates and asserting
the expected output.
*/
package filetests

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/version"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yamltemplate"
	"carvel.dev/ytt/pkg/yttlibrary"
	"github.com/k14s/starlark-go/starlark"
)

// MarshalableResult is a template evaluation result that can be (likely) marshaled into a slice of bytes.
type MarshalableResult interface {
	AsBytes() ([]byte, error)
}

// EvaluateTemplate is the processing desired from a source template to the final result.
type EvaluateTemplate func(src string) (MarshalableResult, *TestErr)

// FileTests contain a suite of test cases, each described in a separate file, verifying the behavior of templates.
//
// Test cases:
// - are found within the directory at "PathToTests"
// - conventionally have a .tpltest extension
// - top-half is the template; bottom-half is the expected output; divided by `+++` and a blank line.
//
// Types of template tests:
// - expected output starting with `ERR:` indicate that expected output is an error message
// - expected output starting with `OUTPUT POSITION:` indicate that expected output is "pos" format
// - otherwise expected output is the literal output from template
//
// For example:
//
//	#! my-test.tpltest
//	---
//	#@ msg = "hello"
//	msg: #@ msg
//	+++
//
//	msg: hello
type FileTests struct {
	PathToTests      string
	EvalFunc         EvaluateTemplate
	ShowTemplateCode bool
	DataValues       yamlmeta.Document
}

// Run runs each tests: enumerates each file within FileTests.PathToTests; splits and evaluates using FileTests.EvalFunc
// optionally supplying FileTests.DataValues to that evaluation.
//
// If an error occurs and FileTests.ShowTemplateCode is set, then the output includes the debug output of the template.
func (f FileTests) Run(t *testing.T) {
	var files []string
	version.Version = "0.0.0"

	err := filepath.Walk(f.PathToTests, func(walkedPath string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		files = append(files, walkedPath)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to enumerate filetests: %s", err)
	}

	if f.EvalFunc == nil {
		f.EvalFunc = f.DefaultEvalTemplate
	}

	for _, filePath := range files {
		t.Run(filePath, func(t *testing.T) {
			contents, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatal(err)
			}

			pieces := strings.SplitN(string(contents), "\n+++\n\n", 2)

			if len(pieces) != 2 {
				t.Fatalf("expected file %s to include +++ separator", filePath)
			}
			expectedStr := pieces[1]

			result, testErr := f.EvalFunc(pieces[0])

			switch {
			case strings.HasPrefix(expectedStr, "ERR:"):
				if testErr == nil {
					err = fmt.Errorf("expected eval error, but did not receive it")
				} else {
					resultStr := testErr.UserErr().Error()
					resultStr = regexp.MustCompile("__ytt_tpl\\d+_").ReplaceAllString(resultStr, "__ytt_tplXXX_")
					resultStr = TrimTrailingMultilineWhitespace(resultStr)

					expectedStr = strings.TrimPrefix(expectedStr, "ERR:")
					expectedStr = strings.TrimPrefix(expectedStr, " ")
					expectedStr = strings.ReplaceAll(expectedStr, "__YTT_VERSION__", version.Version)
					expectedStr = TrimTrailingMultilineWhitespace(expectedStr)
					err = f.expectEquals(resultStr, expectedStr)
				}
			case strings.HasPrefix(expectedStr, "OUTPUT POSITION:"):
				if testErr == nil {
					resultStr, strErr := f.asFilePositionsStr(result)
					if strErr != nil {
						err = strErr
					} else {
						expectedStr = strings.TrimPrefix(expectedStr, "OUTPUT POSITION:\n")
						expectedStr = strings.ReplaceAll(expectedStr, "__YTT_VERSION__", version.Version)
						err = f.expectEquals(resultStr, expectedStr)
					}
				} else {
					err = testErr.TestErr()
				}
			default:
				if testErr == nil {
					resultStr, strErr := f.asString(result)
					if strErr != nil {
						err = strErr
					} else {
						err = f.expectEquals(resultStr, expectedStr)
					}
				} else {
					err = testErr.TestErr()
				}
			}

			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func (f FileTests) asFilePositionsStr(result MarshalableResult) (string, error) {
	printerFunc := func(w io.Writer) yamlmeta.DocumentPrinter {
		return yamlmeta.WrappedFilePositionPrinter{yamlmeta.NewFilePositionPrinter(w)}
	}
	docSet := result.(*yamlmeta.DocumentSet)
	combinedDocBytes, err := docSet.AsBytesWithPrinter(printerFunc)
	if err != nil {
		return "", fmt.Errorf("expected result docSet to be printable: %v", err)
	}
	return string(combinedDocBytes), nil
}

func (f FileTests) asString(result MarshalableResult) (string, error) {
	resultBytes, err := result.AsBytes()
	if err != nil {
		return "", fmt.Errorf("marshal error: %v", err)
	}

	return string(resultBytes), nil
}

// TestErr captures an error result from a single test.
type TestErr struct {
	realErr error
	testErr error
}

// NewTestErr creates a new TestErr
func NewTestErr(realErr, testErr error) *TestErr {
	return &TestErr{realErr, testErr}
}

// UserErr yields the error returned to the user
func (e TestErr) UserErr() error { return e.realErr }

// TestErr yields the error wrapped with helpful test context
func (e TestErr) TestErr() error { return e.testErr }

func (f FileTests) expectEquals(resultStr, expectedStr string) error {
	if resultStr != expectedStr {
		return fmt.Errorf("not equal\n\n### result %d chars:\n>>>%s<<<\n###expected %d chars:\n>>>%s<<<", len(resultStr), resultStr, len(expectedStr), expectedStr)
	}
	return nil
}

// DefaultEvalTemplate evaluates the YAML template "src"
func (f FileTests) DefaultEvalTemplate(src string) (MarshalableResult, *TestErr) {
	docSet, err := yamlmeta.NewDocumentSetFromBytes([]byte(src), yamlmeta.DocSetOpts{AssociatedName: "stdin"})
	if err != nil {
		return nil, NewTestErr(err, fmt.Errorf("unmarshal error: %v", err))
	}

	compiledTemplate, err := yamltemplate.NewTemplate("stdin", yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		return nil, NewTestErr(err, fmt.Errorf("build error: %v", err))
	}

	if f.ShowTemplateCode {
		fmt.Printf("### ast:\n")
		docSet.Print(os.Stdout)

		fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())
	}

	loader := DefaultTemplateLoader{
		CompiledTemplate: compiledTemplate,
		DataValues:       f.DataValues,
	}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		return nil, NewTestErr(err, fmt.Errorf("eval error: %v\ncode:\n%s", err, compiledTemplate.DebugCodeAsString()))
	}

	typedNewVal, ok := newVal.(MarshalableResult)
	if !ok {
		return nil, NewTestErr(err, fmt.Errorf("expected eval result to be marshalable"))
	}

	if f.ShowTemplateCode {
		fmt.Printf("### result ast:\n")
		typedNewVal.(*yamlmeta.DocumentSet).Print(os.Stdout)
	}
	return typedNewVal, nil
}

// DefaultTemplateLoader is a template.NoopCompiledTemplateLoader that allows any standard library
// module to be loaded (including "@ytt:data", populated with "DataValues")
type DefaultTemplateLoader struct {
	template.NoopCompiledTemplateLoader

	CompiledTemplate *template.CompiledTemplate
	DataValues       yamlmeta.Document
}

// FindCompiledTemplate returns the compiled template that is the subject of the current executing test.
func (l DefaultTemplateLoader) FindCompiledTemplate(_ string) *template.CompiledTemplate {
	return l.CompiledTemplate
}

// Load provides the ability to load any module from the ytt standard library (including `@ytt:data` populated with
// DefaultTemplateLoader.DataValues)
func (l DefaultTemplateLoader) Load(_ *starlark.Thread, module string) (starlark.StringDict, error) {
	api := yttlibrary.NewAPI(l.CompiledTemplate.TplReplaceNode,
		yttlibrary.NewDataModule(&l.DataValues, nil), nil, nil)
	return api.FindModule(strings.TrimPrefix(module, "@ytt:"))
}

// TrimTrailingMultilineWhitespace returns a string with trailing whitespace trimmed from every line as well
// as trimmed trailing empty lines
func TrimTrailingMultilineWhitespace(s string) string {
	var trimmedLines []string
	for _, line := range strings.Split(s, "\n") {
		trimmedLine := strings.TrimRight(line, "\t ")
		trimmedLines = append(trimmedLines, trimmedLine)
	}
	multiline := strings.Join(trimmedLines, "\n")
	return strings.TrimRight(multiline, "\n")
}
