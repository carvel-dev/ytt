// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"testing"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestSchema_Passes_when_DataValues_conform(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("when document's value is a map", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
db_conn:
- hostname: ""
  port: 0
  username: ""
  password: ""
  metadata:
    run: jobName
  timeout: 1.0
  ttl: 3.5
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
  timeout: 7.5
  ttl: 1
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
    timeout: 7.5
    ttl: 1
  top_level: key
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when document's value is an array", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
- ""
`
		dataValuesYAML := `#@data/values
---
- first
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

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when document's value is a scalar", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
42
`
		dataValuesYAML := `#@data/values
---
13
`
		templateYAML := `#@ load("@ytt:data", "data")
---
data_value: #@ data.values
`
		expected := "data_value: 13\n"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})

	t.Run("when a data value is passed using --data-value", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		schemaYAML := `#@schema/match data_values=True
---
foo: bar
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags.KVsFromStrings = []string{"foo=myVal"}
		expected := `rendered: myVal
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, cmdOpts)
	})
	t.Run("when a data value is passed using --data-value-yaml", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		schemaYAML := `#@schema/match data_values=True
---
foo: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags.KVsFromYAML = []string{"foo=42"}
		expected := `rendered: 42
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, cmdOpts)
	})

	t.Run("when neither schema nor data values are given", func(t *testing.T) {
		assertSucceeds(t,
			files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte("true"))),
			}),
			"true\n", opts)
	})
}

func TestSchema_Reports_violations_when_DataValues_do_NOT_conform(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("when map item's key is not among those declared in schema", func(t *testing.T) {
		t.Skip("This test case will be covered in https://github.com/vmware-tanzu/carvel-ytt/issues/344")
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
			files.MustNewFileFromSource(files.NewBytesSource("data_values.yml", []byte(dataValuesYAML))),
		})
		expectedErr := `data_values.yml:5 | password: i should not be here
                  |
                  | UNEXPECTED KEY - the key of this item was not found in the schema's corresponding map:
                  |      found: password
                  |   expected: (a key defined in map) (by schema.yml:3)
                  |   (hint: declare data values in schema and override them in a data values document)`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when map item's value is the wrong type", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
db_conn:
  port: 0
  username:
    main: "0"
  timeout: 1.0
`
		dataValuesYAML := `#@data/values
---
db_conn:
  port: localhost
  username:
    main: 123
  timeout: 5m
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("data_values.yml", []byte(dataValuesYAML))),
		})
		expectedErr := `
data_values.yml:4 |   port: localhost
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: integer (by schema.yml:4)


data_values.yml:6 |     main: 123
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: integer
                  |   expected: string (by schema.yml:6)


data_values.yml:7 |   timeout: 5m
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: float (by schema.yml:7)
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when map item's value is null but is not nullable", func(t *testing.T) {
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
		expectedErr := `
dataValues.yml:3 | app: null
                 |
                 | TYPE MISMATCH - the value of this item is not what schema expected:
                 |      found: null
                 |   expected: integer (by schema.yml:3)`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when map item's value is wrong type and schema/nullable is set", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/nullable
foo: 0
`
		dataValuesYAML := `#@data/values
---
foo: "bar"
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
dataValues.yml:3 | foo: "bar"
                 |
                 | TYPE MISMATCH - the value of this item is not what schema expected:
                 |      found: string
                 |   expected: integer (by schema.yml:4)

`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when array item's value is the wrong type", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
clients:
- flags:
  - floats:
    - 1.0
`
		dataValuesYAML := `#@data/values
---
clients:
- flags: secure  #! expecting an array, got a string
- flags:
  - secure  #! expecting a map, got a string
- flags:
  - floats:
    - one  #! expecting a float, got a string
    - true  #! expecting a float, got a bool
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("data_values.yml", []byte(dataValuesYAML))),
		})

		expectedErr := `
data_values.yml:4 | - flags: secure  #! expecting an array, got a string
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: array (by schema.yml:4)


data_values.yml:6 |   - secure  #! expecting a map, got a string
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: map (by schema.yml:5)


data_values.yml:9 |     - one  #! expecting a float, got a string
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: float (by schema.yml:6)


