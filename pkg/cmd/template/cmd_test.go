// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestDataListRelativeToDir(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("funcs/funcs.lib.yml", "nested_data_list", "nested_data_read")

rootlist: #@ data.list()
rootdata: #@ data.read("funcs/data")

nestedlist: #@ nested_data_list()
nesteddata: #@ nested_data_read()`)

	expectedYAMLTplData := `rootlist:
- tpl.yml
- funcs/funcs.lib.yml
- funcs/data
rootdata: |-
  data
  data
nestedlist:
  list:
  - funcs.lib.yml
  - data
nesteddata:
  data: |-
    data
    data
`

	yamlFuncsData := []byte(`
#@ load("@ytt:data", "data")

#@ def/end nested_data_list():
list: #@ data.list()

#@ def/end nested_data_read():
data: #@ data.read("data")`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/data", []byte("data\ndata"))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/other", []byte("lib1\ndata"))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]
	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}
	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestDataListRelativeToLibraryRoot(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("funcs/funcs.lib.yml", "nested_data_list", "nested_data_read")

rootlist: #@ data.list("/")
rootdata: #@ data.read("/funcs/data")

nestedlist: #@ nested_data_list()
nesteddata: #@ nested_data_read()`)

	yamlFuncsTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("funcs.lib.yml", "nested_data_list", "nested_data_read")

rootlist: #@ data.list("/")
rootdata: #@ data.read("/funcs/data")

nestedlist: #@ nested_data_list()
nesteddata: #@ nested_data_read()
`)

	expectedYAMLTplData := `rootlist:
- /tpl.yml
- /funcs/funcs.lib.yml
- /funcs/tpl.yml
- /funcs/data
rootdata: |-
  data
  data
nestedlist:
  list:
  - /tpl.yml
  - /funcs/funcs.lib.yml
  - /funcs/tpl.yml
  - /funcs/data
nesteddata:
  data: |-
    data
    data
`

	yamlFuncsData := []byte(`
#@ load("@ytt:data", "data")

#@ def/end nested_data_list():
list: #@ data.list("/")

#@ def/end nested_data_read():
data: #@ data.read("/funcs/data")`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/tpl.yml", yamlFuncsTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/data", []byte("data\ndata"))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/other", []byte("lib1\ndata"))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 2 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]
	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}
	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}

	file = out.Files[1]
	if file.RelativePath() != "funcs/tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}
	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestDataListRelativeToLibraryRootWithinALibrary(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@lib1:funcs/funcs.lib.yml", "lib_data_list", "lib_data_read")
liblist: #@ lib_data_list()
libdata: #@ lib_data_read()`)

	expectedYAMLTplData := `liblist:
  liblist1:
  - funcs.lib.yml
  liblist2:
  - /other
  - /funcs/funcs.lib.yml
libdata:
  libdata2: |-
    lib1
    data
`

	yamlLibFuncsData := []byte(`
#@ load("@ytt:data", "data")

#@ def lib_data_list():
liblist1: #@ data.list()
liblist2: #@ data.list("/")
#@ end

#@ def lib_data_read():
libdata2: #@ data.read("/other")
#@ end`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/other", []byte("lib1\ndata"))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/funcs/funcs.lib.yml", yamlLibFuncsData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
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
    in some_data
      funcs/funcs.lib.yml:3 | #@   return 1+"2"
    in another_data
      tpl.yml:7 | #@   return some_data()
    in <toplevel>
      tpl.yml:9 | simple_key: #@ another_data()`

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
- cannot load _ytt_lib/data.lib.star: Expected to find file '_ytt_lib/data.lib.star', but did not: Encountered private library '_ytt_lib'
    in <toplevel>
      tpl.yml:1 | #@ load("_ytt_lib/data.lib.star", "data")`

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
load("@funcs:funcs.star", "libstarfunc")
def starfunc():
  return libstarfunc()
end`)

	starlarkFuncsLibData := []byte(`
def libstarfunc():
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
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/library2/_ytt_lib/funcs/funcs.star", starlarkFuncsLibData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestRelativeLoadInLibrariesForNonRootTemplates(t *testing.T) {
	expectedYAMLTplData := `libstarfunc:
- 1
- 2
`

	nonTopLevelYmlTplData := []byte(`
#@ load("@funcs:funcs.star", "libstarfunc")
libstarfunc: #@ libstarfunc()`)

	nonTopLevelStarlarkFuncsLibData := []byte(`
def libstarfunc():
  return [1,2]
end`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("non-top-level/tpl.yml", nonTopLevelYmlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("non-top-level/_ytt_lib/funcs/funcs.star", nonTopLevelStarlarkFuncsLibData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "non-top-level/tpl.yml" {
		t.Fatalf("Expected output file to be non-top-level/tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestIgnoreUnknownCommentsFalse(t *testing.T) {
	yamlTplData := []byte(`
# plain YAML comment
#@ load("funcs/funcs.lib.yml", "yamlfunc")
yamlfunc: #@ yamlfunc()`)

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	if !strings.Contains(out.Err.Error(), "Non-ytt comment at line tpl.yml:2: '# plain YAML comment': Unrecognized comment type (expected '#@' or '#!')") {
		t.Fatalf("Expected RunWithFiles to fail with error, but was '%s'", out.Err.Error())
	}
}

func TestIgnoreUnknownCommentsTrue(t *testing.T) {
	yamlTplData := []byte(`
# plain YAML comment
#@ load("funcs/funcs.lib.yml", "yamlfunc")
yamlfunc: #@ yamlfunc()`)

	expectedYAMLTplData := `yamlfunc:
  yamlfunc: yamlfunc
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.IgnoreUnknownComments = true

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestParseErrTemplateFile(t *testing.T) {
	yamlTplData := []byte(`
key: val
yamlfunc yamlfunc`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	if out.Err.Error() != "Unmarshaling YAML template 'tpl.yml': yaml: line 4: could not find expected ':'" {
		t.Fatalf("Expected RunWithFiles to fail with error, but was '%s'", out.Err.Error())
	}
}

func TestParseErrLoadFile(t *testing.T) {
	yamlTplData := []byte(`
#@ load("funcs/funcs.lib.yml", "yamlfunc")
yamlfunc: #@ yamlfunc()`)

	yamlFuncsData := []byte(`
#@ def yamlfunc():
key: val
yamlfunc yamlfunc
#@ end`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	if !strings.Contains(out.Err.Error(), "cannot load funcs/funcs.lib.yml: Unmarshaling YAML template 'funcs/funcs.lib.yml': yaml: line 5: could not find expected ':'") {
		t.Fatalf("Expected RunWithFiles to fail with error, but was '%s'", out.Err.Error())
	}
}

func TestPlainYAMLNoTemplateProcessing(t *testing.T) {
	yamlTplData := []byte(`
#@ load("funcs/funcs.lib.yml", "yamlfunc")
annotation: 5 #@ 1 + 2
text_template: (@= "string" @)`)

	expectedYAMLTplData := `annotation: 5
text_template: (@= "string" @)
`

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
	}

	filesToProcess[0].MarkTemplate(false)

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestPlainTextNoTemplateProcessing(t *testing.T) {
	txtTplData := []byte(`text (@= "string" @)`)
	expectedTxtTplData := `text (@= "string" @)`

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.txt", txtTplData)),
	}

	filesToProcess[0].MarkTemplate(false)

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.txt" {
		t.Fatalf("Expected output file to be tpl.txt, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedTxtTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestStrictInTemplate(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: yes`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.StrictYAML = true

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	expectedErr := "Unmarshaling YAML template 'tpl.yml': yaml: Strict parsing: " +
		"Found 'yes' ambigious (could be !!str or !!bool)"

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected RunWithFiles to fail with err: %s", out.Err)
	}
}

func TestStrictInDataValues(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	yamlData := []byte(`
#@data/values
---
int: 123
str: yes`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.StrictYAML = true

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	expectedErr := "Unmarshaling YAML template 'data.yml': yaml: Strict parsing: " +
		"Found 'yes' ambigious (could be !!str or !!bool)"

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected RunWithFiles to fail with err: %s", out.Err)
	}
}

func TestStrictInDataValueFlags(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	yamlData := []byte(`
#@data/values
---
int:
str: `)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.StrictYAML = true
	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		KVsFromYAML: []string{"str=yes"},
	}

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	expectedErr := "Extracting data value from KV: Deserializing value for key 'str': " +
		"Deserializing YAML value: yaml: Strict parsing: Found 'yes' ambigious (could be !!str or !!bool)"

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected RunWithFiles to fail with err: %s", out.Err)
	}
}

func TestLoadNestedYttLib(t *testing.T) {
	configTplData := []byte(`
#@ load("@dir:dir2/config.lib.yml", "func1")
func1: #@ func1()`)

	dir2ConfigLibData := []byte(`
#@ load("@dir3:funcs.lib.yml", "func2")
#@ def func1():
func2: #@ func2()
#@ end`)

	dir3FuncsLibData := []byte(`
#@ def func2():
func2: true
#@ end`)

	expectedYAMLTplData := `func1:
  func2:
    func2: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		// Note that config.lib.yml is under another directory level
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/config.lib.yml", dir2ConfigLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/_ytt_lib/dir3/funcs.lib.yml", dir3FuncsLibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLoadYTTModuleFailEarly(t *testing.T) {
	configTplData := []byte(`#@ load("@ytt:not-exist", "a")`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail")
	}

	if !strings.Contains(out.Err.Error(), "cannot load @ytt:not-exist: builtin ytt library does not have module 'not-exist'") {
		t.Fatalf("Expected RunWithFiles to fail with error, but was '%s'", out.Err.Error())
	}
}

func TestNonYAMLWithOutputFlagShowsNoWarning(t *testing.T) {
	iniData := []byte(`
[owner]
name=John Doe
organization=Acme Widgets Inc.`)
	yamlData := []byte(`foo: bar`)

	testIniFile, err := createTempFileWithContent(iniData, "ini.txt")
	if err != nil {
		t.Fatalf("Expected writing to test file not to fail: %v", err)
	}
	defer os.Remove(testIniFile.Name())

	testYamlFile, err := createTempFileWithContent(yamlData, "yml")
	if err != nil {
		t.Fatalf("Expected writing to test file not to fail: %v", err)
	}
	defer os.Remove(testYamlFile.Name())

	outputDir, err := ioutil.TempDir(os.TempDir(), "fakedir")
	if err != nil {
		t.Fatalf("Expected creating a temp dir to not fail: %v", err)
	}
	defer os.RemoveAll(outputDir)

	iniFilePath := filepath.Join(outputDir, filepath.Base(testIniFile.Name()))
	yamlFilePath := filepath.Join(outputDir, filepath.Base(testYamlFile.Name()))
	expectedStdOut := `creating: ` + iniFilePath + "\n" + `creating: ` + yamlFilePath + "\n"
	expectedStdErr := ""

	fakeStdOut, fakeStdErr, err := redirectStdOutAndErr()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	defer os.Remove(fakeStdErr.Name())
	defer os.Remove(fakeStdOut.Name())
	defer restoreStdOutAndErr()

	templateOptions := &cmdtpl.TemplateOptions{}
	cmd := cmdtpl.NewCmd(templateOptions)
	cmd.SetArgs([]string{"-f", testYamlFile.Name(), "-f", testIniFile.Name(), "--output-files", outputDir})
	cmd.SilenceUsage = true

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error running command: %v", err)
	}
	restoreStdOutAndErr()

	err = assertStdOutAndStdErr(fakeStdOut, fakeStdErr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n%v", err)
	}
}

func TestFileMarkedAsYamlFileDoesNotSuggestOutputFlags(t *testing.T) {
	expectedStdOut := "a: b\n"
	expectedStdErr := ""

	testYamlFile, err := createTempFileWithContent([]byte(expectedStdOut), "txt")
	if err != nil {
		t.Fatalf("Expected writing to test file not to fail: %v", err)
	}
	defer os.Remove(testYamlFile.Name())

	fakeStdOut, fakeStdErr, err := redirectStdOutAndErr()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	defer os.Remove(fakeStdErr.Name())
	defer os.Remove(fakeStdOut.Name())
	defer restoreStdOutAndErr()

	templateOptions := &cmdtpl.TemplateOptions{}
	cmd := cmdtpl.NewCmd(templateOptions)
	cmd.SetArgs([]string{"-f", testYamlFile.Name(), "--file-mark", fmt.Sprintf("%s:type=yaml-plain", filepath.Base(testYamlFile.Name()))})
	cmd.SilenceUsage = true

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error running command: %v", err)
	}
	restoreStdOutAndErr()

	err = assertStdOutAndStdErr(fakeStdOut, fakeStdErr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n%v", err)
	}
}

func createTempFileWithContent(yamlData []byte, suffix string) (*os.File, error) {
	testYamlFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("*test-yaml.%s", suffix))
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(testYamlFile.Name(), yamlData, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return testYamlFile, err
}
