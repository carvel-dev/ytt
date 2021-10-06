package template_test

import (
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOpenapi_map(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.DataValuesFlags.InspectSchema = true
	opts.RegularFilesSourceOpts.OutputType = "openapi-v3"
	t.Run("with a single map item", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
foo: some value
`
	expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Openapi schema generated from ytt Data Values Schema
paths: {}
components:
  schemas:
    type: object
    additionalProperties: false
    properties:
      foo:
        type: string
        default: some value
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

	assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("on nested maps", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
foo:
  int_key: 10
  bool_key: true
  false_key: false
  string_key: some text
  inner_map:
    float_key: 9.1`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Openapi schema generated from ytt Data Values Schema
paths: {}
components:
  schemas:
    type: object
    additionalProperties: false
    properties:
      foo:
        type: object
        additionalProperties: false
        properties:
          int_key:
            type: integer
            default: 10
          bool_key:
            type: boolean
            default: true
          false_key:
            type: boolean
            default: false
          string_key:
            type: string
            default: some text
          inner_map:
            type: object
            additionalProperties: false
            properties:
              float_key:
                type: number
                format: float
                default: 9.1`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
}

//Returning a DocSet
func assertSucceedsDocSet(t *testing.T, filesToProcess []*files.File, expectedOut string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	outBytes, err := out.DocSet.AsBytes()
	require.NoError(t, err)

	require.Equal(t, expectedOut, string(outBytes))
}
