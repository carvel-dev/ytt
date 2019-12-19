package template_test

import (
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestLibraryModule(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

	libConfig1TplData := []byte(`
lib_config_1: true`)

	libConfig2TplData := []byte(`
lib_config_2: true`)

	libConfig3TplData := []byte(`
lib_config_3: true`)

	expectedYAMLTplData := `lib_config_1: true
---
lib_config_2: true
---
lib_config_3: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.yml", libConfig1TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config2.yml", libConfig2TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/dir/config3.yml", libConfig3TplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleNested(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

	libConfig1TplData := []byte(`
lib_config_1: true`)

	libConfig2TplData := []byte(`
lib_config_2: true`)

	libConfig3TplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

lib_config_3: true
--- #@ template.replace(library.get("nested-lib").eval())`)

	nestedLibConfigTplData := []byte(`
lib_config_nested: true`)

	expectedYAMLTplData := `lib_config_1: true
---
lib_config_2: true
---
lib_config_3: true
---
lib_config_nested: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.yml", libConfig1TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config2.yml", libConfig2TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/dir/config3.yml", libConfig3TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/dir/_ytt_lib/nested-lib/config.yml", nestedLibConfigTplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithDataValues(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def additional_vals():
int: 124
#@overlay/match missing_ok=True
str: string
#@ end

#@ lib = library.get("lib")
#@ lib2 = lib.with_data_values({"int": 123})
#@ lib3 = lib.with_data_values(additional_vals())
--- #@ template.replace(lib2.eval())
--- #@ template.replace(lib3.eval())
--- #@ template.replace(lib.eval())`)

	libValuesTplData := []byte(`
#@data/values
---
int: 100`)

	libConfig1TplData := []byte(`
#@ load("@ytt:data", "data")
lib_int: #@ data.values.int`)

	libConfig2TplData := []byte(`
#@ load("@ytt:data", "data")
lib_int: #@ data.values.int
lib_vals: #@ data.values`)

	expectedYAMLTplData := `lib_int: 123
---
lib_int: 123
lib_vals:
  int: 123
---
lib_int: 124
---
lib_int: 124
lib_vals:
  int: 124
  str: string
---
lib_int: 100
---
lib_int: 100
lib_vals:
  int: 100
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.yml", libConfig1TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config2.yml", libConfig2TplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithDataValuesStruct(t *testing.T) {
	valuesTplData := []byte(`
#@data/values
---
int: 100
str: string`)

	configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ lib = library.get("lib").with_data_values(data.values)
--- #@ template.replace(lib.eval())`)

	libValuesTplData := []byte(`
#@data/values
---
int: 10
str: str`)

	libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

	expectedYAMLTplData := `vals:
  int: 100
  str: string
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithExports(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def additional_vals():
int: 124
#@overlay/match missing_ok=True
str: string
#@ end

#@ lib = library.get("lib").with_data_values(additional_vals())
vals: #@ lib.export("vals")("arg1")`)

	libValuesTplData := []byte(`
#@data/values
---
int: 100`)

	libConfigLibData := []byte(`
#@ load("@ytt:data", "data")
#@ def vals(arg1): return [arg1, data.values]`)

	expectedYAMLTplData := `vals:
- arg1
- int: 124
  str: string
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.lib.yml", libConfigLibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithExportByPath(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ lib = library.get("lib").with_data_values({"int": 124})
vals1: #@ lib.export("vals", path="config1.lib.yml")()
vals2: #@ lib.export("vals", path="config2.lib.yml")()`)

	libValuesTplData := []byte(`
#@data/values
---
int: 100`)

	libConfig1LibData := []byte(`
#@ load("@ytt:data", "data")
#@ def vals(): return data.values.int + 10`)

	libConfig2LibData := []byte(`
#@ load("@ytt:data", "data")
#@ def vals(): return data.values.int + 20`)

	expectedYAMLTplData := `vals1: 134
vals2: 144
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.lib.yml", libConfig1LibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config2.lib.yml", libConfig2LibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithExportConflicts(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:library", "library")
#@ library.get("lib").export("vals")`)

	libConfig1LibData := []byte(`
#@ load("@ytt:data", "data")
#@ def vals(): return 10`)

	libConfig2LibData := []byte(`
#@ load("@ytt:data", "data")
#@ def vals(): return 20`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.lib.yml", libConfig1LibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config2.lib.yml", libConfig2LibData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail, but was no error")
	}

	expectedErr := `
- library.export: Expected to find exactly one exported symbol 'vals', but found multiple across files: config1.lib.yml, config2.lib.yml
    in <toplevel>
      config.yml:3 | #@ library.get("lib").export("vals")`

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected RunWithFiles to fail, but was '%s'", out.Err)
	}
}

func TestLibraryModuleWithExportPrivate(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:library", "library")
#@ library.get("lib").export("_vals")`)

	libConfig1LibData := []byte(`
#@ load("@ytt:data", "data")
#@ def _vals(): return 10`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.lib.yml", libConfig1LibData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail, but was no error")
	}

	expectedErr := `
- library.export: Symbols starting with '_' are private, and cannot be exported
    in <toplevel>
      config.yml:3 | #@ library.get("lib").export("_vals")`

	if out.Err.Error() != expectedErr {
		t.Fatalf("Expected RunWithFiles to fail, but was '%s'", out.Err)
	}
}

func TestLibraryModuleWithOverlays(t *testing.T) {
	valuesTplData := []byte(`
#@data/values
---
int: 100
str: string`)

	configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:yaml", "yaml")

---
overlay_unaffected: true
---
#@ lib = library.get("lib").with_data_values(data.values)
lib: #@ yaml.encode(lib.eval())`)

	libValuesTplData := []byte(`
#@data/values
---
int: 10
str: str`)

	libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

	libOverlay1TplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
vals:
  str: string-over`)

	libOverlay2TplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
vals:
  #@overlay/match missing_ok=True
  bool: true`)

	expectedYAMLTplData := `overlay_unaffected: true
---
lib: |
  vals:
    int: 100
    str: string-over
    bool: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/overlay1.yml", libOverlay1TplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/overlay2.yml", libOverlay2TplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func runAndCompare(t *testing.T, filesToProcess []*files.File, expectedYAMLTplData string) {
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

	if file.RelativePath() != "config.yml" {
		t.Fatalf("Expected output file to be config.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}
