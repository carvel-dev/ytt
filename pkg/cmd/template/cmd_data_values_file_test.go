// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
)

func TestDataValuesWithDataValuesFileFlags(t *testing.T) {
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
		ReadFileFunc: func(path string) ([]byte, error) {
			switch path {
			case "dvs1.yml":
				return dvs1, nil
			case "dvs2.yml":
				return dvs2, nil
			case "dvs3.yml":
				return dvs3, nil
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

func TestDataValuesFileFlagsWithLibRefPath(t *testing.T) {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
---
fromRoot: #@ data.values

#@ lib = library.get("lib")
--- #@ template.replace(lib.eval())`)

	libYamlTplData := []byte(`#@ load("@ytt:data", "data")
---
fromLibrary: #@ data.values`)

	dvs1 := []byte(`val1: 1`)

	dvs2 := []byte(`val2: 2`)

	dvs3 := []byte(`3`)

	expectedYAMLTplData := `fromRoot:
  val1: 1
fromLibrary: 
  val2: 2
  val3: 3
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/tpl.yml", libYamlTplData)),
	})

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags = cmdtpl.DataValuesFlags{
		FromFiles: []string{"c:\\User\\user\\dvs1.yml", "@lib:dvs2.yml"},
		KVsFromFiles: []string{"val3=@lib:c:\\User\\user\\dvs3.yml"},
		ReadFileFunc: func(path string) ([]byte, error) {
			switch path {
			case "c:\\User\\user\\dvs1.yml":
				return dvs1, nil
			case "dvs2.yml":
				return dvs2, nil
			case "c:\\User\\user\\dvs3.yml":
				return dvs3, nil
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

func TestDataValuesWithDataValuesFileFlagsForbiddenComment(t *testing.T) {
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
		ReadFileFunc: func(path string) ([]byte, error) {
			switch path {
			case "dvs1.yml":
				return dvs1, nil
			default:
				return nil, fmt.Errorf("Unknown file '%s'", path)
			}
		},
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.EqualError(t, out.Err, "Extracting data value from file: Checking data values file 'dvs1.yml': Expected to be plain YAML, having no annotations (hint: remove comments starting with `#@`)")
}
