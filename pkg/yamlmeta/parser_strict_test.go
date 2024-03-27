// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/yamlmeta"
)

var _ = fmt.Sprintf

func TestParserStrict(t *testing.T) {
	parserStrictExamples{
		{Description: "ambigious string that looks like map kv",
			Data:        "foo:\n  blah:key",
			ExpectedErr: "yaml: Strict parsing: Strings with colon must be explicitly quoted: 'blah:key'",
		},
		{Description: "ambigious string that looks like map kv",
			Data:        "foo:\n  blah:key",
			ExpectedErr: "yaml: Strict parsing: Strings with colon must be explicitly quoted: 'blah:key'",
		},
		{Description: "multi line string explicit", Data: "|\n  123\n  123", ExpectedVal: "123\n123"},
		{Description: "multi line strings that doesnt look like single string",
			Data:        "100\n100",
			ExpectedErr: "yaml: Strict parsing: Strings with whitespace must be explicitly quoted: '100 100'",
		},
		{Description: "ambigious booleans",
			Data:        "YES",
			ExpectedErr: "yaml: Strict parsing: Found 'YES' ambigious (could be !!str or !!bool)",
		},
		{Description: "ambigious null",
			Data:        "~",
			ExpectedErr: "yaml: Strict parsing: Found '~' ambigious (could be !!str or !!null)",
		},
		{Description: "ambigious NaN",
			Data:        ".nan",
			ExpectedErr: "yaml: Strict parsing: Found '.nan' ambigious (could be !!str or !!float)",
		},
		{Description: "ambigious number starting with 0",
			Data:        "011",
			ExpectedErr: "yaml: Strict parsing: Found '011' ambigious (could be !!str or !!int)",
		},
		{Description: "plain int", Data: "123", ExpectedVal: 123},
		{Description: "plain int 9-0", Data: "91230", ExpectedVal: 91230},
		{Description: "plain int negative", Data: "-123", ExpectedVal: -123},
		{Description: "plain int zero", Data: "0", ExpectedVal: 0},
		{Description: "plain int64", Data: "9223372036854775807", ExpectedVal: 9223372036854775807},
		{Description: "plain int64 negative", Data: "-9223372036854775808", ExpectedVal: -9223372036854775808},
		{Description: "plain float zero", Data: "0.0", ExpectedVal: float64(0)},
		{Description: "plain float", Data: "0.123", ExpectedVal: 0.123},
		{Description: "plain float 9-0", Data: "9230.9230", ExpectedVal: 9230.9230},
		{Description: "plain float negative", Data: "-0.1", ExpectedVal: -0.1},
		{Description: "ambigious float",
			Data:        ".1",
			ExpectedErr: "yaml: Strict parsing: Found '.1' ambigious (could be !!str or !!float)",
		},
		{Description: "ambigious binary repr",
			Data:        "0b0001",
			ExpectedErr: "yaml: Strict parsing: Found '0b0001' ambigious (could be !!str or !!int)",
		},
		{Description: "ambigious binary repr",
			Data:        "-0b0001",
			ExpectedErr: "yaml: Strict parsing: Found '-0b0001' ambigious (could be !!str or !!int)",
		},
		{Description: "tagged int",
			Data:        "!!int 0b0010",
			ExpectedVal: 2,
		},
		{Description: "ambigious document start",
			Data:        "---key: value",
			ExpectedErr: "yaml: Strict parsing: Strings with triple-dash must be explicitly quoted: '---key'",
		},
		{Description: "ambigious document start middle",
			Data:        "prevval---key: value",
			ExpectedErr: "yaml: Strict parsing: Strings with triple-dash must be explicitly quoted: 'prevval---key'",
		},
	}.Check(t)
}

type parserStrictExamples []parserStrictExample

func (exs parserStrictExamples) Check(t *testing.T) {
	for _, ex := range exs {
		ex.Check(t)
	}
}

type parserStrictExample struct {
	Data        string
	ExpectedErr string
	ExpectedVal interface{}
	Description string
}

func (ex parserStrictExample) Check(t *testing.T) {
	docSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{Strict: true}).ParseBytes([]byte(ex.Data), "")
	if len(ex.ExpectedErr) == 0 {
		ex.checkVal(t, docSet, err)
	} else {
		ex.checkErr(t, err)
	}
}

func (ex parserStrictExample) checkVal(t *testing.T, docSet *yamlmeta.DocumentSet, err error) {
	if err != nil {
		t.Fatalf("[%s] expected no error, but was %s", ex.Description, err)
	}
	if len(docSet.Items) != 1 {
		t.Fatalf("[%s] expected one doc", ex.Description)
	}

	val := docSet.Items[0].AsInterface()

	// Assume that all examples are json
	valBs, _ := json.Marshal(orderedmap.Conversion{val}.AsUnorderedStringMaps())
	expectedValBs, _ := json.Marshal(ex.ExpectedVal)

	if !reflect.DeepEqual(valBs, expectedValBs) {
		t.Fatalf("[%s] not equal\nparsed:\n%#v\nexpected:\n%#v", ex.Description, valBs, expectedValBs)
	}
}

func (ex parserStrictExample) checkErr(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("[%s] expected error", ex.Description)
	}

	parsedValStr := err.Error()
	expectedValStr := ex.ExpectedErr

	if parsedValStr != expectedValStr {
		t.Fatalf("[%s] not equal\nparsed:\n%s\nexpected:\n%s", ex.Description, parsedValStr, expectedValStr)
	}
}