data_values.yml:10 |     - true  #! expecting a float, got a bool
                   |
                   | TYPE MISMATCH - the value of this item is not what schema expected:
                   |      found: boolean
                   |   expected: float (by schema.yml:6)
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when a data value is passed using --data-value, but schema expects a non string", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		schemaYAML := `#@schema/match data_values=True
---
foo: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags.KVsFromStrings = []string{"foo=42"}
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `key 'foo' (kv arg):1 |
                     |
                     | TYPE MISMATCH - the value of this item is not what schema expected:
                     |      found: string
                     |   expected: integer (by schema.yml:3)`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("when schema is null and non-empty data values is given", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
`
		dataValuesYAML := `#@data/values
---
foo: non-empty data value
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", []byte(dataValuesYAML))),
		})
		expectedErr := "data values were found in data values file(s), but schema (schema.yml:2) has no values defined\n"
		expectedErr += "(hint: define matching keys from data values files(s) in the schema, or do not enable the schema feature)"

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("checks after every data values document is processed (and stops if there was a violation)", func(t *testing.T) {
		t.Skip("This test case will be covered in https://github.com/vmware-tanzu/carvel-ytt/issues/344")
		schemaYAML := `#@schema/match data_values=True
---
hostname: ""
`
		nonConformingDataValueYAML := `#@data/values
---
not_in_schema: this should be the only violation reported
`

		wouldFixNonConformingDataValueYAML := `#@ load("@ytt:overlay", "overlay")
#@data/values
---
#@overlay/remove
not_in_schema:
`
		templateYAML := `---
rendered: true`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("nonConforming.yml", []byte(nonConformingDataValueYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("fixNonConforming.yml", []byte(wouldFixNonConformingDataValueYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		expectedErr := `
nonConforming.yml:3 | not_in_schema: this should be the only violation reported
                    |
                    | UNEXPECTED KEY - the key of this item was not found in the schema's corresponding map:
                    |      found: not_in_schema
                    |   expected: (a key defined in map) (by schema.yml:2)
                    |   (hint: declare data values in schema and override them in a data values document)`

		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when schema expects a scalar as an array item, but an array is provided", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
array: [true]
`
		dataValuesYAML := `
#@data/values
---
array: [ [1] ]
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", []byte(dataValuesYAML))),
		})
		expectedErr := `TYPE MISMATCH - the value of this item is not what schema expected:
             |      found: array
             |   expected: boolean (by schema.yml:3)`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema expects a scalar as an array item, but a map is provided", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
array: [true]
`
		dataValuesYAML := `
#@data/values
---
array: [ {a: 1} ]
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", []byte(dataValuesYAML))),
		})
		expectedErr := `TYPE MISMATCH - the value of this item is not what schema expected:
             |      found: map
             |   expected: boolean (by schema.yml:3)`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func TestSchema_Provides_default_values(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

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
		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("array default to an empty list", func(t *testing.T) {
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

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when a key in the data value is omitted yet present in the schema, it is filled in", func(t *testing.T) {
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
  - id: 2
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

		assertSucceeds(t, filesToProcess, expected, opts)
	})
}

func TestSchema_Allows_null_values_via_nullable_annotation(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true
	t.Run("when the value is a map", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
defaults:
  #@schema/nullable
  contains_map:
    a: 1
    b: 2
overriden:
  #@schema/nullable
  contains_map:
    a: 1
    b: 1

`
		dataValuesYAML := `#@data/values
---
overriden:
  contains_map:
    b: 2
`
		templateYAML := `#@ load("@ytt:data", "data")
---
defaults: #@ data.values.defaults
overriden: #@ data.values.overriden
`
		expected := `defaults:
  contains_map: null
overriden:
  contains_map:
    b: 2
    a: 1
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when the value is a array", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
defaults:
  #@schema/nullable
  contains_array:
  - ""
overriden:
  #@schema/nullable
  contains_array:
  - a: 1
    b: 0
`
		dataValuesYAML := `#@data/values
---
overriden:
  contains_array:
  - a: 20
`
		templateYAML := `#@ load("@ytt:data", "data")
---
defaults: #@ data.values.defaults
overriden: #@ data.values.overriden
`
		expected := `defaults:
  contains_array: null
overriden:
  contains_array:
  - a: 20
    b: 0
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when the value is a scalar", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
defaults:
  #@schema/nullable
  nullable_string: "empty"
  #@schema/nullable
  nullable_int: 10
  #@schema/nullable
  nullable_bool: true
overriden:
  #@schema/nullable
  nullable_string: "empty"
  #@schema/nullable
  nullable_int: 10
  #@schema/nullable
  nullable_bool: false
`
		dataValuesYAML := `#@data/values
---
overriden:
  nullable_string: set from data value
  nullable_int: 42
  nullable_bool: true
`
		templateYAML := `#@ load("@ytt:data", "data")
---
defaults: #@ data.values.defaults
overriden: #@ data.values.overriden
`

		expected := `defaults:
  nullable_string: null
  nullable_int: null
  nullable_bool: null
overriden:
  nullable_string: set from data value
  nullable_int: 42
  nullable_bool: true
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
}

func TestSchema_Allows_any_value_via_any_annotation(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("when any is true and set on a map", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type any=True
foo: ""
#@schema/type any=True
baz:
  a: 1
`
		dataValuesYAML := `#@data/values
---
foo: ~
baz:
  a: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
baz: #@ data.values.baz
`
		expected := `foo: null
baz:
  a: 7
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when any is true and set on an array", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
foo: 
#@schema/type any=True
- ""
  
`
		dataValuesYAML := `#@data/values
---
foo: ["bar", 7, ~]
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`
		expected := `foo:
- bar
- 7
- null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when any is false and set on a map", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type any=False
foo: 0
`
		dataValuesYAML := `#@data/values
---
foo: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`
		expected := `foo: 7
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when any is set on maps and arrays with nested dvs and overlay/replace", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type any=True
foo: ""
bar:
#@schema/type any=True
- 0
#@schema/type any=True
baz:
  a: 1
`
		dataValuesYAML := `#@data/values
---
#@overlay/replace
foo:
  ball: red
bar:
- newMap: 
  - ""
  - 8
#@overlay/replace
baz:
- newArray: foobar
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
bar: #@ data.values.bar
baz: #@ data.values.baz
`
		expected := `foo:
  ball: red
bar:
- newMap:
  - ""
  - 8
baz:
- newArray: foobar
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})

	t.Run("when any is set on nested maps", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
baz:
  #@schema/type any=True
  a: 1
`
		dataValuesYAML := `#@data/values
---
#@overlay/replace
baz:
  a: foobar
`
		templateYAML := `#@ load("@ytt:data", "data")
---
baz: #@ data.values.baz
`
		expected := `baz:
  a: foobar
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})

	t.Run("when schema/type and schema/nullable annotate a map", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type any=True
#@schema/nullable
foo: 0
`
		dataValuesYAML := `#@data/values
---
foo: "bar" 
`
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`
		expected := `foo: bar
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
}

func TestSchema_Is_scoped_to_a_library(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("when data values are ref'ed to a library, they are only checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		valuesYAML := []byte(`
#@library/ref "@lib"
#@data/values
---
in_library: 7
`)

		schemaYAML := []byte(`
#@schema/match data_values=True
---
in_root_library: ""
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
lib_data_values: #@ data.values.in_library`)

		libSchemaData := []byte(`
#@schema/match data_values=True
---
in_library: 0`)

		expectedYAMLTplData := `lib_data_values: 7
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaData)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when data values are programmatically set on a library, they are checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---
#@ def dvs_from_root():
foo: from "root" library
#@ end
--- #@ template.replace(library.get("lib").with_data_values(dvs_from_root()).eval())`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo: ""`)

		expectedYAMLTplData := `foo: from "root" library
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when data values are programmatically exported from a library, they are checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:library", "library")
--- #@ library.get("lib").data_values()`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo: "from library"`)

		expectedYAMLTplData := `foo: from library
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when symbols are programmatically exported from a library, the library's schema is checked", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
--- #@ library.get("lib").export("exported_func")()`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
#@ def exported_func(): return data.values`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo: "value exported from library"`)

		expectedYAMLTplData := `foo: value exported from library
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.lib.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when a library is evaluated in a text template, data values are checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("data.lib.txt", "dvs_from_text")
---
key: #@ dvs_from_text()`)

		textTemplateData := []byte(`
(@ load("@ytt:library", "library") @)
(@ libDataValues = library.get("lib").data_values() @)

(@ def dvs_from_text(): -@)
(@-= str([libDataValues.bar, libDataValues.foo])  @)
(@- end @)
`)
// set schema with "library/ref" to verify that the data is being included in the library
		schemaData := []byte(`
#@library/ref "@lib"
#@schema/match data_values=True
---
bar: from_root_schema
foo: ""
`)

		libDataValues := []byte(`
#@data/values
---
foo: from_library_dv
`)
		expectedYAMLTplData := `key: '["from_root_schema", "from_library_dv"]'
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("data.lib.txt", textTemplateData)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libDataValues)),
		})
		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when a library is evaluated in a Starlark file, data values are checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("data.lib.star", "dvs_from_starlark")
---
key: #@ dvs_from_starlark()`)

		starTemplateData := []byte(`
load("@ytt:library", "library")
libDataValues = library.get("lib").data_values()

def dvs_from_starlark():
return [libDataValues.bar, libDataValues.foo]
end
`)
// set schema with "library/ref" to verify that the data is being included in the library
		schemaData := []byte(`
#@library/ref "@lib"
#@schema/match data_values=True
---
foo: ""
bar: from_library_schema
`)

		libConfig := []byte(`
#@data/values
---
foo: from_library_dv
`)
		expectedYAMLTplData := `key:
- from_library_schema
- from_library_dv
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("data.lib.star", starTemplateData)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", libConfig)),
		})
		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when a library is evaluated, schema violations are reported", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		valuesYAML := []byte(`
#@library/ref "@lib"
#@data/values
---
foo: bar
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo: 42`)

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		expectedErr := `
- library.eval: Evaluating library 'lib': Overlaying data values (in following order: additional data values): 
    in <toplevel>
      config.yml:4 | --- #@ template.replace(library.get("lib").eval())

    reason:
     values.yml:5 | foo: bar
                  |
                  | TYPE MISMATCH - the value of this item is not what schema expected:
                  |      found: string
                  |   expected: integer (by _ytt_lib/lib/schema.yml:4)
     
     `
		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when data value ref'ed to a library is passed using --data-value, it is checked by that library's schema", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@schema/match data_values=True
---
foo: bar
`)

		libConfigYAML := []byte(`
#! lib/config.yml
#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`)

		cmdOpts.DataValuesFlags.KVsFromStrings = []string{"@lib:foo=myVal"}
		expectedYAMLTplData := `foo: myVal
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, cmdOpts)
	})
	t.Run("when data value ref'ed to a library is passed using --data-value-yaml, it is checked by that library's schema", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@schema/match data_values=True
---
cow: 7
`)

		libConfigYAML := []byte(`
#! lib/config.yml
#@ load("@ytt:data", "data")
---
cow: #@ data.values.cow
`)

		cmdOpts.DataValuesFlags.KVsFromYAML = []string{"@lib:cow=42"}
		expectedYAMLTplData := `cow: 42
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, cmdOpts)
	})
	t.Run("when data value ref'ed to a library is passed using --data-value, but schema expects an int, schema violation is reported", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		cmdOpts.SchemaEnabled = true
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@schema/match data_values=True
---
foo: 7
`)

		libConfigYAML := []byte(`
#! lib/config.yml
#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`)

		cmdOpts.DataValuesFlags.KVsFromStrings = []string{"@lib:foo=42"}
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		expectedErr := `
- library.eval: Evaluating library 'lib': Overlaying data values (in following order: additional data values): 
    in <toplevel>
      root.yml:5 | --- #@ template.replace(library.get("lib").eval())

    reason:
     key 'foo' (kv arg):1 |
                          |
                          | TYPE MISMATCH - the value of this item is not what schema expected:
                          |      found: string
                          |   expected: integer (by _ytt_lib/lib/schema.yml:5)
     
     `
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})
}

