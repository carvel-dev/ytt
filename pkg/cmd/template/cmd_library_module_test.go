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

func TestLibraryModuleWithDataValuesPlainMerge(t *testing.T) {
	// see also: ./cmd_data_values_file_test.go for coverage of "plain merge"

	t.Run("from a struct", func(t *testing.T) {
		libValuesTplData := []byte(`
#@data/values
---
from_lib: lib/values.yml
overridden: lib/values.yml
`)

		libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

		valuesTplData := []byte(`
#@data/values
---
overridden: values.yml
from_root: values.yml`)

		configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("lib").with_data_values(data.values, plain=True).eval())`)

		expectedYAMLTplData := `vals:
  from_lib: lib/values.yml
  overridden: values.yml
  from_root: values.yml
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
		})

		runAndCompare(t, filesToProcess, expectedYAMLTplData)
	})
	t.Run("from a dictionary", func(t *testing.T) {
		libValuesTplData := []byte(`
#@data/values
---
from_lib: lib/values.yml
overridden: lib/values.yml
`)

		libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

		configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ dictionary = {
#@   "overridden": "dictionary",
#@   "from_dictionary": "dictionary"
#@ }

--- #@ template.replace(library.get("lib").with_data_values(dictionary, plain=True).eval())`)

		expectedYAMLTplData := `vals:
  from_lib: lib/values.yml
  overridden: dictionary
  from_dictionary: dictionary
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
		})

		runAndCompare(t, filesToProcess, expectedYAMLTplData)

	})
	t.Run("from a YAML Fragment (containing plain YAML)", func(t *testing.T) {
		libValuesTplData := []byte(`
#@data/values
---
from_lib: lib/values.yml
overridden: lib/values.yml
`)

		libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

		configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def yamlfragment():
overridden: yamlfragment
from_yamlfragment: yamlfragment
#@ end

--- #@ template.replace(library.get("lib").with_data_values(yamlfragment(), plain=True).eval())`)

		expectedYAMLTplData := `vals:
  from_lib: lib/values.yml
  overridden: yamlfragment
  from_yamlfragment: yamlfragment
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
		})

		runAndCompare(t, filesToProcess, expectedYAMLTplData)
	})
	t.Run("from a YAML Fragment (errors because annotations are present)", func(t *testing.T) {
		libValuesTplData := []byte(`
#@data/values
---
from_lib: lib/values.yml
overridden: lib/values.yml
`)

		libConfigTplData := []byte(`
#@ load("@ytt:data", "data")
vals: #@ data.values`)

		configTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def yamlfragment():
overridden: yamlfragment
#@overlay/match missing_ok=True
from_yamlfragment: yamlfragment
#@ end

--- #@ template.replace(library.get("lib").with_data_values(yamlfragment(), plain=True).eval())`)

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigTplData)),
		})
		assertFails(t, filesToProcess, "Expected to be plain YAML", cmdtpl.NewOptions())
	})
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	expectedErr := `
- library.export: Exporting from library 'lib': Expected to find exactly one exported symbol 'vals', but found multiple across files: config1.lib.yml, config2.lib.yml
    in <toplevel>
      config.yml:3 | #@ library.get("lib").export("vals")`
	require.EqualError(t, out.Err, expectedErr)
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

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	expectedErr := `
- library.export: Exporting from library 'lib': Symbols starting with '_' are private, and cannot be exported
    in <toplevel>
      config.yml:3 | #@ library.get("lib").export("_vals")`
	require.EqualError(t, out.Err, expectedErr)
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

func TestLibDVsRefsWithPathNoAlias(t *testing.T) {
	tmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

top_level_value: #@ data.values.top_level_val
--- #@ template.replace(library.get("lib1").eval())
--- #@ template.replace(library.get("with-child-lib").eval())`)

	dataValuesBytes := []byte(`#@data/values
---
top_level_val: top_level_value

#@data/values
#@library/ref "@lib1"
---
lib1_val2: lib_val_override

#@data/values
#@library/ref "@with-child-lib@child-lib"
---
child_val2: child_val_override`)

	lib1TmplBytes := []byte(`#@ load("@ytt:data", "data")
lib1_val1: #@ data.values.lib1_val1
lib1_val2: #@ data.values.lib1_val2`)

	lib1DataValuesBytes := []byte(`
#@data/values
---
lib1_val1: from_lib1_dvs
lib1_val2: override_me`)

	libWithChildTmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-lib").eval())`)

	childLibTmplBytes := []byte(`#@ load("@ytt:data", "data")
child_val1: #@ data.values.child_val1
child_val2: #@ data.values.child_val2`)

	childLibDataValuesBytes := []byte(`#@data/values
---
child_val1: from_child_vals
child_val2: override_me`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/values.yml", lib1DataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/config.yml", lib1TmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/config.yml", libWithChildTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/values.yml", childLibDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/config.yml", childLibTmplBytes)),
	})

	expectedYAML := `top_level_value: top_level_value
---
lib1_val1: from_lib1_dvs
lib1_val2: lib_val_override
---
child_val1: from_child_vals
child_val2: child_val_override
`

	runAndCompare(t, filesToProcess, expectedYAML)
}

