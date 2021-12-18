// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/version"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

var (
	// Example usage:
	//   ./hack/test-all.sh -v -run TestYAMLTemplate/filetest/if.tpltest
	//   ./hack/test-all.sh -v -run TestYAMLTemplate/filetest/if.tpltest TestYAMLTemplate.code=true
	showTemplateCodeFlag = kvArg("TestYAMLTemplate.code")
)

// TestYAMLTemplate runs suite of test cases, each described in a separate file, verifying the behavior of templates.
//
// Test cases:
// - are located within ./filetests/
// - conventionally have a .tpltest extension
// - top-half is the template; bottom-half is the expected output; divided by `+++`
//
// Types of template tests:
// - expected output starting with `ERR:` indicate that expected output is an error message
// - expected output starting with `OUTPUT POSITION:` indicate that expected output is "pos" format
// - otherwise expected output is the literal output from template
func TestYAMLTemplate(t *testing.T) {
	var files []string
	version.Version = "0.0.0"

	err := filepath.Walk("filetests", func(walkedPath string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		files = append(files, walkedPath)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to enumerate filetests: %s", err)
	}

	for _, filePath := range files {
		t.Run(filePath, func(t *testing.T) {
			contents, err := ioutil.ReadFile(filePath)
			if err != nil {
				t.Fatal(err)
			}

			pieces := strings.SplitN(string(contents), "\n+++\n\n", 2)

			if len(pieces) != 2 {
				t.Fatalf("expected file %s to include +++ separator", filePath)
			}
			expectedStr := pieces[1]

			result, testErr := evalTemplate(t, pieces[0])

			if strings.HasPrefix(expectedStr, "ERR: ") {
				if testErr == nil {
					err = fmt.Errorf("expected eval error, but did not receive it")
				} else {
					resultStr := testErr.UserErr().Error()
					resultStr = strings.TrimRight(regexp.MustCompile("__ytt_tpl\\d+_").ReplaceAllString(resultStr, "__ytt_tplXXX_"), "\t \n")
					err = expectEquals(resultStr, strings.ReplaceAll(strings.TrimRight(strings.TrimPrefix(expectedStr, "ERR: "), "\t \n"), "__YTT_VERSION__", version.Version))
				}
			} else if strings.HasPrefix(expectedStr, "OUTPUT POSITION:") {
				if testErr == nil {
					resultStr, strErr := asFilePositionsStr(result)
					if strErr != nil {
						err = strErr
					} else {
						err = expectEquals(resultStr, strings.ReplaceAll(strings.TrimPrefix(expectedStr, "OUTPUT POSITION:\n"), "__YTT_VERSION__", version.Version))
					}
				} else {
					err = testErr.TestErr()
				}
			} else {
				if testErr == nil {
					resultStr, strErr := asString(result)
					if strErr != nil {
						err = strErr
					} else {
						err = expectEquals(resultStr, expectedStr)
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

func asFilePositionsStr(result interface{ AsBytes() ([]byte, error) }) (string, error) {
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

func asString(result interface{ AsBytes() ([]byte, error) }) (string, error) {
	resultBytes, err := result.AsBytes()
	if err != nil {
		return "", fmt.Errorf("marshal error: %v", err)
	}

	return string(resultBytes), nil
}

type testErr struct {
	realErr error // error returned to the user
	testErr error // error wrapped with helpful test context
}

func (e testErr) UserErr() error { return e.realErr }
func (e testErr) TestErr() error { return e.testErr }

func evalTemplate(t *testing.T, data string) (interface{ AsBytes() ([]byte, error) }, *testErr) {
	docSet, err := yamlmeta.NewDocumentSetFromBytes([]byte(data), yamlmeta.DocSetOpts{AssociatedName: "stdin"})
	if err != nil {
		return nil, &testErr{err, fmt.Errorf("unmarshal error: %v", err)}
	}

	compiledTemplate, err := yamltemplate.NewTemplate("stdin", yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		return nil, &testErr{err, fmt.Errorf("build error: %v", err)}
	}

	if showTemplateCode(showTemplateCodeFlag) {
		fmt.Printf("### ast:\n")
		docSet.Print(os.Stdout)

		fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())
	}

	loader := stdTemplateLoader{compiledTemplate: compiledTemplate}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		return nil, &testErr{err, fmt.Errorf("eval error: %v\ncode:\n%s", err, compiledTemplate.DebugCodeAsString())}
	}

	typedNewVal, ok := newVal.(interface{ AsBytes() ([]byte, error) })
	if !ok {
		return nil, &testErr{err, fmt.Errorf("expected eval result to be marshalable")}
	}

	if showTemplateCode(showTemplateCodeFlag) {
		fmt.Printf("### result ast:\n")
		typedNewVal.(*yamlmeta.DocumentSet).Print(os.Stdout)
	}
	return typedNewVal, nil
}

func showTemplateCode(showTemplateCodeFlag string) bool {
	return strings.HasPrefix(strings.ToLower(showTemplateCodeFlag), "t")
}

func expectEquals(resultStr, expectedStr string) error {
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

type stdTemplateLoader struct {
	compiledTemplate *template.CompiledTemplate
	template.NoopCompiledTemplateLoader
}

func (l stdTemplateLoader) FindCompiledTemplate(_ string) (*template.CompiledTemplate, error) {
	return l.compiledTemplate, nil
}

func (l stdTemplateLoader) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	api := yttlibrary.NewAPI(l.compiledTemplate.TplReplaceNode,
		yttlibrary.NewDataModule(defaultInput(), nil), nil)
	return api.FindModule(strings.TrimPrefix(module, "@ytt:"))
}

func defaultInput() *yamlmeta.Document {
	return &yamlmeta.Document{
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
