// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
)

func TestAssertValidateOnDataValuesSucceeds(t *testing.T) {
	t.Run("when validations pass using --data-values-inspect", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.Inspect = true
		dataValuesYAML := `#@ load("@ytt:assert", "assert")

#@data/values
#@assert/validate ("a non empty data values", lambda v: True if v else assert.fail("data values was empty"))
---
#@assert/validate ("a map with more than 3 elements", lambda v: True if len(v) > 3 else assert.fail("length of map was less than 3"))
my_map:
  #@assert/validate ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  string: server.example.com
  #@assert/validate ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  int: 54321
  #@assert/validate ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  float: 2.3
  #@assert/validate ("bool evaluating to true", lambda v:  v)
  bool: true
  #@assert/validate ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  nil: null
  #@assert/validate ("an array with less than or exactly 10 items", lambda v: True if len(v) <= 10 else assert.fail("array was more than 10 items"))
  my_array:
  #@assert/validate ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  - abc
  #@assert/validate ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  - 12345
  #@assert/validate ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  - 3.14
  #@assert/validate ("bool evaluating to true", lambda v:  v)
  - true
  #@assert/validate ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  - null
`

		expected := `my_map:
  string: server.example.com
  int: 54321
  float: 2.3
  bool: true
  nil: null
  my_array:
  - abc
  - 12345
  - 3.14
  - true
  - null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when validations on library data values pass", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---


#@ def additional_vals():
int: 10
#@overlay/match missing_ok=True
#@assert/validate ("a non empty string", lambda v: v )
str: "asdf"
#@ end

#@ lib = library.get("lib")
#@ lib2 = lib.with_data_values(additional_vals())
--- #@ template.replace(lib.eval())
--- #@ template.replace(lib2.eval())
`

		libValuesYAML := `#@ load("@ytt:assert", "assert")

#@data/values
---
#@assert/validate ("an integer over 1", lambda v: True if v > 1 else assert.fail("value was less than 1"))
int: 2
`

		libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

		expected := `values:
  int: 2
---
values:
  int: 10
  str: asdf
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", []byte(libValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
}

func TestAssertValidateOnDataValuesFails(t *testing.T) {
	t.Run("when the annotation is mal-formed because", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate 
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected @assert/validate to have 2-tuple as argument(s), but found no arguments (by schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is not a tuple", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate "string"
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected @assert/validate to have 2-tuple as argument(s), but found: "string" (by schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is an empty tuple", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate ()
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected @assert/validate 2-tuple, but found tuple with length 0 (by schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has incorrect number of args in a validation", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate (lambda v: v,)
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected @assert/validate 2-tuple, but found tuple with length 1 (by schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)

			dataValuesYAML2 := `#@data/values
---
#@assert/validate ("some validation", lambda v: v, "a message if valid", "a message if fail")
foo: bar
`
			expectedErr = `Invalid @assert/validate annotation - expected @assert/validate 2-tuple, but found tuple with length 4 (by schema.yml:3)`

			filesToProcess = files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML2))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)

		})
		t.Run("is a key=value pair", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate foo=True
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected @assert/validate to have 2-tuple as argument(s), but found keyword argument (by schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has incorrect string args in a validation", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate ({"foo":"bar"}, lambda v: v)
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected first item in the 2-tuple to be a string describing a valid value, but was dict (at schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has incorrect function arg in a validation", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			dataValuesYAML := `#@data/values
---
#@assert/validate ("some validation", True)
foo: bar
`

			expectedErr := `Invalid @assert/validate annotation - expected second item in the 2-tuple to be an assertion function, but was bool (at schema.yml:3)`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
	})
	t.Run("when validations fail", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		dataValuesYAML := `#@ load("@ytt:assert", "assert")

#@data/values
---
#@assert/validate ("a map with less than 3 elements", lambda v: True if len(v) < 3 else assert.fail("length of map was more than or equal to 3"))
my_map:
  #@assert/validate ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  string: ""
  #@assert/validate ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  int: 2
  #@assert/validate ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  float: 20.3
  #@assert/validate ("bool evaluating to true", lambda v:  v)
  bool: false
  #@assert/validate ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  nil: anything else
  #@assert/validate ("an array with more than or exactly 10 items", lambda v: True if len(v) >= 10 else assert.fail("array was less than 10 items"))
  my_array:
  #@assert/validate ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  - ""
  #@assert/validate ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  - 123
  #@assert/validate ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  - 3.14159
  #@assert/validate ("bool evaluating to true", lambda v:  v)
  - false
  #@assert/validate ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  - not null
`

		expectedErr := `One or more data values were invalid:
- "my_map" (schema.yml:6) requires "a map with less than 3 elements"; assert.fail: fail: length of map was more than or equal to 3 (by schema.yml:5)
- "string" (schema.yml:8) requires "a non-empty string"; assert.fail: fail: length of string was 0 (by schema.yml:7)
- "int" (schema.yml:10) requires "an int over 9000"; assert.fail: fail: int was less than 9000 (by schema.yml:9)
- "float" (schema.yml:12) requires "a float less than pi"; assert.fail: fail: float was more than 3.1415 (by schema.yml:11)
- "bool" (schema.yml:14) requires "bool evaluating to true" (by schema.yml:13)
- "nil" (schema.yml:16) requires "a null value"; assert.fail: fail: value was not null (by schema.yml:15)
- "my_array" (schema.yml:18) requires "an array with more than or exactly 10 items"; assert.fail: fail: array was less than 10 items (by schema.yml:17)
- array item (schema.yml:20) requires "a non-empty string"; assert.fail: fail: length of string was 0 (by schema.yml:19)
- array item (schema.yml:22) requires "an int over 9000"; assert.fail: fail: int was less than 9000 (by schema.yml:21)
- array item (schema.yml:24) requires "a float less than pi"; assert.fail: fail: float was more than 3.1415 (by schema.yml:23)
- array item (schema.yml:26) requires "bool evaluating to true" (by schema.yml:25)
- array item (schema.yml:28) requires "a null value"; assert.fail: fail: value was not null (by schema.yml:27)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations on a document fail with variables in validation", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		dataValuesYAML := `#@ load("@ytt:assert", "assert")
#@ validString = "a non empty data values"
#@ failureString = "data values was empty"

#@ def simpleExists(v):
#@   if v:
#@     return True
#@   else:
#@     assert.fail(failureString)
#@  end
#@ end

#@data/values
#@assert/validate (validString, simpleExists)
---
`

		expectedErr := `One or more data values were invalid:
- document (schema.yml:15) requires "a non empty data values"; assert.fail: fail: data values was empty (by schema.yml:14)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(dataValuesYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations on library data values fail", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---

#@ lib = library.get("lib")
--- #@ template.replace(lib.eval())
`

		libValuesYAML := `#@ load("@ytt:assert", "assert")

#@data/values
---
#@assert/validate ("an integer over 1", lambda v: True if v > 1 else assert.fail("value was less than 1"))
int: 0
`

		libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

		expectedErr := `- library.eval: Evaluating library 'lib': One or more data values were invalid:
    in <toplevel>
      config.yml:7 | --- #@ template.replace(lib.eval())

    reason:
     - "int" (_ytt_lib/lib/values.yml:6) requires "an integer over 1"; assert.fail: fail: value was less than 1 (by _ytt_lib/lib/values.yml:5)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/values.yml", []byte(libValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func TestSchemaValidationSucceeds(t *testing.T) {
	t.Run("when validations pass using --data-values-inspect", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.Inspect = true

		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
#@schema/validation ("a non empty data values", lambda v: True if v else assert.fail("data values was empty"))
---
#@schema/validation ("a map with more than 3 elements", lambda v: True if len(v) > 3 else assert.fail("length of map was less than 3"))
my_map:
  #@schema/validation ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  string: server.example.com
  #@schema/validation ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  int: 54321
  #@schema/validation ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  float: 2.3
  #@schema/validation ("bool evaluating to true", lambda v:  v)
  bool: true
  #@schema/validation ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  #@schema/nullable 
  nil: ""
  #@schema/validation ("an array with 1 or more items", lambda v: True if len(v) >= 1 else assert.fail("array was empty"))
  #@schema/default ['abc']
  my_array:
  #@schema/validation ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  - abc
`

		expected := `my_map:
  string: server.example.com
  int: 54321
  float: 2.3
  bool: true
  nil: null
  my_array:
  - abc
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when validations pass on an array item using --data-values-inspect", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.Inspect = true
		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
---
my_array:
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
- 5
`
		dvYAML := `#@data/values
---
my_array:
- 5
- 6
- 7
`

		expected := `my_array:
- 5
- 6
- 7
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dv.yml", []byte(dvYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when validations pass with other optional schema annotations", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.Inspect = true
		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
---
#@schema/default [5]
my_array:
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
- 6

#@schema/type any= True
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
my_map: 5

#@schema/nullable
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
my_other_map: 5
`
		dvYAML := `#@data/values
---
my_other_map: 6
`

		expected := `my_array:
- 5
my_map: 5
my_other_map: 6
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dv.yml", []byte(dvYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("when validations on library values pass", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---

#@ def additional_vals():
int: 10
#@ end

#@ lib = library.get("lib")
#@ lib2 = lib.with_data_values_schema(additional_vals())
--- #@ template.replace(lib.eval())
--- #@ template.replace(lib2.eval())
`

		libSchemaYAML := `#@ load("@ytt:assert", "assert")

#@data/values-schema
---
#@schema/validation ("an integer over 1", lambda v: True if v > 1 else assert.fail("value was less than 1"))
int: 2
`

		libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

		expected := `values:
  int: 2
---
values:
  int: 10
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", []byte(libSchemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
}

func TestSchemaValidationFails(t *testing.T) {
	t.Run("when validations fail", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		schemaYAML := `#@ load("@ytt:assert", "assert")

#@data/values-schema
---
  #@schema/validation ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  string: ""
  #@schema/validation ("an int over 9000", lambda v: True if v > 9000 else assert.fail("int was less than 9000"))
  int: 5432
  #@schema/validation ("a float less than pi", lambda v: True if v < 3.1415 else assert.fail("float was more than 3.1415"))
  float: 21.3
  #@schema/validation ("bool evaluating to true", lambda v:  v)
  bool: false
  #@schema/validation ("a null value", lambda v: True if v == None else assert.fail("value was not null"))
  nil: ""
  #@schema/validation ("an array with 1 or more items", lambda v: True if len(v) >= 1 else assert.fail("array was empty"))
  my_array:
  #@schema/validation ("a non-empty string", lambda v: True if len(v) > 0 else assert.fail("length of string was 0"))
  - abc
`

		expectedErr := `One or more data values were invalid:
- "string" (schema.yml:6) requires "a non-empty string"; assert.fail: fail: length of string was 0 (by schema.yml:5)
- "int" (schema.yml:8) requires "an int over 9000"; assert.fail: fail: int was less than 9000 (by schema.yml:7)
- "float" (schema.yml:10) requires "a float less than pi"; assert.fail: fail: float was more than 3.1415 (by schema.yml:9)
- "bool" (schema.yml:12) requires "bool evaluating to true" (by schema.yml:11)
- "nil" (schema.yml:14) requires "a null value"; assert.fail: fail: value was not null (by schema.yml:13)
- "my_array" (schema.yml:16) requires "an array with 1 or more items"; assert.fail: fail: array was empty (by schema.yml:15)`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations fail on an array item with data values overlays", func(t *testing.T) {
		opts := cmdtpl.NewOptions()

		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
---
my_array:
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
- 5
`
		dvYAML := `#@data/values
---
my_array:
- 5
- 6
- 7
- 1
- 2
- 3
`

		expectedErr := `One or more data values were invalid:
- array item (dv.yml:7) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:5)
- array item (dv.yml:8) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:5)
- array item (dv.yml:9) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:5)`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dv.yml", []byte(dvYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations on a document fail", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
#@schema/validation ("need more than 1 data value", lambda v: True if len(v) > 1 else assert.fail("less than 1 data values present"))
---
my_map: "abc"
`

		expectedErr := `One or more data values were invalid:
- document (schema.yml:4) requires "need more than 1 data value"; assert.fail: fail: less than 1 data values present (by schema.yml:3)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations fail with other optional schema annotations", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.Inspect = true
		schemaYAML := `#@ load("@ytt:assert", "assert")
#@data/values-schema
---
#@schema/default [3]
my_array:
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
- 6

#@schema/type any= True
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
my_map: 3

#@schema/nullable
#@schema/validation ("an integer larger than 4", lambda v: True if v > 4 else assert.fail("values less than 5"))
my_other_map: 5
`
		dvYAML := `#@data/values
---
my_other_map: 3
`

		expectedErr := `One or more data values were invalid:
- array item (schema.yml:5) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:6)
- "my_map" (schema.yml:11) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:10)
- "my_other_map" (dv.yml:3) requires "an integer larger than 4"; assert.fail: fail: values less than 5 (by schema.yml:14)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dv.yml", []byte(dvYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when validations on library schema fail", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		configYAML := `
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---

#@ lib = library.get("lib")
--- #@ template.replace(lib.eval())
`

		libSchemaYAML := `#@ load("@ytt:assert", "assert")

#@data/values-schema
---
#@schema/validation ("an integer over 1", lambda v: True if v > 1 else assert.fail("value was less than 1"))
int: 1
`

		libConfigYAML := `
#@ load("@ytt:data", "data")
---
values: #@ data.values
`

		expectedErr := `- library.eval: Evaluating library 'lib': One or more data values were invalid:
    in <toplevel>
      config.yml:7 | --- #@ template.replace(lib.eval())

    reason:
     - "int" (_ytt_lib/lib/schema.yml:6) requires "an integer over 1"; assert.fail: fail: value was less than 1 (by _ytt_lib/lib/schema.yml:5)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", []byte(libSchemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", []byte(libConfigYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})

}
