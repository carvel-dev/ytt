// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/files"
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

func TestSchema_When_malformed_reports_error(t *testing.T) {
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
	t.Run("when schema/type annotation", func(t *testing.T) {
		t.Run("has keyword other than any", func(t *testing.T) {
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
  3 | #@schema/type unknown_kwarg=False
  4 | foo: 0
    |

    = found: unknown_kwarg (by schema.yml:3)
    = expected: A valid kwarg
    = hint: Supported kwargs are 'any'
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has value for any other than a bool", func(t *testing.T) {
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
  3 | #@schema/type any=1
  4 | foo: 0
    |

    = found: starlark.Int (by schema.yml:3)
    = expected: starlark.Bool
    = hint: Supported kwargs are 'any'
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has incomplete key word args", func(t *testing.T) {
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
  3 | #@schema/type
  4 | foo: 0
    |

    = found: missing keyword argument and value (by schema.yml:3)
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
			expectedErr2 := `
Invalid schema
==============

expected @schema/type annotation to have keyword argument and value
schema.yml:
    |
  3 | #@schema/type any
  4 | foo: 0
    |

    = found: missing keyword argument and value (by schema.yml:3)
    = expected: valid keyword argument and value
    = hint: Supported key-value pairs are 'any=True', 'any=False'
`

			filesToProcess = files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML2))),
			})

			assertFails(t, filesToProcess, expectedErr2, opts)
		})
		t.Run("has nested schema/... annotations", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/type any=True
foo:
  #@schema/nullable
  #@schema/type any=False
  #@schema/default 1
  bar: 7
`

			expectedErr := `
Invalid schema
==============

Schema was specified within an "any type" fragment
schema.yml:
    |
  5 |   #@schema/nullable
  6 |   #@schema/type any=False
  7 |   #@schema/default 1
  8 |   bar: 7
    |

    = found: @schema/nullable, @schema/type, @schema/default annotation(s)
    = expected: no '@schema/...' on nodes within a node annotated '@schema/type any=True'
    = hint: an "any type" fragment has no constraints; nested schema annotations conflict and are disallowed
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
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
  3 | #@schema/default
  4 | foo: 0
    |

    = found: missing value in @schema/default (by schema.yml:3)
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
  3 | #@schema/default 1, 2
  4 | foo: 0
    |

    = found: 2 values in @schema/default (by schema.yml:3)
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
  3 | #@schema/default any=True
  4 | foo: 0
    |

    = found: keyword argument in @schema/default (by schema.yml:3)
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
  4 | #@schema/default "baz"
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
  3 | #@schema/default 1
  4 | foo: a string
    |

    = found: integer (at schema.yml:3)
    = expected: string (by schema.yml:4)
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
  3 | #@schema/default [{"item": 1}]
  4 | map: thing
    |

    = found: array (at schema.yml:3)
    = expected: string (by schema.yml:4)
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
  3 | #@schema/desc
  4 | key: val
    |

    = found: missing value in @schema/desc (by schema.yml:3)
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
  3 | #@schema/desc "two", "strings"
  4 | key: val
    |

    = found: 2 values in @schema/desc (by schema.yml:3)
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
  3 | #@schema/desc 1
  4 | key: val
    |

    = found: Non-string value in @schema/desc (by schema.yml:3)
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
  3 | #@schema/desc key=True
  4 | key: val
    |

    = found: keyword argument in @schema/desc (by schema.yml:3)
    = expected: string
    = hint: this annotation only accepts one argument: a string.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
	})
	t.Run("when schema/title annotation value", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/title
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/title annotation
schema.yml:
    |
  3 | #@schema/title
  4 | key: val
    |

    = found: missing value in @schema/title (by schema.yml:3)
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
#@schema/title "two", "strings"
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/title annotation
schema.yml:
    |
  3 | #@schema/title "two", "strings"
  4 | key: val
    |

    = found: 2 values in @schema/title (by schema.yml:3)
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
#@schema/title 1
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/title annotation
schema.yml:
    |
  3 | #@schema/title 1
  4 | key: val
    |

    = found: Non-string value in @schema/title (by schema.yml:3)
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
#@schema/title key=True
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/title annotation
schema.yml:
    |
  3 | #@schema/title key=True
  4 | key: val
    |

    = found: keyword argument in @schema/title (by schema.yml:3)
    = expected: string
    = hint: this annotation only accepts one argument: a string.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
	})
	t.Run("when schema/deprecated annotation value", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/deprecated
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/deprecated annotation
schema.yml:
    |
  3 | #@schema/deprecated
  4 | key: val
    |

    = found: missing value in @schema/deprecated (by schema.yml:3)
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
#@schema/deprecated "two", "strings"
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/deprecated annotation
schema.yml:
    |
  3 | #@schema/deprecated "two", "strings"
  4 | key: val
    |

    = found: 2 values in @schema/deprecated (by schema.yml:3)
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
#@schema/deprecated 1
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/deprecated annotation
schema.yml:
    |
  3 | #@schema/deprecated 1
  4 | key: val
    |

    = found: Non-string value in @schema/deprecated (by schema.yml:3)
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
#@schema/deprecated key=True
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/deprecated annotation
schema.yml:
    |
  3 | #@schema/deprecated key=True
  4 | key: val
    |

    = found: keyword argument in @schema/deprecated (by schema.yml:3)
    = expected: string
    = hint: this annotation only accepts one argument: a string.
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is on an a document", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
#@schema/deprecated ""
---
foo:
- bar
`

			expectedErr := `
Invalid schema
==============

@schema/deprecated not supported on a document
schema.yml:
    |
  2 | #@schema/deprecated ""
  3 | ---
    |



    = hint: do you mean to deprecate the entire schema document?
    = hint: use schema/deprecated on individual keys.
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
	})
	t.Run("when schema/examples annotation value", func(t *testing.T) {
		t.Run("is empty", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples
  4 | key: val
    |

    = found: missing value in @schema/examples (by schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is not a tuple", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples "example value"
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples "example value"
  4 | key: val
    |

    = found: string for @schema/examples (at schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is an empty tuple", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples ()
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples ()
  4 | key: val
    |

    = found: empty tuple in @schema/examples (by schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("has more than two args in an example", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples ("key example", "val", "value")
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples ("key example", "val", "value")
  4 | key: val
    |

    = found: 3-tuple argument in @schema/examples (by schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("example description is not a string", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples (3.14, 5)
key: 7.3
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples (3.14, 5)
  4 | key: 7.3
    |

    = found: float value for @schema/examples (at schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("is key=value pair", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples key=True
key: val
`
			expectedErr := `
Invalid schema
==============

syntax error in @schema/examples annotation
schema.yml:
    |
  3 | #@schema/examples key=True
  4 | key: val
    |

    = found: keyword argument in @schema/examples (by schema.yml:3)
    = expected: 2-tuple containing description (string) and example value (of expected type)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertFails(t, filesToProcess, expectedErr, opts)
		})
		t.Run("does not match type of annotated node", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/examples ("Zero value", 0)
enabled: false
`
			expectedErr := `Invalid schema - @schema/examples has wrong type
================================================

schema.yml:
    |
  3 | #@schema/examples ("Zero value", 0)
  4 | enabled: false
    |

    = found: integer (by schema.yml:3)
    = expected: boolean (by schema.yml:4)
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
