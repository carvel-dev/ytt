// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestSchema_Unused_returns_error(t *testing.T) {
	opts := cmdtpl.NewOptions()

	t.Run("An unused schema ref'd to a library", func(t *testing.T) {
		schemaBytes := []byte(`
#@library/ref "@libby"
#@data/values-schema
---
fooX: not-used
`)

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaBytes)),
		})

		assertFails(t, filesToProcess, "Expected all provided library schema documents to be used "+
			"but found unused: Schema belonging to library '@libby' on line schema.yml:4", opts)
	})

}

func TestSchema_When_invalid_reports_error(t *testing.T) {
	opts := cmdtpl.NewOptions()

	t.Run("array with fewer than one element", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
vpc:
  subnet_ids: []
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		expectedErr := `
Invalid schema - wrong number of items in array definition
==========================================================

schema.yml:
    |
  4 |   subnet_ids: []
    |

    = found: 0 array items
    = expected: exactly 1 array item, of the desired type
    = hint: in schema, the one item of the array implies the type of its elements.
    = hint: in schema, the default value for an array is always an empty list.
    = hint: default values can be overridden via a data values overlay.
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("array with more than one element", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
vpc:
  subnet_ids:
  - 0
  - 1
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		expectedErr := `
Invalid schema - wrong number of items in array definition
==========================================================

schema.yml:
    |
  4 |   subnet_ids:
    |

    = found: 2 array items
    = expected: exactly 1 array item, of the desired type
    = hint: in schema, the one item of the array implies the type of its elements.
    = hint: in schema, the default value for an array is always an empty list.
    = hint: default values can be overridden via a data values overlay.
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("array with a null value", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
vpc:
  subnet_ids:
  - null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		expectedErr := `
Invalid schema - null value not allowed here
============================================

schema.yml:
    |
  5 |   - null
    |

    = found: null value
    = expected: non-null value
    = hint: in YAML, omitting a value implies null.
    = hint: to set the default value to null, annotate with @schema/nullable.
    = hint: to allow any value, annotate with @schema/type any=True.
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("item with null value", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
vpc:
  subnet_ids: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})
		expectedErr := `
Invalid schema - null value not allowed here
============================================

schema.yml:
    |
  4 |   subnet_ids: null
    |

    = found: null value
    = expected: non-null value
    = hint: in YAML, omitting a value implies null.
    = hint: to set the default value to null, annotate with @schema/nullable.
    = hint: to allow any value, annotate with @schema/type any=True.
`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when a nullable map item has null value", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/nullable  
  bar: null
  #@schema/type any=True
  baz: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})
		expectedErr := `
Invalid schema - null value not allowed here
============================================

schema.yml:
    |
  5 |   bar: null
    |

    = found: null value
    = expected: non-null value
    = hint: in YAML, omitting a value implies null.
    = hint: to set the default value to null, annotate with @schema/nullable.
    = hint: to allow any value, annotate with @schema/type any=True.
`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has keyword other than any", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
#@schema/type unknown_kwarg=False
foo: 0
`
		expectedErr := `
Invalid schema
==============

unknown @schema/type annotation keyword argument
schema.yml:
    |
  4 | foo: 0
    |

    = found: unknown_kwarg
    = expected: A valid kwarg
    = hint: Supported kwargs are 'any'
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has value for any other than a bool", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
#@schema/type any=1
foo: 0
`

		expectedErr := `
Invalid schema
==============

unknown @schema/type annotation keyword argument
schema.yml:
    |
  4 | foo: 0
    |

    = found: starlark.Int
    = expected: starlark.Bool
    = hint: Supported kwargs are 'any'
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type has incomplete key word args", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
#@schema/type
foo: 0
`

		expectedErr := `
Invalid schema
==============

expected @schema/type annotation to have keyword argument and value
schema.yml:
    |
  4 | foo: 0
    |

    = found: missing keyword argument and value
    = expected: valid keyword argument and value
    = hint: Supported key-value pairs are 'any=True', 'any=False'
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)

		schemaYAML2 := `#@data/values-schema
---
#@schema/type any
foo: 0
`
		filesToProcess = files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML2))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema/type and schema/nullable annotate a map", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
#@schema/type any=True
#@schema/nullable
foo: 0
`

		expectedErr := `
Invalid schema
==============

@schema/nullable, and @schema/type any=True are mutually exclusive
schema.yml:
    |
  5 | foo: 0
    |

    = found: both @schema/nullable, and @schema/type any=True annotations
    = expected: one of schema/nullable, or schema/type any=True
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)

	})

	t.Run("when schema/default annotation value", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/default
foo: 0
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/default annotation
schema.yml:
    |
  4 | foo: 0
    |

    = found: missing value (in @schema/default above this item)
    = expected: integer (by schema.yml:4)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has multiple values", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/default 1, 2
foo: 0
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/default annotation
schema.yml:
    |
  4 | foo: 0
    |

    = found: 2 values (in @schema/default above this item)
    = expected: integer (by schema.yml:4)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is an invalid starlark Tuple", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/default any=True
foo: 0
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/default annotation
schema.yml:
    |
  4 | foo: 0
    |

    = found: (keyword argument in @schema/default above this item)
    = expected: integer (by schema.yml:4)
    = hint: this annotation only accepts one argument: the default value.
    = hint: value must be in Starlark format, e.g.: {'key': 'value'}, True.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is on an array item", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
foo:
#@schema/default "baz"
- bar

`

			expectedErr := `
Invalid schema - @schema/default not supported on array item
============================================================

schema.yml:
    |
  5 | - bar
    |



    = hint: do you mean to set a default value for the array?
    = hint: set an array's default by annotating its parent.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is incorrect type", func(t *testing.T) {
			t.Run("as a scalar", func(t *testing.T) {
				schemaYAML := `#@data/values-schema
---
#@schema/default 1
foo: a string

`

				expectedErr := `
Invalid schema - @schema/default is wrong type
==============================================

schema.yml:
    |
  4 | foo: a string
    |

    = found: integer
    = expected: string (by schema.yml:4)
    = hint: is the default value set using @schema/default?
`

				filesToProcess := files.NewSortedFiles([]*files.File{
					files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				})

				assertFails(t, filesToProcess, expectedErr, opts)
			})
			t.Run("as an array (a node)", func(t *testing.T) {

				schemaYAML := `#@data/values-schema
---
#@schema/default [{"item": 1}]
map: thing
`

				expectedErr := `
Invalid schema - @schema/default is wrong type
==============================================

schema.yml:
    |
  4 | map: thing
    |

    = found: array
    = expected: string (by schema.yml:4)
    = hint: is the default value set using @schema/default?
`

				filesToProcess := files.NewSortedFiles([]*files.File{
					files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				})

				assertFails(t, filesToProcess, expectedErr, opts)
			})
		})
	})
	t.Run("when schema/desc annotation value", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/desc
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/desc annotation
schema.yml:
    |
  4 | key: val
    |

    = found: missing value (in @schema/desc above this item)
    = expected: string
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has more than one arg", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/desc "two", "strings"
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/desc annotation
schema.yml:
    |
  4 | key: val
    |

    = found: 2 values (in @schema/desc above this item)
    = expected: string
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is not a string", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/desc 1
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/desc annotation
schema.yml:
    |
  4 | key: val
    |

    = found: Non-string value (in @schema/desc above this item)
    = expected: string
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is key=value pair", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/desc key=True
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/desc annotation
schema.yml:
    |
  4 | key: val
    |

    = found: keyword argument (in @schema/desc above this item)
    = expected: string
    = hint: this annotation only accepts one argument: a string.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
	})
}

func TestSchema_Provides_default_values(t *testing.T) {
	opts := cmdtpl.NewOptions()
	t.Run("initializes data values to the values set in the schema", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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

	t.Run("when the the @schema/default annotation used", func(t *testing.T) {
		t.Run("on a map item with scalar values", func(t *testing.T) {

			schemaYAML := `#@data/values-schema
---
#@schema/default 1
int: 0
#@schema/default "a string"
str: ""
#@schema/default True
bool: false
`
			templateYAML := `#@ load("@ytt:data", "data")
---
int: #@ data.values.int
str: #@ data.values.str
bool: #@ data.values.bool
`
			expected := `int: 1
str: a string
bool: true
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("on an array", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/default ["new", "array", "strings"]
foo:
- the array holds strings
#@schema/default [1,2,3]
bar:
- 7
`
			templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
bar: #@ data.values.bar
`
			expected := `foo:
- new
- array
- strings
bar:
- 1
- 2
- 3
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("on nested map", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/default [{'name': 'null_db'}]
databases:
- name: ""
  host: ""
#@schema/default {'admin': 'admin'}
users:
  admin: ""
  user: 
  - ""
`

			templateYAML := `#@ load("@ytt:data", "data")
---
databases: #@ data.values.databases
users: #@ data.values.users
`
			expected := `databases:
- name: null_db
  host: ""
users:
  admin: admin
  user: []
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("on a document", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
#@schema/default {'databases': [{'name': 'default', 'host': 'localhost'}]}
---
databases:
- name: ""
  host: ""
`

			templateYAML := `#@ load("@ytt:data", "data")
---
databases: #@ data.values.databases
`
			expected := `databases:
- name: default
  host: localhost
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("in combination with @schema/nullable and @schema/type", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/type any=True 
#@schema/default None
databases:
- name: ""
  host: ""

#@schema/nullable
#@schema/default {'admin':'admin'}
users:
  admin: ""
  user: 
  - ""
`

			templateYAML := `#@ load("@ytt:data", "data")
---
databases: #@ data.values.databases
users: #@ data.values.users
`
			expected := `databases: null
users:
  admin: admin
  user: []
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
	})
}
