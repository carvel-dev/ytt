package yamltemplate_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"go.starlark.net/starlark"
)

var (
	// Example usage:
	//   ./hack/test.sh -run TestYAMLTemplate TestYAMLTemplate.filetest=yield-def.yml
	selectedFileTestPath = kvArg("TestYAMLTemplate.filetest")
	showTemplateCode     = kvArg("TestYAMLTemplate.code") // eg t|...
	showErrs             = kvArg("TestYAMLTemplate.errs") // eg t|...
)

func TestYAMLTemplate(t *testing.T) {
	var files []string

	err := filepath.Walk("filetests", func(walkedPath string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		files = append(files, walkedPath)
		return nil
	})
	if err != nil {
		t.Fatalf("Listing files")
	}

	if len(selectedFileTestPath) > 0 {
		fmt.Printf("only running %s test(s)\n", selectedFileTestPath)
	}

	var errs []error

	for _, filePath := range files {
		if len(selectedFileTestPath) > 0 && !strings.HasPrefix(filePath, selectedFileTestPath) {
			continue
		}

		testDesc := fmt.Sprintf("checking %s ...\n", filePath)
		fmt.Printf("%s", testDesc)

		contents, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		pieces := strings.SplitN(string(contents), "\n+++\n\n", 2)

		if len(pieces) != 2 {
			t.Fatalf("expected file %s to include +++ separator", filePath)
		}

		resultStr, testErr := evalTemplate(t, pieces[0])
		expectedStr := pieces[1]

		if strings.HasPrefix(expectedStr, "ERR: ") {
			if testErr == nil {
				err = fmt.Errorf("expected eval error, but did not receive it")
			} else {
				resultStr := testErr.UserErr().Error()
				resultStr = regexp.MustCompile("__ytt_tpl\\d+_").ReplaceAllString(resultStr, "__ytt_tplXXX_")
				err = expectEquals(t, resultStr, strings.TrimPrefix(expectedStr, "ERR: "))
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
	docSet, err := yamlmeta.NewDocumentSetFromBytes([]byte(data), yamlmeta.DocSetOpts{AssociatedName: "stdin"})
	if err != nil {
		return "", &testErr{err, fmt.Errorf("unmarshal error: %v", err)}
	}

	compiledTemplate, err := yamltemplate.NewTemplate("stdin", yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		return "", &testErr{err, fmt.Errorf("build error: %v", err)}
	}

	if showTemplateCode == "t" {
		fmt.Printf("### ast:\n")
		docSet.Print(os.Stdout)

		fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())
	}

	loader := stdTemplateLoader{compiledTemplate: compiledTemplate}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		return "", &testErr{err, fmt.Errorf("eval error: %v\ncode:\n%s", err, compiledTemplate.DebugCodeAsString())}
	}

	typedNewVal, ok := newVal.(interface{ AsBytes() ([]byte, error) })
	if !ok {
		return "", &testErr{err, fmt.Errorf("expected eval result to be marshalable")}
	}

	if showTemplateCode == "t" {
		fmt.Printf("### result ast:\n")
		typedNewVal.(*yamlmeta.DocumentSet).Print(os.Stdout)
	}

	resultBytes, err := typedNewVal.AsBytes()
	if err != nil {
		return "", &testErr{err, fmt.Errorf("marshal error: %v", err)}
	}

	return string(resultBytes), nil
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

type stdTemplateLoader struct {
	compiledTemplate *template.CompiledTemplate
	template.NoopCompiledTemplateLoader
}

func (l stdTemplateLoader) FindCompiledTemplate(_ string) (*template.CompiledTemplate, error) {
	return l.compiledTemplate, nil
}

func (l stdTemplateLoader) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	apis := yttlibrary.NewAPI(l.compiledTemplate.TplReplaceNode, defaultInput(), nil, nil)
	if api, found := apis[module]; found {
		return api, nil
	}
	return nil, fmt.Errorf("Unknown module '%s'", module)
}

func defaultInput() interface{} {
	return orderedmap.NewMapWithItems([]orderedmap.MapItem{
		{Key: "int", Value: 123},
		{Key: "intNeg", Value: -49},
		{Key: "float", Value: 123.123},
		{Key: "t", Value: true},
		{Key: "f", Value: false},
		{Key: "nullz", Value: nil},
		{Key: "string", Value: "string"},
		{Key: "map", Value: orderedmap.NewMapWithItems([]orderedmap.MapItem{{Key: "a", Value: 123}})},
		{Key: "list", Value: []interface{}{"a", 123, orderedmap.NewMapWithItems([]orderedmap.MapItem{{Key: "a", Value: 123}})}},
	})
}