func TestLibDVsRefsWithAliasNoPath(t *testing.T) {
	tmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

top_level_value: #@ data.values.top_level_val
--- #@ template.replace(library.get("lib1", alias="no-nesting").eval())
--- #@ template.replace(library.get("with-child-lib", alias="nesting").eval())`)

	dataValuesBytes := []byte(`#@data/values
---
top_level_val: top_level_value

#@data/values
#@library/ref "@~no-nesting"
---
lib1_val2: lib_val_override

#@data/values
#@library/ref "@~nesting@~child"
---
child_val2: child_val_override`)

	lib1TmplBytes := []byte(`#@ load("@ytt:data", "data")
lib1_val1: #@ data.values.lib1_val1
lib1_val2: #@ data.values.lib1_val2`)

	lib1DataValuesBytes := []byte(`#@data/values
---
lib1_val1: from_lib1_dvs
lib1_val2: override_me`)

	libWithChildTmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-lib", alias="child").eval())`)

	childLibTmplBytes := []byte(`#@ load("@ytt:data", "data")
child_val1: #@ data.values.child_val1
child_val2: #@ data.values.child_val2`)

	childLibDataValuesBytes := []byte(`#@data/values
---
child_val1: from_child_vals
child_val2: override_me`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/values.yml", lib1DataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/config.yml", lib1TmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/config.yml", libWithChildTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/values.yml", childLibDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/config.yml", childLibTmplBytes)),
	})

	expectedYAML := `top_level_value: top_level_value
---
lib1_val1: from_lib1_dvs
lib1_val2: lib_val_override
---
child_val1: from_child_vals
child_val2: child_val_override
`

	runAndCompare(t, filesToProcess, expectedYAML)
}

func TestLibDVsRefsWithPathAndAlias(t *testing.T) {
	tmplBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

top_level_value: #@ data.values.top_level_val
--- #@ template.replace(library.get("lib1", alias="no-nesting").eval())
--- #@ template.replace(library.get("with-child-lib", alias="nesting").eval())`)

	dataValuesBytes := []byte(`#@data/values
---
top_level_val: top_level_value

#@data/values
#@library/ref "@lib1~no-nesting"
---
lib1_val2: lib_val_override

#@data/values
#@library/ref "@with-child-lib~nesting@child-lib~child"
---
child_val2: child_val_override
`)

	lib1TmplBytes := []byte(`
#@ load("@ytt:data", "data")
lib1_val1: #@ data.values.lib1_val1
lib1_val2: #@ data.values.lib1_val2`)

	lib1DataValuesBytes := []byte(`
#@data/values
---
lib1_val1: from_lib1_dvs
lib1_val2: override_me`)

	libWithChildTmplBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-lib", alias="child").eval())`)

	childLibTmplBytes := []byte(`
#@ load("@ytt:data", "data")
child_val1: #@ data.values.child_val1
child_val2: #@ data.values.child_val2`)

	childLibDataValuesBytes := []byte(`
#@data/values
---
child_val1: from_child_vals
child_val2: override_me`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/values.yml", lib1DataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib1/config.yml", lib1TmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/config.yml", libWithChildTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/values.yml", childLibDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/config.yml", childLibTmplBytes)),
	})

	expectedYAML := `top_level_value: top_level_value
---
lib1_val1: from_lib1_dvs
lib1_val2: lib_val_override
---
child_val1: from_child_vals
child_val2: child_val_override
`

	runAndCompare(t, filesToProcess, expectedYAML)
}

func TestLibDVsParentAndGrandparentOrdering(t *testing.T) {
	tmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("with-child-lib").eval())`)

	dataValuesBytes := []byte(`#@data/values
#@library/ref "@with-child-lib@child-lib"
---
child_val: grandparent_value`)

	libWithChildTmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-lib").eval())`)

	libWithChildDataValuesBytes := []byte(`#@data/values
#@library/ref "@child-lib"
---
child_val: parent_value`)

	childLibTmplBytes := []byte(`#@ load("@ytt:data", "data")
child_val: #@ data.values.child_val`)

	childLibDataValuesBytes := []byte(`#@data/values
---
child_val: child_val`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/values.yml", libWithChildDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/config.yml", libWithChildTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/values.yml", childLibDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/config.yml", childLibTmplBytes)),
	})

	expectedYAML := `child_val: grandparent_value
`

	runAndCompare(t, filesToProcess, expectedYAML)
}

