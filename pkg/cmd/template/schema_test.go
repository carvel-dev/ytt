// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"strings"
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestMapOnlySchemaChecksOk(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
db_conn:
  hostname: ""
  port: 0
  username: ""
  password: ""
  metadata:
    run: jobName
top_level: ""
`
	dataValuesYAML := `#@data/values
---
db_conn:
  hostname: server.example.com
  port: 5432
  username: sa
  password: changeme
  metadata:
    run: ./build.sh
top_level: key
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was: %d", len(out.Files))
	}

	if string(out.Files[0].Bytes()) != "rendered: true\n" {
		t.Fatalf("Expected output to only include template YAML, but got: %s", out.Files[0].Bytes())
	}
}

func TestDataValuesNotConformingToEmptySchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
`
	dataValuesYAML := `#@data/values
---
not_in_schema: "this should fail the type check!"
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("values.yml", []byte(dataValuesYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)

	if out.Err == nil {
		t.Fatalf("Expected an error about a schema check failure, but succeeded.")
	}
	expectedErr := "Typechecking violations found: [Expected node at values.yml:2 to be nil, but was a *yamlmeta.Map]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}

}

func TestMapOnlyDataValuesNotConformingToSchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
db_conn:
  port: 0
  username:
    main: "0"
`
	dataValuesYAML := `#@data/values
---
db_conn:
  port: localHost
  username:
    main: 123
  password: changeme
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)

	if out.Err == nil {
		t.Fatalf("Expected an error about the schema check failures, but succeeded.")
	}
	expectedErr := "Typechecking violations found: [Map item 'port' at dataValues.yml:4 was type string when int was expected, Map item 'main' at dataValues.yml:6 was type int when string was expected, Map item 'password' at dataValues.yml:7 is not defined in schema]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}

}

func TestDataValuesAndSchemaContainsArrayFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
- ""
`
	dataValuesYAML := `#@data/values
---
- test
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)

	if out.Err == nil {
		t.Fatalf("Expected an error about arrays not being supported in schemas, but succeeded.")
	}
	expectedErr := "Arrays are currently not supported in schema"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about about arrays not being supported, but got: %s", out.Err.Error())
	}

}

func TestMultiDataValuesOneDataValuesNotConformingToSchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
hostname: ""
port: 0
username: ""
password: ""
`
	dataValuesYAML1 := `#@data/values
---
hostname: server.example.com
port: 5432
secret: super
`

	dataValuesYAML2 := `#@ load("@ytt:overlay", "overlay")
#@data/values
---
#@overlay/remove
secret:
#@overlay/match missing_ok=True
username: sa
#@overlay/match missing_ok=True
password: changeme
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues1.yml", []byte(dataValuesYAML1))),
		files.MustNewFileFromSource(files.NewBytesSource("dataValues2.yml", []byte(dataValuesYAML2))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected an error about a schema check failure, but succeeded.")
	}

	expectedErr := "Typechecking violations found: [Map item 'secret' at dataValues1.yml:5 is not defined in schema]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}
}

func TestSchemaFileButNoSchemaFlagExpectsWarning(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but got failure: %v", out.Err.Error())
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was: %d", len(out.Files))
	}

	if string(out.Files[0].Bytes()) != "rendered: true\n" {
		t.Fatalf("Expected output to only include template YAML, but got: %s", out.Files[0].Bytes())
	}
}

func TestNoSchemaFileSchemaFlagSet(t *testing.T) {
	dataValuesYAML := `#@data/values
---
db_conn:
  hostname: server.example.com
  port: 5432
  username: sa
  password: changeme
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err == nil {
		t.Fatalf("Expected RunWithFiles to fail with message about schema enabled but no schema provided, but was a success")
	}

	if !strings.Contains(out.Err.Error(), "Schema experiment flag was enabled but no schema document was provided.") {
		t.Fatalf("Expected an error about schema enabled but no schema provided, but got: %v", out.Err.Error())
	}

}
