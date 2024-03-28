// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"fmt"
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyDataValues(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values`)

	expectedYAMLTplData := `data_int: {}
`
	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
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

func TestDataValues(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	expectedYAMLTplData := `data_int: 123
data_str: str
`

	yamlData := []byte(`
#@data/values
---
int: 123
str: str`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
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

func TestDataValues_with_flags(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str
values: #@ data.values`)

	expectedYAMLTplData := `data_int: 124
data_str: str
values:
  int: 124
  str: str
  boolean: true
  nested:
    value: str
  another:
    nested:
      map: 567
`

	yamlData := []byte(`
#@data/values
---
int: 123
str: str
boolean: false
nested:
  value: not-str
another:
  nested:
    map: {"a": 123}`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		// TODO env and files?
		KVsFromYAML: []string{"int=124", "boolean=true", "nested.value=\"str\"", "another.nested.map=567"},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}

func TestDataValues_with_flags_without_data_values_overlay(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str
values: #@ data.values`)

	expectedYAMLTplData := `data_int: 124
data_str: str
values:
  int: 124
  another:
    nested:
      map: 567
  str: str
  boolean: true
  nested:
    value: str
`

	// Only some values are prespecified by the overlay
	yamlData := []byte(`
#@data/values
---
int: 123
another:
  nested:
    map: {"a": 123}`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		EnvFromStrings: []string{"DVS"},
		EnvironFunc:    func() []string { return []string{"DVS_str=str"} },
		KVsFromYAML:    []string{"int=124", "boolean=true", "nested.value=\"str\"", "another.nested.map=567"},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}

func TestDataValues_with_library_attached_flags(t *testing.T) {
	tplBytes := `
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")

root: #@ data.values
--- #@ template.replace(library.get("lib", alias="inst1").eval())`

	libTplBytes := `
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")

from_library: #@ data.values
--- #@ template.replace(library.get("nested-lib").eval())
`

	libValuesBytes := `
#@data/values
---
val0: override-me
`

	nestedLibTplBytes := `
#@ load("@ytt:data", "data")

from_nested_lib: #@ data.values
`

	nestedLibValuesBytes := `
#@data/values
---
val1: override-me
`

	dvs2 := `val2: 2`

	dvs3 := `val3: 3`

	dvs4 := `val4: 4`

	dvs6 := `6`

	expectedYAMLTplData := `root:
  val2: 2
---
from_library:
  val0: 0
  val3: 3
  val4: 4
  val5: "5"
  val6: "6"
---
from_nested_lib:
  val1: 1
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(tplBytes))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", []byte(libValuesBytes))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libTplBytes))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/nested-lib/values.yml", []byte(nestedLibValuesBytes))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/nested-lib/config.yml", []byte(nestedLibTplBytes))),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		KVsFromYAML:    []string{"@~inst1:val0=0", "@~inst1@nested-lib:val1=1"},
		FromFiles:      []string{"c:\\User\\user\\dvs2.yml", "@~inst1:dvs3.yml", "@lib:D:\\User\\user\\dvs4.yml"},
		EnvFromStrings: []string{"@lib:DVS"},
		EnvironFunc:    func() []string { return []string{"DVS_val5=5"} },
		KVsFromFiles:   []string{"@lib:val6=c:\\User\\user\\dvs6.yml"},
		ReadFilesFunc: func(path string) ([]*files.File, error) {
			switch path {
			case "c:\\User\\user\\dvs2.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs2.yml", []byte(dvs2)))}, nil
			case "dvs3.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs3.yml", []byte(dvs3)))}, nil
			case "D:\\User\\user\\dvs4.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs4.yml", []byte(dvs4)))}, nil
			case "c:\\User\\user\\dvs6.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs6.yml", []byte(dvs6)))}, nil
			default:
				return nil, fmt.Errorf("Unknown file '%s'", path)
			}
		},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "config.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}

func TestDataValues_multiple_files(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	expectedYAMLTplData := `data_int: 123
data_str: str2
`

	yamlData1 := []byte(`
#@data/values
---
int: 123
str: str`)

	yamlData2 := []byte(`
#@data/values
---
str: str2`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
		files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
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

func TestDataValues_multiple_in_one_file(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	expectedYAMLTplData := `data_int: 123
data_str: str2
`

	yamlData := []byte(`
#@data/values
---
str: str

#@data/values
---
str: str2
#@overlay/match missing_ok=True
int: 123`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
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

func TestDataValues_overlay_new_key(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	expectedYAMLTplData := `data_int: 123
data_str: str2
`

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

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
		files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
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

func TestDataValues_overlay_remove_key(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data: #@ data.values.data`)

	expectedYAMLTplData := `data:
  str: str
`

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

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
		files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
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

func TestDataValues_with_non_data_values_docs_err(t *testing.T) {
	yamlData := []byte(`
#@data/values
---
str: str
---
non-data-values-doc`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	expectedError := "Overlaying data values (in following order: data.yml): Templating file 'data.yml': Expected data values file 'data.yml' to only have data values documents"
	require.EqualError(t, out.Err, expectedError)
}

func TestDataValues_with_non_doc_data_values_err(t *testing.T) {
	yamlData := []byte(`
---
#@data/values
str: str`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	expectedErrorMessage := "Found @data/values on map item (data.yml:4); only documents (---) can be annotated with @data/values"

	require.EqualError(t, out.Err, expectedErrorMessage)
}

func TestDataValues_overlay_child_defaults(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data: #@ data.values.data`)

	expectedYAMLTplData := `data:
  str: str
  int: 123
  bool: true
`

	yamlData1 := []byte(`
#@data/values
---
data:
  str: str
  int: 123`)

	yamlData2 := []byte(`
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
data:
  bool: true`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data1.yml", yamlData1)),
		files.MustNewFileFromSource(files.NewBytesSource("data2.yml", yamlData2)),
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

func TestDataValues_data_list_relative_to_root(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
_: #@ template.replace(data.values)
`)

	expectedYAMLTplData := `Files_in_root_schema:
  /config.yml: /config.yml
  /other: /other
  /schema.yml: /schema.yml
  /values.yml: /values.yml
Files_in_schema:
  config.yml: config.yml
  other: other
  schema.yml: schema.yml
  values.yml: values.yml
Files_in_root_values:
- name: /config.yml
- name: /other
- name: /schema.yml
- name: /values.yml
Files_in_values:
- name: config.yml
- name: other
- name: schema.yml
- name: values.yml
`

	yamlSchemaData := []byte(`
#@ load("@ytt:yaml", "yaml")
#@ load("@ytt:data", "data")
#@data/values-schema
---

#@ rootFiles = data.list("/")
Files_in_root_schema:
    #@ for/end file in rootFiles:
    #@yaml/text-templated-strings
    (@= file @): #@ file
#@ files = data.list("")
Files_in_schema:
    #@ for/end file in files:
    #@yaml/text-templated-strings
    (@= file @): #@ file
Files_in_root_values:
- name: ""
Files_in_values:
- name: ""
`)

	yamlDataValuesData := []byte(`
#@data/values
---

#@ load("@ytt:yaml", "yaml")
#@ load("@ytt:data", "data")

#@ rootFiles = data.list("/")
Files_in_root_values:
    #@ for/end file in rootFiles:
    - name: #@ file
#@ files = data.list("")
Files_in_values:
    #@ for/end file in files:
    - name: #@ file`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("other", []byte("lib1\ndata"))),
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", yamlSchemaData)),
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", yamlDataValuesData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "config.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}

func TestDataValues_data_list_relative_to_library_root(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ lib = library.get("lib")

--- #@ template.replace(lib.eval())`)

	expectedYAMLTplData := `Files_in_root_schema:
  /other: /other
  /config.yml: /config.yml
  /schema/schema.yml: /schema/schema.yml
  /values/values.yml: /values/values.yml
Files_in_schema:
  schema.yml: schema.yml
Files_in_root_values:
- name: /other
- name: /config.yml
- name: /schema/schema.yml
- name: /values/values.yml
Files_in_values:
- name: values.yml
Files_in_template:
- name: /other
- name: /config.yml
- name: /schema/schema.yml
- name: /values/values.yml
`

	yamlSchemaData := []byte(`
#@ load("@ytt:yaml", "yaml")
#@ load("@ytt:data", "data")
#@data/values-schema
---

#@ rootFiles = data.list("/")
Files_in_root_schema:
    #@ for/end file in rootFiles:
    #@yaml/text-templated-strings
    (@= file @): #@ file
#@ files = data.list("")
Files_in_schema:
    #@ for/end file in files:
    #@yaml/text-templated-strings
    (@= file @): #@ file
Files_in_root_values:
- name: ""
Files_in_values:
- name: ""
`)
	yamlLibDataValues := []byte(`#@data/values
---
#@ load("@ytt:yaml", "yaml")
#@ load("@ytt:data", "data")

#@ file = data.list("")
Files_in_values:
    #@ for/end file in file:
    - name: #@ file
#@ rootFiles = data.list("/")
Files_in_root_values:
    #@ for/end file in rootFiles:
    - name: #@ file
`)

	yamlLibConfigData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")

_: #@ template.replace(data.values)
#@ files = data.list("/")
Files_in_template:
    #@ for/end file in files:
    - name: #@ file`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/other", []byte("lib1\ndata"))),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema/schema.yml", yamlSchemaData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values/values.yml", yamlLibDataValues)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", yamlLibConfigData)),
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

func TestDataValues_from_env(t *testing.T) {
	tmplBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("lib1").eval())`)

	lib1TmplBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

lib_val: #@ data.values.lib_val
--- #@ template.replace(library.get("nested").eval())`)

	lib1DataBytes := []byte(`
#@data/values
---
lib_val: override-me`)

	nestedTmplBytes := []byte(`
#@ load("@ytt:data", "data")

nested_val: #@ data.values.nested_val
`)

	nestedDataBytes := []byte(`
#@data/values
---
nested_val: override-me-too
`)

	expectedYAMLTplData := `lib_val: lib_from_env
---
nested_val: nested_from_env
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/data.yml", lib1DataBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/tpl.yml", lib1TmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/_ytt_lib/nested/data.yml", nestedDataBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/_ytt_lib/nested/tpl.yml", nestedTmplBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		EnvFromYAML: []string{"@lib1:DVAL", "@lib1@nested:NESTED_DVAL"},
		EnvironFunc: func() []string {
			return []string{
				"DVAL_lib_val=lib_from_env",
				"NESTED_DVAL_nested_val=nested_from_env",
			}
		},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "tpl.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}

func TestDataValues_with_invalid_flags_fail(t *testing.T) {
	t.Run("when `--data-value-yaml` has a `:` in the key name", func(t *testing.T) {

		expectedErr := `Extracting data value from KV: Expected at most one library-key separator ':' in 'i:nt'`

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		opts.DataValuesFlags = cmdtpl.DataValuesFlags{
			KVsFromYAML: []string{"i:nt=124"},
		}

		out := opts.RunWithFiles(cmdtpl.Input{}, ui)
		require.Errorf(t, out.Err, expectedErr)
		require.Equal(t, expectedErr, out.Err.Error())
	})
}

func TestDataValues_validations_are_skipped_when_disabled(t *testing.T) {
	t.Run("via the --dangerous-data-values-disable-validation flag", func(t *testing.T) {
		t.Run("in the root library", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.SkipValidation = true
			opts.DataValuesFlags.Inspect = true
			dataValuesYAML := `#@ load("@ytt:assert", "assert")
#@data/values
---
#@assert/validate ("nothing is valid", lambda v: False)
foo: bar
`
			expected := `foo: bar
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("or in any evaluated private libraries, regardless", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.SkipValidation = true
			configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("lib").eval())

#! even if validations are explicitly enabled...
--- #@ template.replace(library.get("lib", dangerous_data_values_disable_validation=False).eval())
`

			libValuesYAML := `#@ load("@ytt:assert", "assert")
#@data/values
---
#@assert/validate ("nothing is valid", lambda v: False)
foo: bar
`

			libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

			expected := `values:
  foo: bar
---
values:
  foo: bar
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", []byte(libValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
	})
	t.Run("via the dangerous_data_values_disable_validation= kwarg", func(t *testing.T) {
		t.Run("in the evaluated library", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("lib", dangerous_data_values_disable_validation=True).eval())
`

			libValuesYAML := `#@ load("@ytt:assert", "assert")
#@data/values
---
#@assert/validate ("nothing is valid", lambda v: False)
foo: bar
`

			libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

			expected := `values:
  foo: bar
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", []byte(libValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("or any of its dependencies, regardless", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("foo", dangerous_data_values_disable_validation=True).eval())
`
			fooConfigYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("bar").eval())
--- #@ template.replace(library.get("bar", dangerous_data_values_disable_validation=False).eval())
`

			barValuesYAML := `#@ load("@ytt:assert", "assert")
#@data/values
---
#@assert/validate ("nothing is valid", lambda v: False)
foo: bar
`

			barConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

			expected := `values:
  foo: bar
---
values:
  foo: bar
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/foo/config.yml", []byte(fooConfigYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/foo/_ytt_lib/bar/values.yml", []byte(barValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/foo/_ytt_lib/bar/config.yml", []byte(barConfigYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
	})
}
