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

func TestDataValuesFilesFlag_acceptsPlainYAML(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

	builtinDVs := []byte(`
#@data/values
---
predefined: true`)

	// Ensure various non-annotation YAML comments
	// are allowed and do not affect parsed content
	dvs1 := []byte(`
# top comment
int: 123
str: str
boolean: false
nested:
  #comment without space
  # comment with space
  value: not-str
  ### some other unknown comment
  nested:
    #! ytt comment1
    #! ytt comment2
    subnested: true
another:
  nested:
    map: {"a": 123}
array:
- 123
- str

# bottom comment`)

	dvs2 := []byte(`
int: 123
str: str
boolean: true
nested:
  value: not-str
  nested: true
another:
  nested:
    map: {"a": 123}
# Multiple documents in one file
---
array:
- str`)

	// Ensure file with only comments (no structures) is allowed
	dvs3 := []byte(`# value: 1
# value: 2`)

	expectedYAMLTplData := `values:
  predefined: true
  int: 123
  str: str
  boolean: true
  nested:
    value: not-str
    nested: true
  another:
    nested:
      map:
        a: 123
  array:
  - str
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", builtinDVs)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		FromFiles: []string{"dvs1.yml", "dvs2.yml", "dvs3.yml"},
		ReadFilesFunc: func(path string) ([]*files.File, error) {
			switch path {
			case "dvs1.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs1.yml", dvs1))}, nil
			case "dvs2.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs2.yml", dvs2))}, nil
			case "dvs3.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs3.yml", dvs3))}, nil
			default:
				return nil, fmt.Errorf("Unknown file '%s'", path)
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

func TestDataValuesFilesFlag_rejectsTemplatedYAML(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

	dvs1 := []byte(`
#@ top comment
int: 123`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		FromFiles: []string{"dvs1.yml"},
		ReadFilesFunc: func(path string) ([]*files.File, error) {
			switch path {
			case "dvs1.yml":
				return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("dvs1.yml", dvs1))}, nil
			default:
				return nil, fmt.Errorf("Unknown file '%s'", path)
			}
		},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.EqualError(t, out.Err, "Extracting data value from file: Checking data values file 'dvs1.yml': Expected to be plain YAML, having no annotations (hint: remove comments starting with `#@`)")
}

func TestDataValuesFilesFlag_WithNonYAMLFiles(t *testing.T) {
	t.Run("errors when file is explicitly specified", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

		toml1 := `
[foo]
  bar = 123
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		opts.DataValuesFlags = cmdtpl.DataValuesFlags{
			FromFiles: []string{"toml1.toml"},
			ReadFilesFunc: func(path string) ([]*files.File, error) {
				switch path {
				case "toml1.toml":
					return []*files.File{files.MustNewFileFromSource(files.NewBytesSource("toml1.toml", []byte(toml1)))}, nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
		require.EqualError(t, out.Err, "Extracting data value from file: Unmarshaling YAML data values file 'toml1.toml': yaml: line 2: did not find expected <document start>")
	})
	t.Run("skips when path given is a directory", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

		dvs1 := `---
int: 123
`
		toml1 := `
[foo]
  bar = 456
`

		expectedYAMLTplData := `values:
  int: 123
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		})

		ui := ui.NewTTY(false)
		opts := cmdtpl.NewOptions()

		opts.DataValuesFlags = cmdtpl.DataValuesFlags{
			FromFiles: []string{"values"},
			ReadFilesFunc: func(path string) ([]*files.File, error) {
				switch path {
				case "values":
					return []*files.File{
						files.MustNewFileFromSource(files.NewBytesSource("values/dvs1.yml", []byte(dvs1))),
						files.MustNewFileFromSource(files.NewBytesSource("values/toml1.toml", []byte(toml1))),
					}, nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
		require.NoError(t, out.Err)
		require.Len(t, out.Files, 1, "unexpected number of output files")

		file := out.Files[0]

		assert.Equal(t, "tpl.yml", file.RelativePath())
		assert.Equal(t, expectedYAMLTplData, string(file.Bytes()))
	})
}
