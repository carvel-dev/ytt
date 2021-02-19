// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"strings"
	"testing"

	"github.com/k14s/difflib"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
)

func TestDocumentOverlayWithNewKeyAsFunction(t *testing.T) {
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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

func TestDocumentOverlayAddAndRemoveKey(t *testing.T) {
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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

func TestDocumentOverlayAppendArrayItems(t *testing.T) {
	yamlTplData := []byte(`
- item1
`)

	yamlOverlayTplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
- item2
#@overlay/append
- item3

`)

	expectedYAMLTplData := `- item1
- item2
- item3
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay.yml", yamlOverlayTplData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
		t.Fatalf("Expected output file to have: >>>%s<<<, but was: >>>%s<<<", expectedYAMLTplData, file.Bytes())
	}
}

func TestDataValuesOverlay(t *testing.T) {
	t.Run("overlays one data value file onto another", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

		yamlData1 := []byte(`
#@data/values
---
int: 123
str: str`)

		yamlData2 := []byte(`
#@data/values
---
str: str2`)

		expectedYAMLTplData := `data_int: 123
data_str: str2
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("overlays when two data values documents defined in same file", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

		yamlData := []byte(`
#@data/values
---
str: str

#@data/values
---
str: str2
#@overlay/match missing_ok=True
int: 123`)

		expectedYAMLTplData := `data_int: 123
data_str: str2
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("overlays data value onto empty data value", func(t *testing.T) {
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

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("empty data values and empty Document overlay results in empty output", func(t *testing.T) {
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

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("appends array items", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

		yamlData1 := []byte(`
#@data/values
---
int:
- 123
str:
- str1`)

		yamlData2 := []byte(`
#@data/values
---
int: 
- 456
str:
- str2`)

		expectedYAMLTplData := `data_int:
- 123
- 456
data_str:
- str1
- str2
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		opts := cmdtpl.NewOptions()

		assertYTTOverlaySucceedsWithOutput(t, filesToProcess, opts, expectedYAMLTplData)
	})
	t.Run("appends array items when base data values is empty array", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
containers: #@ data.values.containers`)

		yamlData1 := []byte(`
#@data/values
---
containers: []
`)

		yamlData2 := []byte(`
#@data/values
---
containers:
- name: app
- name: app2
`)

		expectedYAMLTplData := `containers:
- name: app
- name: app2
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		opts := cmdtpl.NewOptions()

		assertYTTOverlaySucceedsWithOutput(t, filesToProcess, opts, expectedYAMLTplData)
	})
	t.Run("appends array items when overlay is empty array", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
containers: #@ data.values.containers`)

		yamlData1 := []byte(`
#@data/values
---
containers: 
- name: app
- name: app2
`)

		yamlData2 := []byte(`
#@data/values
---
containers: []
`)

		expectedYAMLTplData := `containers:
- name: app
- name: app2
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		opts := cmdtpl.NewOptions()

		assertYTTOverlaySucceedsWithOutput(t, filesToProcess, opts, expectedYAMLTplData)
	})
	t.Run("'match-child-defaults' annotation on array item sets default behavior for children", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
metadata: #@ data.values.metadata`)

		yamlData1 := []byte(`
#@data/values
---
metadata:
  - annotations:
      ingress.kubernetes.io/rewrite-target: true`)

		yamlData2 := []byte(`
#@data/values
#@ load("@ytt:overlay", "overlay")
---
metadata:
#@overlay/match-child-defaults missing_ok=True
#@overlay/match by=overlay.all
- annotations:
    nginx.ingress.kubernetes.io/limit-rps: 2000
    nginx.ingress.kubernetes.io/enable-access-log: "true"`)

		expectedYAMLTplData := `metadata:
- annotations:
    ingress.kubernetes.io/rewrite-target: true
    nginx.ingress.kubernetes.io/limit-rps: 2000
    nginx.ingress.kubernetes.io/enable-access-log: "true"
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("allows new key annotated with 'missing_ok=True'", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

		yamlData1 := []byte(`
#@data/values
---
str: str`)

		yamlData2 := []byte(`
#@data/values
---
str: str2
#@overlay/match missing_ok=True
int: 123`)

		expectedYAMLTplData := `data_int: 123
data_str: str2
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
	t.Run("removes key annotated with 'remove'", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data: #@ data.values.data`)

		yamlData1 := []byte(`
#@data/values
---
data:
  str: str
  int: 123`)

		yamlData2 := []byte(`
#@data/values
---
data:
  #@overlay/remove
  int: null`)

		expectedYAMLTplData := `data:
  str: str
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
			files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
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
	})
}

func assertYTTOverlaySucceedsWithOutput(t *testing.T, filesToProcess []*files.File, opts *cmdtpl.Options, expectedOut string) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Errorf("Expected number of output files to be 1, but was: %d", len(out.Files))
	}

	if string(out.Files[0].Bytes()) != expectedOut {
		diff := difflib.PPDiff(strings.Split(string(out.Files[0].Bytes()), "\n"), strings.Split(expectedOut, "\n"))
		t.Errorf("Expected output to match expected YAML, differences:\n%s", diff)
	}
}
