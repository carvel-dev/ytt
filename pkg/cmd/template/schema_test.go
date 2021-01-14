// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"strings"
	"testing"

	"github.com/k14s/difflib"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	cmdui "github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
)

var opts *cmdtpl.TemplateOptions

func TestMain(m *testing.M) {
	opts = cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	m.Run()
}

func TestDataValueConformingToSchemaSucceeds(t *testing.T) {
	t.Run("Map Only", func(t *testing.T) {
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

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, "rendered: true\n")
	})
	t.Run("Map And Array", func(t *testing.T) {
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
#@overlay/append
- hostname: server.example.com
  port: 5432
  username: sa
  password: changeme
  metadata:
    run: ./build.sh
top_level: key
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`

		expected := `rendered:
  db_conn:
  - hostname: server.example.com
    port: 5432
    username: sa
    password: changeme
    metadata:
      run: ./build.sh
  top_level: key
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Array Only", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
- ""
`
		dataValuesYAML := `#@data/values
---
#@overlay/append
- first
#@overlay/append
- second
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`
		expected := `rendered:
- first
- second
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
}

func TestNullableAnnotation(t *testing.T) {
	t.Run("Allows null on scalars", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  name: ""
  #@schema/nullable
  nullable_string: "empty"
  #@schema/nullable
  nullable_int: 10
  foo: ""
`
		dataValuesYAML := `#@data/values
---
vpc:
  name: vpc-203d912a
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: true
vpc: #@ data.values.vpc
`

		expected := `rendered: true
vpc:
  name: vpc-203d912a
  nullable_string: null
  nullable_int: null
  foo: ""
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Allows null on top level map item", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/nullable
vpc:
  foo: "bar"
`
		dataValuesYAML := `---
#@data/values
---
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`

		expected := `vpc: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Allows null on map values", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  #@schema/nullable
  subnet_config:
  - id: 0
`
		dataValuesYAML := `#@data/values
---
vpc:
  subnet_config: ~
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`
		expected := `vpc:
  subnet_config: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Allows null on array values", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  #@schema/nullable
  subnet_config:
  - id: 0
    mask: "255.255.0.0"
    private: true
`
		dataValuesYAML := `#@data/values
---
vpc: {}
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`
		expected := `vpc:
  subnet_config: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Data values can override nullables", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  #@schema/nullable
  name: ""
`
		dataValuesYAML := `#@data/values
---
vpc:
  name: vpc-203d912a
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`
		expected := `vpc:
  name: vpc-203d912a
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
}

func TestDataValueNotConformingToSchemaFails(t *testing.T) {
	t.Run("map value type mismatched", func(t *testing.T) {
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
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Typechecking violations found: [Map item 'port' at dataValues.yml:4 was type string when int was expected, Map item 'main' at dataValues.yml:6 was type int when string was expected]"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("array value type mismatched", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
clients:
- flags:
  - name: ""
`
		dataValuesYAML := `#@data/values
---
clients:
- flags: secure  #! expecting a array, got a string
- flags:
  - secure  #! expecting a map, got a string
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Typechecking violations found: [Map item 'flags' at dataValues.yml:4 was type string when *yamlmeta.Array was expected, Array item at dataValues.yml:6 was type string when *yamlmeta.Map was expected]"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("map key is not present in schema", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
db_conn:
  port: 0
`
		dataValuesYAML := `#@data/values
---
db_conn:
  port: not an int  #! wrong type, but check values only when all keys in the map are valid
  password: i should not be here
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Typechecking violations found: [Map item 'password' at dataValues.yml:5 is not defined in schema]"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("null is given to a map item that is not nullable", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
app: 123
`
		dataValuesYAML := `#@data/values
---
app: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Typechecking violations found: [Map item 'app' at dataValues.yml:3 was type <nil> when *schema.ScalarType was expected"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("data values is given but schema is empty", func(t *testing.T) {
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
		expectedErr := "Typechecking violations found: [Expected node at values.yml:2 to be nil, but was a *yamlmeta.Map]"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("second data value conforms but the first data value does not conform", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
hostname: ""
`
		dataValuesYAML1 := `#@data/values
---
secret: super
`

		dataValuesYAML2 := `#@ load("@ytt:overlay", "overlay")
#@data/values
---
#@overlay/remove
secret:
`
		templateYAML := `---
rendered: true`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues1.yml", []byte(dataValuesYAML1))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues2.yml", []byte(dataValuesYAML2))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		expectedErr := "Typechecking violations found: [Map item 'secret' at dataValues1.yml:3 is not defined in schema]"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
}

func TestDefaultValuesAreFilledIn(t *testing.T) {
	t.Run("values specified in the schema are the default data values", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
system_domain: "foo.domain"
`
		templateYAML := `#@ load("@ytt:data", "data")
---
system_domain: #@ data.values.system_domain
`
		expected := `system_domain: foo.domain
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("Array defaults to an empty list", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  subnet_config:
  - id: 0
    mask: "255.255.0.0"
    private: true
`
		dataValuesYAML := `#@data/values
---
vpc: {}
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`
		expected := `vpc:
  subnet_config: []
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
	t.Run("When a key in the data value is omitted yet present in the schema it is filled in", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  name: "name value"
  subnet_config:
  - id: 0
    mask: "255.255.0.0"
    private: true
`
		dataValuesYAML := `#@data/values
---
vpc:
  subnet_config:
  #@overlay/append
  - id: 2
  #@overlay/append
  - id: 3
    mask: 255.255.255.0
`
		templateYAML := `#@ load("@ytt:data", "data")
---
vpc: #@ data.values.vpc
`
		expected := `vpc:
  name: name value
  subnet_config:
  - id: 2
    mask: 255.255.0.0
    private: true
  - id: 3
    mask: 255.255.255.0
    private: true
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
}

func TestNoSchemaProvided(t *testing.T) {
	t.Run("data value is given, provides an error and fails", func(t *testing.T) {
		dataValuesYAML := `#@data/values
---
db_conn:
`
		templateYAML := `---`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		expectedErr := "Schema feature is enabled but no schema document was provided"
		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("data value is not given, should succeed", func(t *testing.T) {
		templateYAML := `---
rendered: true`
		expected := `rendered: true
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
	})
}

func TestSchemaIsInvalidItFails(t *testing.T) {
	t.Run("Array value with less than one elements", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  subnet_ids: []
`
		dataValuesYAML := `#@data/values
---
vpc:
  subnet_ids:
  - 0
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Expected one item in array (describing the type of its elements) at schema.yml:4"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("Array value with more than one elements", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  subnet_ids:
  - 0
  - 1
`
		dataValuesYAML := `#@data/values
---
vpc:
  subnet_ids: []
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "Expected one item (found 2) in array (describing the type of its elements) at schema.yml:4"

		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
	t.Run("Array value with a nullable annotation", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  subnet_ids:
  #@schema/nullable
  - 0
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})
		expectedErr := "Array items cannot be annotated with #@schema/nullable (schema.yml:6). If this behaviour would be valuable, please submit an issue on https://github.com/vmware-tanzu/carvel-ytt"
		assertYTTWorkflowFailsWithErrorMessage(t, filesToProcess, expectedErr)
	})
}

func TestSchemaFeatureIsNotEnabledButSchemaIsPresentReportsAWarning(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = false

	schemaYAML := `#@schema/match data_values=True
---
`
	templateYAML := `---
rendered: true`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
	})
	expected := "rendered: true\n"
	assertYTTWorkflowSucceedsWithOutput(t, filesToProcess, expected)
}

func assertYTTWorkflowSucceedsWithOutput(t *testing.T, filesToProcess []*files.File, expectedOut string) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, cmdui.NewTTY(false))
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Errorf("Expected number of output files to be 1, but was: %d", len(out.Files))
	}

	if string(out.Files[0].Bytes()) != expectedOut {
		diff := difflib.PPDiff(strings.Split(string(out.Files[0].Bytes()), "\n"), strings.Split(expectedOut, "\n"))
		t.Errorf("Expected output to only include template YAML, differences:\n%s", diff)
	}
}

func assertYTTWorkflowFailsWithErrorMessage(t *testing.T, filesToProcess []*files.File, expectedErr string) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, cmdui.NewTTY(false))
	if out.Err == nil {
		t.Fatalf("Expected an error, but succeeded.")
	}

	if !strings.Contains(out.Err.Error(), expectedErr) {
		t.Fatalf("Expected an error %s, but got: %s", expectedErr, out.Err.Error())
	}
}
