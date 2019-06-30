package tests

import (
	"github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"

	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

type Package struct {
	Location  string
	Recursive bool
}

func LoadRootLibrary(absolutePaths []string, ui files.UI, astValues interface{}, templateOpts workspace.TemplateLoaderOpts, symlinkOpts files.SymlinkAllowOpts) (*eval.Result, error) {
	filesToProcess, err := files.NewSortedFilesFromPaths(absolutePaths, symlinkOpts)
	if err != nil {
		return nil, err
	}

	rootLibrary := workspace.NewRootLibrary(filesToProcess)

	ll := workspace.NewLibraryLoader(rootLibrary, nil, ui, templateOpts)

	astValues, err = ll.Values(astValues)
	if err != nil {
		return nil, err
	}

	return ll.Eval(astValues)
}

func (p *Package) Test(t *testing.T) {
	ui := &core.PlainUI{}
	templOpts := workspace.TemplateLoaderOpts{}
	symlinkOpts := files.SymlinkAllowOpts{}
	res, err := LoadRootLibrary([]string{p.Location}, ui, nil, templOpts, symlinkOpts)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range res.Files {
		expectedLocation := path.Join(p.Location, f.RelativePath()+".out")
		expected, err := ioutil.ReadFile(expectedLocation)
		if err != nil && os.IsNotExist(err) {
			t.Logf("Missing expected file %+v, skip check", expectedLocation)
		} else if err != nil {
			t.Error(err)
		} else if !bytes.Equal(expected, f.Bytes()) {
			t.Logf("Expected %s/%s.out: %#v", p.Location, f.RelativePath(), string(expected))
			t.Logf("Actual   %s/%s:     %#v", p.Location, f.RelativePath(), string(f.Bytes()))
			t.Errorf("Expected file %+v does not match actual output:\n%s", expectedLocation, string(f.Bytes()))
		}
	}
}
