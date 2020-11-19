// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"strings"
	"testing"

	"github.com/k14s/difflib"
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
  tls_only: false
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
  tls_only: true
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

func TestMapOnlySchemaFillInDefaults(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
db_conn:
  hostname: default.com
  port: 0
  password: somepassword
  tls_only: false
  jobs:
    run: defaultJob
  metadata:
    missing_key: default value
    present_key:
      inner_key: other
`
	dataValuesYAML := `#@data/values
---
db_conn:
  password: mysecurepassword
  tls_only: true
  metadata:
    present_key:
      inner_key: value present
`
	templateYAML := `#@ load("@ytt:data", "data")
---
rendered: true
db: #@ data.values.db_conn
`

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

	expected := `rendered: true
db:
  password: mysecurepassword
  tls_only: true
  metadata:
    present_key:
      inner_key: value present
    missing_key: default value
  hostname: default.com
  port: 0
  jobs:
    run: defaultJob
`
	if string(out.Files[0].Bytes()) != expected {
		diff := difflib.PPDiff(strings.Split(string(out.Files[0].Bytes()), "\n"), strings.Split(expected, "\n"))
		t.Fatalf("Expected output to only include template YAML, differences:\n%s", diff)
	}
}

func TestArrayOnlySchemaChecksOk(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
- ""
`
	dataValuesYAML := `#@data/values
---
- first
- second
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

func TestMapAndArraySchemaChecksOk(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
db_conn:
- hostname: ""
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
- hostname: server.example.com
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

func TestArrayDataValuesNotConformingToSchemaFailsCheck(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
clients:
- id: 0
  name: ""
  flags:
  - name: ""
    set: false
`
	dataValuesYAML := `#@data/values
---
clients:
- id: 1
  name: Alice
  flags:
  - name: secure
    value: true   #! "value" is not in the schema
- id: 2
  name: Bob
  flags:
  - secure  #! expecting a map, got a string
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
	expectedErr := "Typechecking violations found: [Map item 'value' at dataValues.yml:8 is not defined in schema, Array item at dataValues.yml:12 was type string when *yamlmeta.MapType was expected]"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about a schema check failure, but got: %s", out.Err.Error())
	}
}

func TestEmptyArraySchemaErrors(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
vpc:
  name: ""
  subnet_ids: []
`
	dataValuesYAML := `#@data/values
---
vpc:
  name: "beax-a3543-5555"
  subnet_ids:
  - 0
  - 1
  - 10
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

	if out.Err == nil {
		t.Fatalf("Expected an error about the empty array in schema, but succeeded.")
	}

	expectedErr := "Expected one item in array (describing the type of its elements) at schema.yml:5"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about the empty array schema check failure, but got: %s", out.Err.Error())
	}
}

func TestArraySchemaWithMultipleValuesErrors(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
vpc:
  name: ""
  subnet_ids:
  - 0
  - 1
`
	dataValuesYAML := `#@data/values
---
vpc:
  name: "beax-a3543-5555"
  subnet_ids: []
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

	if out.Err == nil {
		t.Fatalf("Expected an error about the empty array in schema, but succeeded.")
	}

	expectedErr := "Expected one item (found 2) in array (describing the type of its elements) at schema.yml:5"
	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error about the empty array schema check failure, but got: %s", out.Err.Error())
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