func TestSchema_Overlay_multiple_schema_files(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("when additional schema file is scoped to root library", func(t *testing.T) {
		schemaYAML1 := `#@schema/match data_values=True
---
db_conn:
- hostname: ""
`

		schemaYAML2 := `#@ load("@ytt:overlay", "overlay")
#@schema/match data_values=True
---
db_conn:
#@overlay/match by=overlay.all, expects="1+"
-  
  #@overlay/match missing_ok=True
  metadata:
    run: jobName
#@overlay/match missing_ok=True
top_level: ""
`
		dataValuesYAML := `#@data/values
---
db_conn:
- hostname: server.example.com
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
    metadata:
      run: ./build.sh
  top_level: key
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema1.yml", []byte(schemaYAML1))),
			files.MustNewFileFromSource(files.NewBytesSource("schema2.yml", []byte(schemaYAML2))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when additional schema file is scoped to private library", func(t *testing.T) {
		rootYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")
--- #@ template.replace(library.get("libby").with_data_values({"foo": {"ree": "set from root"}}).eval())
---
root_data_values: #@ data.values`)

		overlayLibSchemaYAML := []byte(`
#@library/ref "@libby"
#@schema/match data_values=True
---
foo:
  #@overlay/match missing_ok=True
  ree: ""
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
libby_data_values: #@ data.values`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo:
  bar: 3`)

		expectedYAMLTplData := `libby_data_values:
  foo:
    bar: 3
    ree: set from root
---
root_data_values: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("more-schema.yml", overlayLibSchemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when additional schema for private library is provided programmatically", func(t *testing.T) {
		rootYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")

#@ def more_schema():
foo:
  #@overlay/match missing_ok=True
  ree: ""
#@ end

#@ libby = library.get("libby")
#@ libby = libby.with_schema(more_schema())
#@ libby = libby.with_data_values({"foo": {"ree": "set from root"}})

--- #@ template.replace(libby.eval())
---
root_data_values: #@ data.values`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
libby_data_values: #@ data.values`)

		libSchemaYAML := []byte(`
#@schema/match data_values=True
---
foo:
  bar: 3`)

		expectedYAMLTplData := `libby_data_values:
  foo:
    bar: 3
    ree: set from root
---
root_data_values: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
}

func TestSchema_When_invalid_reports_error(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = true

	t.Run("array with fewer than one element", func(t *testing.T) {
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
		expectedErr := `
schema.yml:4 |   subnet_ids: []
             |
             | INVALID ARRAY DEFINITION IN SCHEMA - unable to determine the desired type
             |      found: 0 array items
             |   expected: exactly 1 array item, of the desired type
             |   (hint: in a schema, the item of an array defines the type of its elements; its default value is an empty list)`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("array with more than one element", func(t *testing.T) {
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
		expectedErr := `
schema.yml:4 |   subnet_ids:
             |
             | INVALID ARRAY DEFINITION IN SCHEMA - unable to determine the desired type
             |      found: 2 array items
             |   expected: exactly 1 array item, of the desired type
             |   (hint: to add elements to the default value of an array (i.e. an empty list), declare them in a @data/values document)`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("array with a nullable annotation", func(t *testing.T) {
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
		expectedErr := `
schema.yml:6 |   - 0
             |
             | INVALID SCHEMA - @schema/nullable is not supported on array items`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("item with null value", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
vpc:
  subnet_ids: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})
		expectedErr := `
schema.yml:4 |   subnet_ids: null
             |
             | INVALID SCHEMA - null value is not allowed in schema (no type can be inferred from it)
             |   (hint: to default to null, specify a value of the desired type and annotate with @schema/nullable)`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has keyword other than any", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type unknown_kwarg=False
foo: 0
`
		expectedErr := `schema.yml:4 | foo: 0
             |
             | INVALID SCHEMA - unknown @schema/type annotation keyword argument 'unknown_kwarg'. Supported kwargs are 'any'
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has value for any other than a bool", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type any=1
foo: 0
`
		expectedErr := `schema.yml:4 | foo: 0
             |
             | INVALID SCHEMA - processing @schema/type 'any' argument: expected starlark.Bool, but was starlark.Int
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has incomplete key word args", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
#@schema/type
foo: 0
`
		expectedErr := `schema.yml:4 | foo: 0
             |
             | INVALID SCHEMA - expected @schema/type annotation to have keyword argument and value. Supported key-value pairs are 'any=True', 'any=False'
`
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)

		schemaYAML2 := `#@schema/match data_values=True
---
#@schema/type any
foo: 0
`
		filesToProcess = files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML2))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func TestSchema_feature_disabled(t *testing.T) {
	opts := cmdtpl.NewOptions()
	opts.SchemaEnabled = false
	t.Run("warns when a schema is provided", func(t *testing.T) {
		stdout := bytes.NewBufferString("")
		stderr := bytes.NewBufferString("")
		ui := ui.NewCustomWriterTTY(false, stdout, stderr)

		schemaYAML := `#@schema/match data_values=True
---
`
		templateYAML := `---
rendered: true`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedStdErr := "Warning: schema document was detected (schema.yml), but schema experiment flag is not enabled. Did you mean to include --enable-experiment-schema?\n"
		expectedOut := "rendered: true\n"

		out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
		require.NoError(t, out.Err)

		assertStdoutAndStderr(t, bytes.NewBuffer(out.Files[0].Bytes()), stderr, expectedOut, expectedStdErr)
	})
	t.Run("errors when a schema used as a base data values", func(t *testing.T) {
		schemaYAML := `#@schema/match data_values=True
---
system_domain: "foo.domain"
`
		templateYAML := `#@ load("@ytt:data", "data")
---
system_domain: #@ data.values.system_domain
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := "NoneType has no .system_domain field or method"
		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func assertSucceeds(t *testing.T, filesToProcess []*files.File, expectedOut string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	require.Len(t, out.Files, 1, "unexpected number of output files")

	require.Equal(t, expectedOut, string(out.Files[0].Bytes()))
}

func assertFails(t *testing.T, filesToProcess []*files.File, expectedErr string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.Error(t, out.Err)

	require.Contains(t, out.Err.Error(), expectedErr)
}
