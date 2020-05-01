package template_test

import (
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

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

func TestDataValuesWithFlags(t *testing.T) {
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

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		// TODO env and files?
		KVsFromYAML: []string{"int=124", "boolean=true", "nested.value=\"str\"", "another.nested.map=567"},
	}

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
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<< vs >>>%s<<<", file.Bytes(), expectedYAMLTplData)
	}
}

func TestDataValuesWithFlagsMarkedMissingOk(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

	expectedYAMLTplData := `values:
  nested:
    value: str
  another_nested:
    other_value: str2
`

	yamlData := []byte(`
#@data/values
---
nested:
  value: str
`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		// TODO add nested.value2*=str2 since replace with 0 nodes does not do anything
		KVsFromYAML: []string{"another_nested+.other_value=str2"},
	}

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
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<< vs >>>%s<<<", file.Bytes(), expectedYAMLTplData)
	}
}

func TestDataValuesWithLibraryAttachedFlags(t *testing.T) {
	tplBytes := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")

--- #@ template.replace(library.get("lib", alias="inst1").eval())`)

	libTplBytes := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")

lib-val: #@ data.values.lib_val
--- #@ template.replace(library.get("nested-lib").eval())
`)

	libValuesBytes := []byte(`
#@data/values
---
lib_val: override-me
`)

	nestedLibTplBytes := []byte(`
#@ load("@ytt:data", "data")

nested-lib-val: #@ data.values.nested_lib_val
`)

	nestedLibValuesBytes := []byte(`
#@data/values
---
nested_lib_val: override-me
`)

	expectedYAMLTplData := `lib-val: test
---
nested-lib-val: passes
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libTplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/nested-lib/values.yml", nestedLibValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/nested-lib/config.yml", nestedLibTplBytes)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		// TODO env and files?
		KVsFromYAML: []string{"@~inst1:lib_val=test", "@~inst1@nested-lib:nested_lib_val=passes"},
	}

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "config.yml" {
		t.Fatalf("Expected output file to be config.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<< vs >>>%s<<<", file.Bytes(), expectedYAMLTplData)
	}
}

func TestDataValuesMultipleFiles(t *testing.T) {
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

func TestDataValuesMultipleInOneFile(t *testing.T) {
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

func TestDataValuesOverlayNewKey(t *testing.T) {
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

func TestDataValuesOverlayRemoveKey(t *testing.T) {
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

func TestDataValuesWithNonDataValuesDocsErr(t *testing.T) {
	yamlData := []byte(`
#@data/values
---
str: str
---
non-data-values-doc`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail, but was no error")
	}

	if out.Err.Error() != "Overlaying data values (in following order: data.yml): Templating file 'data.yml': Expected data values file 'data.yml' to only have data values documents" {
		t.Fatalf("Expected RunWithFiles to fail, but was '%s'", out.Err)
	}
}

func TestDataValuesWithNonDocDataValuesErr(t *testing.T) {
	yamlData := []byte(`
---
#@data/values
str: str`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("data.yml", yamlData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail, but was no error")
	}

	if out.Err.Error() != "Expected YAML document to be annotated with data/values but was *yamlmeta.MapItem" {
		t.Fatalf("Expected RunWithFiles to fail, but was '%s'", out.Err)
	}
}

func TestDataValuesOverlayChildDefaults(t *testing.T) {
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

func TestDataValuesFromEnv(t *testing.T) {
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

	ui := cmdcore.NewPlainUI(false)
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
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<< vs >>>%s<<<", file.Bytes(), expectedYAMLTplData)
	}
}
