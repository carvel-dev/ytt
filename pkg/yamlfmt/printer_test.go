// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlfmt_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/yamlfmt"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/difflib"
)

var (
	// Example usage:
	//   ./hack/test.sh -run TestYAMLFmt TestYAMLFmt.filetest=yield-def.yml
	selectedFileTestPath = kvArg("TestYAMLFmt.filetest")
	showErrs             = kvArg("TestYAMLFmt.errs") // eg t|...
)

func TestYAMLFmt(t *testing.T) {
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

		contents, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		pieces := strings.SplitN(string(contents), "\n+++\n\n", 2)

		if len(pieces) != 2 {
			t.Fatalf("expected file %s to include +++ separator", filePath)
		}

		resultStr, testErr := evalTemplate(t, pieces[0])
		expectedStr := pieces[1]

		if testErr == nil {
			err = expectEquals(t, resultStr, expectedStr)
		} else {
			err = testErr.TestErr()
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

	return yamlfmt.NewPrinter(nil).PrintStr(docSet), nil
}

func expectEquals(t *testing.T, resultStr, expectedStr string) error {
	if resultStr != expectedStr {
		diff := difflib.PPDiff(strings.Split(expectedStr, "\n"), strings.Split(resultStr, "\n"))
		return fmt.Errorf("Not equal; diff expected...actual:\n%v", diff)
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
