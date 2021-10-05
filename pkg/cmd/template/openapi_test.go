package template_test

import (
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
	"testing"
)

func TestOpenapi_map(t *testing.T) {
	opts := cmdtpl.NewOptions()
	// TODO pass in the --data-value-schema-inspect flag

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

	assertSucceeds(t, filesToProcess, expected, opts)
}
