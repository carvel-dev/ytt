// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestSchema_Passes_when_DataValues_conform(t *testing.T) {
	opts := cmdtpl.NewOptions()

	t.Run("when document's value is a map", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
		schemaYAML := `#@data/values-schema
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
		schemaYAML := `#@data/values-schema
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
		schemaYAML := `#@data/values-schema
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
		schemaYAML := `#@data/values-schema
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

	t.Run("when additional schema file is overlay'd", func(t *testing.T) {
		schemaYAML1 := `#@data/values-schema
---
db_conn:
- hostname: ""
`

		schemaYAML2 := `#@ load("@ytt:overlay", "overlay")
#@data/values-schema
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
}

func TestSchema_Reports_violations_when_DataValues_do_NOT_conform(t *testing.T) {
	opts := cmdtpl.NewOptions()

	t.Run("when map item's key is not among those declared in schema", func(t *testing.T) {
		yamlTplData := []byte(`
#@ load("@ytt:data", "data")
values: #@ data.values`)

		schemaData := []byte(`#@data/values-schema
---
foo:
  bar: 42
`)
		dvs1 := `
foo:
  wrong_key: not right key`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaData)),
		})

		opts := cmdtpl.NewOptions()

		opts.DataValuesFlags = cmdtpl.DataValuesFlags{
			FromFiles: []string{"dvs1.yml"},
			ReadFileFunc: func(path string) ([]byte, error) {
				switch path {
				case "dvs1.yml":
					return []byte(dvs1), nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		expectedErrMsg := `Overlaying data values (in following order: additional data values): 
One or more data values were invalid
====================================

dvs1.yml:
    |
  3 |   wrong_key: not right key
    |

    = found: wrong_key
    = expected: (a key defined in map) (by schema.yml:3)
    = hint: declare data values in schema and override them in a data values document
`
		assertFails(t, filesToProcess, expectedErrMsg, opts)
	})
	t.Run("when map item's value is the wrong type", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
One or more data values were invalid
====================================

data_values.yml:
    |
  4 |   port: localhost
    |

    = found: string
    = expected: integer (by schema.yml:4)
    = hint: is the default value set using @schema/default?

data_values.yml:
    |
  6 |     main: 123
    |

    = found: integer
    = expected: string (by schema.yml:6)
    = hint: is the default value set using @schema/default?

data_values.yml:
    |
  7 |   timeout: 5m
    |

    = found: string
    = expected: float (by schema.yml:7)
    = hint: is the default value set using @schema/default?
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when map item's value is null but is not nullable", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
One or more data values were invalid
====================================

dataValues.yml:
    |
  3 | app: null
    |

    = found: null
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when map item's value is wrong type and schema/nullable is set", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
One or more data values were invalid
====================================

dataValues.yml:
    |
  3 | foo: "bar"
    |

    = found: string
    = expected: integer (by schema.yml:4)
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when array item's value is the wrong type", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
One or more data values were invalid
====================================

data_values.yml:
     |
   4 | - flags: secure  #! expecting an array, got a string
     |

     = found: string
     = expected: array (by schema.yml:4)
     = hint: is the default value set using @schema/default?

data_values.yml:
     |
   6 |   - secure  #! expecting a map, got a string
     |

     = found: string
     = expected: map (by schema.yml:5)
     = hint: is the default value set using @schema/default?

data_values.yml:
     |
   9 |     - one  #! expecting a float, got a string
     |

     = found: string
     = expected: float (by schema.yml:6)
     = hint: is the default value set using @schema/default?

data_values.yml:
     |
  10 |     - true  #! expecting a float, got a bool
     |

     = found: boolean
     = expected: float (by schema.yml:6)
     = hint: is the default value set using @schema/default?
`
		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when a data value map of the wrong type is passed using template replace", func(t *testing.T) {

		schemaYAML := `#@data/values-schema
---
map:
  nestedMap:
    key: 1
  otherMap: 2
  array:
  - 3
`
		dataValuesYAML := `#@ load("@ytt:template", "template")
#@data/values
---
#@ def frag_func():
key: one
#@ end

_: #@ template.replace({'map': { 'nestedMap': frag_func(), 'otherMap': 'two', 'array': ['three']}})
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.map
`
		expectedErr := `
One or more data values were invalid
====================================

dataValues.yml:
    |
  5 | key: one
    |

    = found: string
    = expected: integer (by schema.yml:5)
    = hint: is the default value set using @schema/default?

dataValues.yml:
    |
  8 | _: #@ template.replace({'map': { 'nestedMap': frag_func(), 'otherMap': 'two', 'array': ['three']}})
    |

    = found: string
    = expected: integer (by schema.yml:6)
    = hint: is the default value set using @schema/default?

dataValues.yml:
    |
  8 | _: #@ template.replace({'map': { 'nestedMap': frag_func(), 'otherMap': 'two', 'array': ['three']}})
    |

    = found: string
    = expected: integer (by schema.yml:8)
    = hint: is the default value set using @schema/default?

`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when a data value array of the wrong type is passed using template replace", func(t *testing.T) {

		schemaYAML := `#@data/values-schema
---
- key: 1
`
		dataValuesYAML := `#@ load("@ytt:template", "template")
#@data/values
---
- #@ template.replace([{'key': 'not an integer'}])
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`
		expectedErr := `
One or more data values were invalid
====================================

dataValues.yml:
    |
  4 | - #@ template.replace([{'key': 'not an integer'}])
    |

    = found: string
    = expected: integer (by schema.yml:3)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when a data value of the wrong type is passed using --data-value", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags.KVsFromStrings = []string{"foo=not an integer"}
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

(data-value arg):
    |
  1 | foo=not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})
	t.Run("when a data value of the wrong type is passed using --data-value-yaml", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 7
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags.KVsFromYAML = []string{"foo=not an integer"}
		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

(data-value-yaml arg):
    |
  1 | foo=not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("when a data value of the wrong type is passed using --data-value-env", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 0
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags = cmdtpl.DataValuesFlags{
			EnvFromStrings: []string{"DVS"},
			EnvironFunc:    func() []string { return []string{"DVS_foo=not an integer"} },
		}

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

(data-values-env arg) DVS:
    |
  1 | DVS_foo=not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})
	t.Run("when a data value of the wrong type is passed using --data-value-env-yaml", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 0
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags = cmdtpl.DataValuesFlags{
			EnvFromYAML: []string{"DVS"},
			EnvironFunc: func() []string { return []string{"DVS_foo=not an integer"} },
		}

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

(data-values-env-yaml arg) DVS:
    |
  1 | DVS_foo=not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("when a data value of the wrong type is passed using --data-value-file", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 0
`

		dvs1 := `not an integer`

		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags = cmdtpl.DataValuesFlags{
			KVsFromFiles: []string{"foo=dvs1.yml"},
			ReadFileFunc: func(path string) ([]byte, error) {
				switch path {
				case "dvs1.yml":
					return []byte(dvs1), nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

(data-value-file arg) foo=dvs1.yml:
    |
  1 | not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("when a data value of the wrong type is passed using --data-values-file", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		schemaYAML := `#@data/values-schema
---
foo: 0
`

		dvs1 := `foo: not an integer`

		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values.foo
`
		cmdOpts.DataValuesFlags = cmdtpl.DataValuesFlags{
			FromFiles: []string{"dvs1.yml"},
			ReadFileFunc: func(path string) ([]byte, error) {
				switch path {
				case "dvs1.yml":
					return []byte(dvs1), nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		expectedErr := `
One or more data values were invalid
====================================

dvs1.yml:
    |
  1 | foo: not an integer
    |

    = found: string
    = expected: integer (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("checks after every data values document is processed (and stops if there was a violation)", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
hostname: ""
`
		dvs1Data := `---
not_in_schema: this should be the only violation reported
`

		dvs2Data := `---
hostname: 14   # wrong type; but will never be caught
`
		templateYAML := `---
rendered: true`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})
		opts.DataValuesFlags = cmdtpl.DataValuesFlags{
			FromFiles: []string{"dvs1.yml", "dvs2.yml"},
			ReadFileFunc: func(path string) ([]byte, error) {
				switch path {
				case "dvs1.yml":
					return []byte(dvs1Data), nil
				case "dvs2.yml":
					return []byte(dvs2Data), nil
				default:
					return nil, fmt.Errorf("Unknown file '%s'", path)
				}
			},
		}

		expectedErrMsg := `Overlaying data values (in following order: additional data values): 
One or more data values were invalid
====================================

dvs1.yml:
    |
  2 | not_in_schema: this should be the only violation reported
    |

    = found: not_in_schema
    = expected: (a key defined in map) (by schema.yml:2)
    = hint: declare data values in schema and override them in a data values document
`
		assertFails(t, filesToProcess, expectedErrMsg, opts)
	})

	t.Run("when schema expects a scalar as an array item, but an array is provided", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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

		expectedErr := `
One or more data values were invalid
====================================

values.yml:
    |
  4 | array: [ [1] ]
    |

    = found: array
    = expected: boolean (by schema.yml:3)
`

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when schema expects a scalar as an array item, but a map is provided", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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

		expectedErr := `
One or more data values were invalid
====================================

values.yml:
    |
  4 | array: [ {a: 1} ]
    |

    = found: map
    = expected: boolean (by schema.yml:3)
`
		assertFails(t, filesToProcess, expectedErr, opts)
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
	t.Run("defaults arrays to an empty list", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
	t.Run("as data values are typed, schema 'fills in' missing parts", func(t *testing.T) {
		t.Run("when a map item is missing, adds it (with its defaults)", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
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
		t.Run("including the entire sub-contents of the missing item", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
clients:
- name: ""
  config:
    args:
    - arg: ""
      value: ""
    #@schema/nullable
    options:
      a: true
      b: false
`
			dataValuesYAML := `#@data/values
---
clients:
- name: foo
`
			templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`

			expected := `rendered:
  clients:
  - name: foo
    config:
      args: []
      options: null
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("even when a data value is marked schema/type any=True", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/type any=True
foo:
  bar:
  - 7
  #@schema/nullable
  bat: I should not be there
`
			templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`
			expected := `foo:
  bar: []
  bat: null
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
		t.Run("when a document is nullable, and then a data value makes it non-null", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
#@schema/nullable
---
foo: 0
bar: Hello, world
`
			dataValuesYAML := `#@data/values
---
foo: 42
`
			templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`
			expected := `rendered:
  foo: 42
  bar: Hello, world
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertSucceeds(t, filesToProcess, expected, opts)
		})
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

func TestSchema_Allows_null_values_via_nullable_annotation(t *testing.T) {
	opts := cmdtpl.NewOptions()
	t.Run("on a document", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
#@schema/nullable
---
optional: true
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`
		expected := `rendered: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("on a map", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
	t.Run("on a map item", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
map:
  #@schema/nullable
  a: 1
  #@schema/nullable
  b: 2
`
		dataValuesYAML := `#@data/values
---
map:
  a: 1
`
		templateYAML := `#@ load("@ytt:data", "data")
---
map: #@ data.values.map
`
		expected := `map:
  a: 1
  b: null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("on an array", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
	t.Run("on an array item", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
---
array:
#@schema/nullable
- ""
`
		dataValuesYAML := `#@data/values
---
array:
- one
- null
- two
-
`
		templateYAML := `#@ load("@ytt:data", "data")
---
array: #@ data.values.array
`
		expected := `array:
- one
- null
- two
- null
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("on a scalar", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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

	t.Run("on a map", func(t *testing.T) {
		t.Run("when any=True", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
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
		t.Run("but not when any=False (meaning values *are* checked against schema)", func(t *testing.T) {
			schemaYAML := `#@data/values-schema
---
#@schema/type any=False
foo: 0
`
			dataValuesYAML := `#@data/values
---
foo: ""
`
			templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo
`
			expected := `Overlaying data values (in following order: dataValues.yml): 
One or more data values were invalid
====================================

dataValues.yml:
    |
  3 | foo: ""
    |

    = found: string
    = expected: integer (by schema.yml:4)
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
				files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
			})

			assertFails(t, filesToProcess, expected, opts)
		})
	})
	t.Run("on an array", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
	t.Run("on a document", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
#@schema/type any=True
---
foo: 13
  
`
		dataValuesYAML := `#@data/values
#@overlay/match-child-defaults missing_ok=True
---
foo: ""
bar: 42
`
		templateYAML := `#@ load("@ytt:data", "data")
---
rendered: #@ data.values
`
		expected := `rendered:
  foo: ""
  bar: 42
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("dataValues.yml", []byte(dataValuesYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
		})

		assertSucceeds(t, filesToProcess, expected, opts)
	})
	t.Run("when any is set on maps and arrays with nested dvs and overlay/replace", func(t *testing.T) {
		schemaYAML := `#@data/values-schema
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
		schemaYAML := `#@data/values-schema
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
}

func TestSchema_Is_scoped_to_a_library(t *testing.T) {
	opts := cmdtpl.NewOptions()

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
#@data/values-schema
---
in_root_library: ""
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
lib_data_values: #@ data.values.in_library`)

		libSchemaData := []byte(`
#@data/values-schema
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
	t.Run("when data values are ref'ed to a child library, they are only checked by that library's schema", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		valuesYAML := []byte(`
#@data/values
---
in_root_library: "expected-root-value"
`)

		schemaYAML := []byte(`
#@data/values-schema
---
in_root_library: "override-root-value"

#@library/ref "@lib@child"
#@data/values-schema
---
in_child_library: "override-lib-child-value"
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child").eval())
---
lib_data_values: #@ data.values.in_library`)

		libSchemaData := []byte(`
#@data/values-schema
---
in_library: 0`)

		libChildConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
lib_child_data_values: #@ data.values.in_child_library`)

		libChildSchemaData := []byte(`
#@data/values-schema
---
in_child_library: "default-child"`)

		expectedYAMLTplData := `lib_child_data_values: override-lib-child-value
---
lib_data_values: 0
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("values.yml", valuesYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaData)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/child/config.yml", libChildConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/_ytt_lib/child/values.yml", libChildSchemaData)),
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
#@data/values-schema
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
#@data/values-schema
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
#@data/values-schema
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
#@data/values-schema
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
#@data/values-schema
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
#@data/values-schema
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
     One or more data values were invalid
     ====================================
     
     values.yml:
         |
       5 | foo: bar
         |
     
         = found: string
         = expected: integer (by _ytt_lib/lib/schema.yml:4)
     `
		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when data value ref'ed to a library is passed using --data-value, it is checked by that library's schema", func(t *testing.T) {
		cmdOpts := cmdtpl.NewOptions()
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@data/values-schema
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
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@data/values-schema
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
		rootYAML := []byte(`
#! root.yml
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").eval())`)

		libSchemaYAML := []byte(`
#! lib/schema.yml
#@data/values-schema
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
     One or more data values were invalid
     ====================================
     
     (data-value arg):
         |
       1 | @lib:foo=42
         |
     
         = found: string
         = expected: integer (by _ytt_lib/lib/schema.yml:5)
`
		assertFails(t, filesToProcess, expectedErr, cmdOpts)
	})

	t.Run("when schema is ref'ed to a library, values specified in the schema are the default data values", func(t *testing.T) {
		rootYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")
--- #@ template.replace(library.get("libby").eval())`)

		rootSchemaYAML := []byte(`
#@library/ref "@libby"
#@data/values-schema
---
foo: from_root_schema
`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
lib_data_values: #@ data.values.foo`)

		expectedYAMLTplData := `lib_data_values: from_root_schema
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", rootSchemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libConfigYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})

	t.Run("when schema is ref'ed to a child library, values specified in the schema are the default data values", func(t *testing.T) {
		rootYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")
--- #@ template.replace(library.get("libby").eval())`)

		rootSchemaYAML := []byte(`
#@library/ref "@libby@child"
#@data/values-schema
---
foo: from_root_schema
`)
		libbyConfigYAML := []byte(`
#@ load("@ytt:data", "data")
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")

--- #@ template.replace(library.get("child").eval())`)

		libbyChildConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
lib_data_values: #@ data.values.foo`)

		expectedYAMLTplData := `lib_data_values: from_root_schema
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", rootSchemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libbyConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/_ytt_lib/child/config.yml", libbyChildConfigYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when schema is ref'ed to a child library that has data values", func(t *testing.T) {
		configYaml := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("libby").eval())`)

		libbyConfigYaml := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("libbychild").eval())
---
should: succeed`)

		schemaYaml := []byte(`
#@library/ref "@libby@libbychild"
#@data/values-schema
---
foo: used
`)
		libChildValuesYaml := []byte(`
#@data/values
---
foo: used
`)

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYaml)),
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", schemaYaml)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libbyConfigYaml)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/_ytt_lib/libbychild/values.yml", libChildValuesYaml)),
		})

		assertSucceeds(t, filesToProcess, `should: succeed
`, opts)
	})

	t.Run("when schema is ref'd to a library, data values are only checked by that library's schema", func(t *testing.T) {
		rootYAML := []byte(`
#@ load("@ytt:library", "library")
#@ load("@ytt:template", "template")
#@ load("@ytt:data", "data")
--- #@ template.replace(library.get("libby").with_data_values({"foo": {"ree": "set from root"}}).eval())
---
root_data_values: #@ data.values`)

		overlayLibSchemaYAML := []byte(`
#@library/ref "@libby"
#@data/values-schema
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
#@data/values-schema
---
foo:
  bar: 3`)

		expectedYAMLTplData := `libby_data_values:
  foo:
    bar: 3
    ree: set from root
---
root_data_values: {}
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("more-schema.yml", overlayLibSchemaYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when schema is programmatically set on a library, data values are checked by that library's schema", func(t *testing.T) {
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
#@ libby = libby.with_data_values_schema(more_schema())
#@ libby = libby.with_data_values({"foo": {"ree": "set from root"}})

--- #@ template.replace(libby.eval())
---
root_data_values: #@ data.values`)

		libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
libby_data_values: #@ data.values`)

		libSchemaYAML := []byte(`
#@data/values-schema
---
foo:
  bar: 3`)

		expectedYAMLTplData := `libby_data_values:
  foo:
    bar: 3
    ree: set from root
---
root_data_values: {}
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("root.yml", rootYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/config.yml", libConfigYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/schema.yml", libSchemaYAML)),
		})

		assertSucceeds(t, filesToProcess, expectedYAMLTplData, opts)
	})
	t.Run("when data values are programmatically set on a library, but library's schema expects an int, type violation is reported", func(t *testing.T) {
		configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("lib").with_data_values({'foo':'4'}).eval())`)

		libSchemaYAML := []byte(`
#@data/values-schema
---
foo: 3`)

		expectedErr := `
- library.eval: Evaluating library 'lib': Overlaying data values (in following order: additional data values): 
    in <toplevel>
      config.yml:4 | --- #@ template.replace(library.get("lib").with_data_values({'foo':'4'}).eval())

    reason:
     One or more data values were invalid
     ====================================
     
     Data value calculated:
         #
         # foo: "4"
         #
     
         = found: string
         = expected: integer (by _ytt_lib/lib/schema.yml:4)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/schema.yml", libSchemaYAML)),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
	t.Run("when data values are programmatically set on a library with imported starlark function, but schema expects int, type violation is reported", func(t *testing.T) {

		schemaYAML := `#@data/values-schema
---
foo: 3
cat: meow
`
		funcslibYAML := `#@ def values():
#@   return {'foo': 'bar', 'cat': 'cow'}
#@ end
`
		configYAML := `#@ load("@ytt:template", "template")
#@ load("funcs.lib.yml", "values")
#@ load("@ytt:library", "library")
--- #@ template.replace(library.get("libby").with_data_values(values()).eval())
`
		expectedErr := `- library.eval: Evaluating library 'libby': Overlaying data values (in following order: additional data values): 
    in <toplevel>
      config.yml:4 | --- #@ template.replace(library.get("libby").with_data_values(values()).eval())

    reason:
     One or more data values were invalid
     ====================================
     
     Data value calculated:
         #
         # foo: "bar"
         #
     
         = found: string
         = expected: integer (by _ytt_lib/libby/schema.yml:3)
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/libby/schema.yml", []byte(schemaYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(configYAML))),
			files.MustNewFileFromSource(files.NewBytesSource("funcs.lib.yml", []byte(funcslibYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

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
		templateYAML := `#@ load("@ytt:data", "data")
---
foo: #@ data.values.foo`
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
			files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(templateYAML))),
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
}

func TestSchema_With_fuzzed_inputs(t *testing.T) {
	opts := cmdtpl.NewOptions()

	validIntegerRange := fuzz.UnicodeRange{First: '0', Last: '9'}
	randSource := getYttRandSource(t)

	fuzzLargeNumber := fuzz.New().RandSource(randSource).Funcs(func(s *string, c fuzz.Continue) {
		validIntegerRange.CustomStringFuzzFunc()(s, c)
		// We remove '0' in the prefix to only test base 10 numbers.
		// For more info refer to the yaml spec: http://yaml.org/type/int.html
		removePrefix(s, "0")

		if *s == "" {
			*s = strconv.Itoa(c.Int())
		}
	})

	fuzzFloat := fuzz.New().RandSource(randSource).Funcs(func(s *string, c fuzz.Continue) {
		*s = strconv.FormatFloat(c.Float64(), 'f', -1, 64)

	})

	fuzzStrings := fuzz.New().RandSource(randSource).Funcs(func(s *string, c fuzz.Continue) {
		*s += c.RandString()
		*s = strings.ReplaceAll(*s, "'", `"`)
		// starlark uses the '\' char as an escape character. ignore the escape char to simplify writing assertions.
		*s = strings.ReplaceAll(*s, "\\", `/`)
	})

	for i := 0; i < 100; i++ {
		var expectedInt, expectedString, expectedFloat string
		fuzzLargeNumber.Fuzz(&expectedInt)
		fuzzStrings.Fuzz(&expectedString)
		fuzzFloat.Fuzz(&expectedFloat)
		starlarkEvals := []string{"", "#@ "}
		starlarkEvalUsed := starlarkEvals[rand.New(randSource).Intn(2)]

		t.Run(fmt.Sprintf("A schema programatically set to a library: int: [%v], string: [%v], float64: [%v], starlark eval: [%v]", expectedInt, expectedString, expectedFloat, starlarkEvalUsed), func(t *testing.T) {

			configYAML := []byte(`
#@ load("@ytt:template", "template")
#@ load("@ytt:library", "library")
---
#@ def dvs_from_root():
someInt: ` + starlarkEvalUsed + expectedInt + `
someString: ` + starlarkEvalUsed + "'" + expectedString + "'" + `
someFloat: ` + starlarkEvalUsed + expectedFloat + `
#@ end
--- #@ template.replace(library.get("lib").with_data_values_schema(dvs_from_root()).eval())`)

			libConfigYAML := []byte(`
#@ load("@ytt:data", "data")
---
someInt: #@ data.values.someInt
someString: #@ data.values.someString
someFloat: #@ data.values.someFloat
`)

			expectedFloatParsed, err := strconv.ParseFloat(expectedFloat, 64)
			require.NoError(t, err)
			expectedYAMLTplData := `someInt: ` + expectedInt + `
someString: ('|")*` + regexp.QuoteMeta(expectedString) + `('|")*
someFloat: ` + "(" + expectedFloat + "|" + fmt.Sprintf("%g", expectedFloatParsed) + ")" + `
`

			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("config.yml", configYAML)),
				files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/lib/config.yml", libConfigYAML)),
			})

			assertSucceedsWithRegexp(t, filesToProcess, expectedYAMLTplData, opts)
		})

	}
}

func getYttRandSource(t *testing.T) rand.Source {
	var seed int64
	if os.Getenv("YTT_SEED") == "" {
		seed = time.Now().UnixNano()
	} else {
		envSeed, err := strconv.Atoi(os.Getenv("YTT_SEED"))
		require.NoError(t, err)
		seed = int64(envSeed)
	}

	t.Log(fmt.Sprintf("YTT Seed used was: [%v]. To reproduce this test failure, re-run the test with `export YTT_SEED=%v`", seed, seed))

	t.Cleanup(func() {
		fmt.Printf("\n\n*** To reproduce this test run, re-run the test with `export YTT_SEED=%v` ***\n\n", seed)
	})

	return rand.NewSource(seed)
}

func removePrefix(s *string, prefix string) {
	for {
		if strings.HasPrefix(*s, prefix) {
			*s = strings.TrimPrefix(*s, prefix)
		} else {
			break
		}
	}
}

func assertSucceeds(t *testing.T, filesToProcess []*files.File, expectedOut string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	require.Len(t, out.Files, 1, "unexpected number of output files")

	require.Equal(t, expectedOut, string(out.Files[0].Bytes()))
}

func assertSucceedsWithRegexp(t *testing.T, filesToProcess []*files.File, expectedOutRegex string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	require.Len(t, out.Files, 1, "unexpected number of output files")

	require.Regexp(t, expectedOutRegex, string(out.Files[0].Bytes()))
}

func assertFails(t *testing.T, filesToProcess []*files.File, expectedErr string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.Error(t, out.Err)

	require.Contains(t, out.Err.Error(), expectedErr)
}
