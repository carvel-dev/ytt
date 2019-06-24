package tests

import (
	"github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/workspace"

	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

type Package struct {
	Location string
}

func (p *Package) Test(t *testing.T) {
	ui := &core.PlainUI{}
	opts := workspace.TemplateLoaderOpts{}
	res, err := workspace.LoadRootLibrary([]string{p.Location}, ui, nil, opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range res.Files {
		t.Logf("Result: %v", f.RelativePath())
		expectedLocation := path.Join(p.Location, f.RelativePath()+".out")
		expected, err := ioutil.ReadFile(expectedLocation)
		if err != nil && os.IsNotExist(err) {
			t.Errorf("Missing expected file %+v", expectedLocation)
		} else if err != nil {
			t.Error(err)
		} else if !bytes.Equal(expected, f.Bytes()) {
			t.Logf("Expected %s/%s.out: %#v", p.Location, f.RelativePath(), string(expected))
			t.Logf("Actual   %s/%s:     %#v", p.Location, f.RelativePath(), string(f.Bytes()))
			t.Errorf("Expected file %+v does not match actual output:\n%s", expectedLocation, string(f.Bytes()))
		}
	}
}