// Test one with a mix of all 3
func TestLibDVsComboRefWithNesting(t *testing.T) {
	tmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("with-child-lib").eval())`)

	dataValuesBytes := []byte(`#@data/values
#@library/ref "@with-child-lib@child-lib~child@~child-child"
---
child_child: great_grandparent_value`)

	libWithChildTmplBytes := []byte(`#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-lib", alias="child").eval())`)

	childLibTmplBytes := []byte(`#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child-child-lib", alias="child-child").eval())`)

	childChildTmplBytes := []byte(`#@ load("@ytt:data", "data")
child_child_val: #@ data.values.child_child`)

	childChildDataValuesBytes := []byte(`#@data/values
---
child_child: override-me`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", tmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/config.yml", libWithChildTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/config.yml", childLibTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/_ytt_lib/child-child-lib/values.yml", childChildDataValuesBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-child-lib/_ytt_lib/child-lib/_ytt_lib/child-child-lib/config.yml", childChildTmplBytes)),
	})

	expectedYAML := `child_child_val: great_grandparent_value
`

	runAndCompare(t, filesToProcess, expectedYAML)
}

func TestLibDVsAfterLibModule(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def dv1():
lib_val1: "foo"
lib_val2: "bar"
#@ end

--- #@ template.replace(library.get("lib").with_data_values(dv1()).eval())`)

	dataValueBytes := []byte(`
#@library/ref "@lib"
#@data/values
---
lib_val1: val1

#@library/ref "@lib"
#@data/values after_library_module=True
---
lib_val2: val2`)

	libDVBytes := []byte(`
#@data/values
---
lib_val1: "unused"
lib_val2: "unused"`)

	libConfigBytes := []byte(`
#@ load("@ytt:data", "data")

lib_val1: #@ data.values.lib_val1
lib_val2: #@ data.values.lib_val2`)

	expectedYAMLTplData := `lib_val1: foo
lib_val2: val2
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libDVBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigBytes)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

// Test multiple .with_data_values calls -> ensure we dont make any lasting changes
func TestLibDVsNoInstancePollution(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ def dv1():
lib_val1: "foo"
#@ end

#@ def dv2():
lib_val2: "bar"
#@ end

--- #@ template.replace(library.get("lib", alias="inst1").with_data_values(dv1()).with_data_values(dv2()).eval())
`)

	dataValueBytes := []byte(`
#@library/ref "@lib~inst1"
#@data/values after_library_module=True
---
lib_val1: val1`)

	libDVBytes := []byte(`
#@data/values
---
lib_val1: "unchanged1"
lib_val2: "unchanged2"`)

	libConfigBytes := []byte(`
#@ load("@ytt:data", "data")

lib_val1: #@ data.values.lib_val1
lib_val2: #@ data.values.lib_val2`)

	expectedYAMLTplData := `lib_val1: val1
lib_val2: bar
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libDVBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigBytes)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestUnusedLibraryDataValues(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

--- #@ template.replace(library.get("lib", alias="inst1").eval())`)

	dataValueBytes := []byte(`
#@library/ref "@~inst1"
#@data/values
---
lib_val1: val1

#@library/ref "@~inst2"
#@data/values
---
lib_val2: val2`)

	libDVBytes := []byte(`
#@data/values
---
lib_val1: "library-defined"
lib_val2: "library-defined"`)

	libConfigBytes := []byte(`
#@ load("@ytt:data", "data")

lib_val1: #@ data.values.lib_val1
lib_val2: #@ data.values.lib_val2`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libDVBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)

	expectedErr := "Expected all provided library data values documents to be used " +
		"but found unused: Data Value belonging to library '@~inst2' on line values.yml:9"
	require.EqualError(t, out.Err, expectedErr)
}

func TestUnusedLibraryDataValuesNested(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

--- #@ template.replace(library.get("with-nested-lib", alias="inst1").eval())`)

	dataValueBytes := []byte(`
#@library/ref "@~inst1@~inst2"
#@data/values
---
nested_lib_val1: new-val1`)

	withNestedLibTmplBytes := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")

--- #@ template.replace(library.get("lib", alias="inst1").eval())`)

	nestedLibTmplBytes := []byte(`
#@ load("@ytt:data", "data")

nested-lib: #@ data.values.nested_lib_val1`)

	nestedLibDVBytes := []byte(`
#@data/values
---
nested_lib_val1: override-me`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-nested-lib/config.yml", withNestedLibTmplBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-nested-lib/_ytt_lib/lib/values.yml", nestedLibDVBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-nested-lib/_ytt_lib/lib/config.yml", nestedLibTmplBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "Expected all provided library data values documents to be used "+
		"but found unused: Data Value belonging to library '@~inst1@~inst2' on line values.yml:4")
}

