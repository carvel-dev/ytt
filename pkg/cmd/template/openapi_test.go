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

	schemaYAML := `#@data/values-schema
---
foo: some value
`
	expected := `
openapi: 3.0.0
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
        default: some value`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

	assertSucceedsDocSet(t, filesToProcess, expected, opts)
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
