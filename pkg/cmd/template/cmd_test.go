package template_test

import (
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestLoad(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("funcs/funcs.lib.yml", "yamlfunc")
#@ load("funcs/funcs.lib.txt", "textfunc")
#@ load("funcs/funcs.star", "starfunc")
yamlfunc: #@ yamlfunc()
textfunc: #@ textfunc()
starfunc: #@ starfunc()
listdata: #@ data.list()
loaddata: #@ data.read("funcs/funcs.star")`)

	expectedYAMLTplData := `yamlfunc:
  yamlfunc: yamlfunc
textfunc: textfunc
starfunc:
- 1
- 2
listdata:
- tpl.yml
- funcs/funcs.lib.yml
- funcs/funcs.lib.txt
- funcs/funcs.star
loaddata: |2-

  def starfunc():
    return [1,2]
  end
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	starlarkFuncsData := []byte(`
def starfunc():
  return [1,2]
end`)

	txtFuncsData := []byte(`(@ def textfunc(): @)textfunc(@ end @)`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.txt", txtFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.star", starlarkFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %#v", out.Files)
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestDocumentOverlays(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlOverlayTplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@ load("funcs/funcs.lib.yml", "yamlfunc")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/match missing_ok=True
  subarray2:
  - #@ yamlfunc()
`)

	expectedYAMLTplData := `array:
- name: item1
  subarray:
  - item1
  subarray2:
  - yamlfunc: yamlfunc
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay.yml", yamlOverlayTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %#v", out.Files)
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestBacktraceAcrossFiles(t *testing.T) {
	yamlTplData := []byte(`
#@ load("funcs/funcs.lib.yml", "some_data")
#! line
#! other line
#! another line
#@ def another_data():
#@   return some_data()
#@ end
simple_key: #@ another_data()
`)

	yamlFuncsData := []byte(`
#@ def some_data():
#@   return 1+"2"
#@ end
`)

	expectedErr := `
- unknown binary op: int + string
    funcs/funcs.lib.yml:3 in some_data
     L #@   return 1+"2"
    tpl.yml:7 in another_data
     L #@   return some_data()
    tpl.yml:9 in <toplevel>
     L simple_key: #@ another_data()`

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles fail")
	}

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected err, but was: >>>%s<<<", out.Err.Error())
	}
}

func TestDisallowDirectLibraryLoading(t *testing.T) {
	yamlTplData := []byte(`#@ load("_ytt_lib/data.lib.star", "data")`)

	expectedErr := `
- cannot load _ytt_lib/data.lib.star: Could not load file '_ytt_lib/data.lib.star' because it's contained in private library '' (use load("@lib:file", "symbol") where 'lib' is library name under _ytt_lib, for example, 'github.com/k14s/test')
    tpl.yml:1 in <toplevel>
     L #@ load("_ytt_lib/data.lib.star", "data")`

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/data.lib.star", []byte("data = 3"))),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles fail")
	}

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected err, but was: >>>%s<<<", out.Err.Error())
	}
}

func TestRelativeLoadInLibraries(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@library1:funcs.lib.yml", "yamlfunc")
#@ load("@library1:sub-dir/funcs.lib.txt", "textfunc")
#@ load("@library2:funcs.star", "starfunc")
#@ load("funcs.star", "localstarfunc")
yamlfunc: #@ yamlfunc()
textfunc: #@ textfunc()
starfunc: #@ starfunc()
localstarfunc: #@ localstarfunc()`)

	expectedYAMLTplData := `yamlfunc:
  yamlfunc: textfunc
textfunc: textfunc
starfunc:
- 1
- 2
localstarfunc:
- 3
- 4
`

	yamlFuncsData := []byte(`
#@ load("sub-dir/funcs.lib.txt", "textfunc")
#@ def/end yamlfunc():
yamlfunc: #@ textfunc()`)

	starlarkFuncsData := []byte(`
def starfunc():
  return [1,2]
end`)

	localStarlarkFuncsData := []byte(`
def localstarfunc():
  return [3,4]
end`)

	txtFuncsData := []byte(`(@ def textfunc(): @)textfunc(@ end @)`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs.star", localStarlarkFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/library1/funcs.lib.yml", yamlFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/library1/sub-dir/funcs.lib.txt", txtFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/library2/funcs.star", starlarkFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %#v", out.Files)
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}