func TestUnusedLibraryDataValuesWithoutLibraryEvalChild(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:library", "library")
#@ library.get("with-nested-lib")`)

	dataValueBytes := []byte(`
#@library/ref "@with-nested-lib@lib"
#@data/values
---
nested_lib_val1: new-val1`)

	withNestedLibTmplBytes := []byte(``)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-nested-lib/config.yml", withNestedLibTmplBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "Expected all provided library data values documents to be used "+
		"but found unused: Data Value belonging to library '@with-nested-lib@lib' on line values.yml:4")
}

func TestUnusedLibraryDataValuesNestedWithoutLibraryEval(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:library", "library")
#@ library.get("with-nested-lib")`)

	dataValueBytes := []byte(`
#@library/ref "@with-nested-lib"
#@data/values
---
nested_lib_val1: new-val1`)

	withNestedLibTmplBytes := []byte(``)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/with-nested-lib/config.yml", withNestedLibTmplBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "Expected all provided library data values documents to be used "+
		"but found unused: Data Value belonging to library '@with-nested-lib' on line values.yml:4")
}

func TestMalformedLibraryRefEmptyAlias(t *testing.T) {
	dataValueBytes := []byte(`
#@library/ref "@lib~"
#@data/values
---
lib_val1: val1
`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "Expected library alias to not be empty")
}

func TestMalformedLibraryRefGeneralError(t *testing.T) {
	dataValueBytes := []byte(`
#@library/ref "@~123~abc~inst1"
#@data/values
---
lib_val1: val1
`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "Expected library ref to have form: '@path', '@~alias', or '@path~alias', got: ")
}

func TestLibraryModuleDataValuesFunc(t *testing.T) {
	configBytes := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
#@ load("@ytt:data", "data")

#@ lib_vals = library.get("lib", alias="inst1").data_values()

lib_val1: #@ lib_vals.lib_val1
lib_val2: #@ lib_vals.lib_val2
`)

	dataValueBytes := []byte(`
#@library/ref "@~inst1"
#@data/values
---
lib_val1: val1
`)

	libDVBytes := []byte(`
#@data/values
---
lib_val1: "library-defined"
lib_val2: "library-defined"`)

	expectedYAMLTplData := `lib_val1: val1
lib_val2: library-defined
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", dataValueBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configBytes)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libDVBytes)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithOpts(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ lib = library.get("lib", ignore_unknown_comments=True)
#@ lib2 = lib.with_data_values({"int": 123})
--- #@ template.replace(lib2.eval())
--- #@ template.replace(lib.eval())`)

	libValuesTplData := []byte(`
#@data/values
---
int: 100`)

	libConfig1TplData := []byte(`
#@ load("@ytt:data", "data")
# ignored!
lib_int: #@ data.values.int`)

	expectedYAMLTplData := `lib_int: 123
---
lib_int: 100
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.yml", libConfig1TplData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLibraryModuleWithoutOpts(t *testing.T) {
	configTplData := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

#@ lib = library.get("lib")
#@ lib2 = lib.with_data_values({"int": 123})
--- #@ template.replace(lib2.eval())
--- #@ template.replace(lib.eval())`)

	libValuesTplData := []byte(`
#@data/values
---
int: 100`)

	libConfig1TplData := []byte(`
#@ load("@ytt:data", "data")
# ignored!
lib_int: #@ data.values.int`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libValuesTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config1.yml", libConfig1TplData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.Error(t, out.Err)
	require.Contains(t, out.Err.Error(), "'# ignored!': Expected ytt-formatted string")
}

func runAndCompare(t *testing.T, filesToProcess []*files.File, expectedYAMLTplData string) {
	t.Helper()
	runAndCompareWithOpts(t, cmdtpl.NewOptions(), filesToProcess, expectedYAMLTplData)
}

func runAndCompareWithOpts(t *testing.T, opts *cmdtpl.Options,
	filesToProcess []*files.File, expectedYAMLTplData string) {
	t.Helper()

	ui := ui.NewTTY(false)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	file := out.Files[0]

	assert.Equal(t, "config.yml", file.RelativePath())
	assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
}
