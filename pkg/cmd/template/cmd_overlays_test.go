// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

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

func TestDocumentOverlays2(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlOverlayTplData1 := []byte(`
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

	yamlOverlayTplData2 := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/remove
  subarray2:
`)

	// subarray2 is not present because it was removed
	// by overlay1 that comes after overlay2
	expectedYAMLTplData := `array:
- name: item1
  subarray:
  - item1
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		// Note that overlay1 is alphanumerically before overlay2
		// but sorting of the files puts them in overlay2 then overlay1 order
		files.MustNewFileFromSource(files.NewBytesSource("overlay2.yml", yamlOverlayTplData1)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay1.yml", yamlOverlayTplData2)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
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

func TestDocumentOverlayDescriptiveError(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlOverlay1TplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/match missing_ok=True
  subarray2: 2
`)

	yamlOverlay2TplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
map: {}
`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay1.yml", yamlOverlay1TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay2.yml", yamlOverlay2TplData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to error")
	}

	expectedErr := "Overlaying (in following order: overlay1.yml, overlay2.yml): " +
		"Document on line overlay2.yml:4: Map item (key 'map') on line overlay2.yml:5: " +
		"Expected number of matched nodes to be 1, but was 0"

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected error to match string but was '%s'", out.Err.Error())
	}
}

func TestDocumentOverlayMultipleMatchesDescriptiveError(t *testing.T) {
	yamlTplData := []byte(`
---
id: 1
name: foo

---
id: 2
name: foo
`)

	yamlOverlayTplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
overlayed: true
`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay.yml", yamlOverlayTplData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to error")
	}

	expectedErr := "Overlaying (in following order: overlay.yml): " +
		"Document on line overlay.yml:4: " +
		"Expected number of matched nodes to be 1, but was 2 (lines: tpl.yml:2, tpl.yml:6)"

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected error to match '%s' but was '%s'", expectedErr, out.Err.Error())
	}
}

func TestMultipleDataValuesOneEmptyAndOneNonEmpty(t *testing.T) {
	yamlTplData := []byte(`#@ load("@ytt:data", "data")
key: #@ data.values.key
`)

	emptyYamlDataValue := []byte(`
#@data/values
---
`)

	nonEmptyYamlDataValue := []byte(`#@ load("@ytt:overlay", "overlay")
#@data/values
#@overlay/replace
---
key: value_from_data_value_overlay
`)

	expectedYAMLTplData := `key: value_from_data_value_overlay
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("datavalue_empty.yml", emptyYamlDataValue)),
		files.MustNewFileFromSource(files.NewBytesSource("datavalue_non_empty.yml", nonEmptyYamlDataValue)),
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

func TestEmptyOverlayAndEmptyDataValues(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlDataValue := []byte(`
#@data/values
---
`)

	yamlOverlayTplData1 := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
`)

	expectedYAMLTplData := ``

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay2.yml", yamlOverlayTplData1)),
		files.MustNewFileFromSource(files.NewBytesSource("datavalue.yml", yamlDataValue)),
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
