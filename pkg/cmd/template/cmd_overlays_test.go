// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
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
	expectedErr := "Overlaying (in following order: overlay1.yml, overlay2.yml): " +
		"Document on line overlay2.yml:4: Map item (key 'map') on line overlay2.yml:5: " +
		"Expected number of matched nodes to be 1, but was 0"
	require.EqualError(t, out.Err, expectedErr)
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
	expectedErr := "Overlaying (in following order: overlay.yml): " +
		"Document on line overlay.yml:4: " +
		"Expected number of matched nodes to be 1, but was 2 (lines: tpl.yml:2, tpl.yml:6)"
	require.EqualError(t, out.Err, expectedErr)
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}
