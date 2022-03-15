// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"

	"testing"
)

func TestAssertValidateOnDataValues(t *testing.T) {
	t.Run("succeeds with data values files using --data-values-inspect", func(t *testing.T) {
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

	t.Run("assert fails with only data values files", func(t *testing.T) {
		t.Run("on data values", func(t *testing.T) {
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
		t.Run("on a document with variables in validation", func(t *testing.T) {
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
	})

	t.Run("fails when @assert/validate is not well-formed", func(t *testing.T) {
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
}

func TestAssertValidateOnLibraryDataValues(t *testing.T) {
 	t.Run("succeeds when library is evaluated", func(t *testing.T) {
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

	t.Run("assert fails when library is evaluated", func(t *testing.T) {
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