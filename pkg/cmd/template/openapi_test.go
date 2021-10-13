package template_test

import (
	"testing"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestOpenapi_is_successful(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.DataValuesFlags.InspectSchema = true
	opts.RegularFilesSourceOpts.OutputType.Types = &[]string{"openapi-v3"}

	t.Run("on maps", func(t *testing.T) {
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
                default: 9.1
                format: float
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
}

func TestOpenapi_fails(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.DataValuesFlags.InspectSchema = true

	t.Run("when `--output` is anything other than 'openapi-v3'", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
foo: doesn't matter
`
		expectedErr := "Data Values Schema export only supported in OpenAPI v3 format; specify format with --output=openapi-v3 flag"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
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
