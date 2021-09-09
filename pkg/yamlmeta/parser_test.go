// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/k14s/difflib"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var _ = fmt.Sprintf

func TestParserDocSetEmpty(t *testing.T) {
	const data = ""

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetNewline(t *testing.T) {
	const data = "\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserOnlyComment(t *testing.T) {
	const data = "#"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: "", Position: filepos.NewPosition(1)},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	if parsedValStr != expectedValStr {
		t.Fatalf("not equal\nparsed:\n>>>%s<<<expected:\n>>>%s<<<", parsedValStr, expectedValStr)
	}
}

func TestParserDoc(t *testing.T) {
	const data = "---\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocWithoutDashes(t *testing.T) {
	const data = "key: 1\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{Key: "key", Value: 1, Position: filepos.NewPosition(1)},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserRootValue(t *testing.T) {
	parserExamples{
		{Description: "string", Data: "abc",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value:    "abc",
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		{Description: "integer", Data: "1",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value:    1,
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		{Description: "float", Data: "2000.1",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value:    2000.1,
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		{Description: "float (exponent)", Data: "9e3",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value:    9000.0,
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		{Description: "array", Data: "- 1",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value: &yamlmeta.Array{
							Items: []*yamlmeta.ArrayItem{
								&yamlmeta.ArrayItem{Value: 1, Position: filepos.NewPosition(1)},
							},
							Position: filepos.NewPosition(1),
						},
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		{Description: "map", Data: "key: val",
			Expected: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{
					&yamlmeta.Document{
						Value: &yamlmeta.Map{
							Items: []*yamlmeta.MapItem{
								&yamlmeta.MapItem{Key: "key", Value: "val", Position: filepos.NewPosition(1)},
							},
							Position: filepos.NewPosition(1),
						},
						Position: filepos.NewPosition(1),
					},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
	}.Check(t)
}

func TestParserRootString(t *testing.T) {
	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment", Position: filepos.NewPosition(1)},
				},
				Value:    "abc",
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	parserExamples{
		// TODO should really be owned by abc
		{Description: "single line", Data: "--- abc # comment", Expected: expectedVal},
		{Description: "common on doc", Data: "--- # comment\nabc", Expected: expectedVal},
		// TODO add *yamlmeta.Value
		// {"comment on value", "---\nabc # comment", expectedVal},
	}.Check(t)
}

func TestParserMapArray(t *testing.T) {
	const data = `---
array:
- 1
- 2
- key: value
`

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Key: "array",
							Value: &yamlmeta.Array{
								Items: []*yamlmeta.ArrayItem{
									&yamlmeta.ArrayItem{Value: 1, Position: filepos.NewPosition(3)},
									&yamlmeta.ArrayItem{Value: 2, Position: filepos.NewPosition(4)},
									&yamlmeta.ArrayItem{
										Value: &yamlmeta.Map{
											Items: []*yamlmeta.MapItem{
												&yamlmeta.MapItem{
													Key:      "key",
													Value:    "value",
													Position: filepos.NewPosition(5),
												},
											},
											Position: filepos.NewPosition(5),
										},
										Position: filepos.NewPosition(5),
									},
								},
								Position: filepos.NewPosition(2),
							},
							Position: filepos.NewPosition(2),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserMapComments(t *testing.T) {
	const data = `---
# before-map
map:
  # before-key1
  key1: val1 # inline-key1
  # after-key1
  # before-key2
  key2: val2
`

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Comments: []*yamlmeta.Comment{
								&yamlmeta.Comment{Data: " before-map", Position: filepos.NewPosition(2)},
							},
							Key: "map",
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " before-key1", Position: filepos.NewPosition(4)},
											&yamlmeta.Comment{Data: " inline-key1", Position: filepos.NewPosition(5)},
										},
										Key:      "key1",
										Value:    "val1",
										Position: filepos.NewPosition(5),
									},
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " after-key1", Position: filepos.NewPosition(6)},
											&yamlmeta.Comment{Data: " before-key2", Position: filepos.NewPosition(7)},
										},
										Key:      "key2",
										Value:    "val2",
										Position: filepos.NewPosition(8),
									},
								},
								Position: filepos.NewPosition(3),
							},
							Position: filepos.NewPosition(3),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserArrayComments(t *testing.T) {
	const data = `---
array:
# before-1
- 1 # inline-1
# after-1
# before-2
- 2
- 3
- # empty
- 
  # on-map
  key: value
# on-array-item-with-map
- key: value
`

	// TODO comment on top of scalar

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Key: "array",
							Value: &yamlmeta.Array{
								Items: []*yamlmeta.ArrayItem{
									&yamlmeta.ArrayItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " before-1", Position: filepos.NewPosition(3)},
											&yamlmeta.Comment{Data: " inline-1", Position: filepos.NewPosition(4)},
										},
										Value:    1,
										Position: filepos.NewPosition(4),
									},
									&yamlmeta.ArrayItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " after-1", Position: filepos.NewPosition(5)},
											&yamlmeta.Comment{Data: " before-2", Position: filepos.NewPosition(6)},
										},
										Value:    2,
										Position: filepos.NewPosition(7),
									},
									&yamlmeta.ArrayItem{Value: 3, Position: filepos.NewPosition(8)},
									&yamlmeta.ArrayItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " empty", Position: filepos.NewPosition(9)},
										},
										Value:    nil,
										Position: filepos.NewPosition(9),
									},
									&yamlmeta.ArrayItem{
										Value: &yamlmeta.Map{
											Items: []*yamlmeta.MapItem{
												&yamlmeta.MapItem{
													Comments: []*yamlmeta.Comment{
														&yamlmeta.Comment{Data: " on-map", Position: filepos.NewPosition(11)},
													},
													Key:      "key",
													Value:    "value",
													Position: filepos.NewPosition(12),
												},
											},
											Position: filepos.NewPosition(10),
										},
										Position: filepos.NewPosition(10),
									},
									&yamlmeta.ArrayItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " on-array-item-with-map", Position: filepos.NewPosition(13)},
										},
										Value: &yamlmeta.Map{
											Items: []*yamlmeta.MapItem{
												&yamlmeta.MapItem{
													Key:      "key",
													Value:    "value",
													Position: filepos.NewPosition(14),
												},
											},
											Position: filepos.NewPosition(14),
										},
										Position: filepos.NewPosition(14),
									},
								},
								Position: filepos.NewPosition(2),
							},
							Position: filepos.NewPosition(2),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetComments(t *testing.T) {
	const data = `---
# comment-first
---
---
# comment-second
`

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-first", Position: filepos.NewPosition(2)},
				},
				Position: filepos.NewPosition(3),
			},
			&yamlmeta.Document{
				Position: filepos.NewPosition(4),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-second", Position: filepos.NewPosition(5)},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetOnlyComments2(t *testing.T) {
	const data = "---\n# comment-first\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-first", Position: filepos.NewPosition(2)},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetOnlyComments3(t *testing.T) {
	const data = "--- # comment\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment", Position: filepos.NewPosition(1)},
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetOnlyComments(t *testing.T) {
	const data = "# comment-first\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-first", Position: filepos.NewPosition(1)},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserDocSetCommentsNoFirstDashes(t *testing.T) {
	const data = `# comment-first
---
---
# comment-second
`

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Position: filepos.NewPosition(1),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-first", Position: filepos.NewPosition(1)},
				},
				Position: filepos.NewPosition(2),
			},
			&yamlmeta.Document{
				Position: filepos.NewPosition(3),
			},
			&yamlmeta.Document{
				Comments: []*yamlmeta.Comment{
					&yamlmeta.Comment{Data: " comment-second", Position: filepos.NewPosition(4)},
				},
				Position: filepos.NewUnknownPosition(),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserUnindentedComment(t *testing.T) {
	const data = `---
key:
  nested: true
# comment
  nested: true
`

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Key: "key",
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										Key:      "nested",
										Value:    true,
										Position: filepos.NewPosition(3),
									},
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: " comment", Position: filepos.NewPosition(4)},
										},
										Key:      "nested",
										Value:    true,
										Position: filepos.NewPosition(5),
									},
								},
								Position: filepos.NewPosition(2),
							},
							Position: filepos.NewPosition(2),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(expectedVal)

	assertEqual(t, parsedValStr, expectedValStr)
}

func TestParserInvalidDoc(t *testing.T) {
	parserExamples{
		{Description: "no doc marker",
			Data:        "apiVersion: @123",
			ExpectedErr: "yaml: line 1: found character that cannot start any token",
		},
		{Description: "doc marker",
			Data:        "---\napiVersion: @123",
			ExpectedErr: "yaml: line 2: found character that cannot start any token",
		},
		{Description: "space before",
			Data:        "\n\n\napiVersion: @123",
			ExpectedErr: "yaml: line 4: found character that cannot start any token",
		},
		{Description: "doc marker with space",
			Data:        "\n\n---\napiVersion: @123",
			ExpectedErr: "yaml: line 4: found character that cannot start any token",
		},
	}.Check(t)
}

func TestParserAnchors(t *testing.T) {
	data := `
#@ variable = 123
value: &value
  path: #@ variable
  #@annotation
  args:
  - 1
  - 2
anchored_value: *value
`

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Key: "value",
							Comments: []*yamlmeta.Comment{
								&yamlmeta.Comment{Data: "@ variable = 123", Position: filepos.NewPosition(2)},
							},
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										// TODO should be here as well
										// Comments: []*yamlmeta.Comment{
										// 	&yamlmeta.Comment{Data: "@ variable", Position: filepos.NewPosition(4)},
										// },
										Key:      "path",
										Value:    nil,
										Position: filepos.NewPosition(4),
									},
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: "@annotation", Position: filepos.NewPosition(5)},
										},
										Key: "args",
										Value: &yamlmeta.Array{
											Items: []*yamlmeta.ArrayItem{
												&yamlmeta.ArrayItem{
													Value:    1,
													Position: filepos.NewPosition(7),
												},
												&yamlmeta.ArrayItem{
													Value:    2,
													Position: filepos.NewPosition(8),
												},
											},
											Position: filepos.NewPosition(6),
										},
										Position: filepos.NewPosition(6),
									},
								},
								Position: filepos.NewPosition(3),
							},
							Position: filepos.NewPosition(3),
						},
						&yamlmeta.MapItem{
							Key: "anchored_value",
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: "@ variable", Position: filepos.NewPosition(4)},
										},
										Key:      "path",
										Value:    nil,
										Position: filepos.NewPosition(4),
									},
									&yamlmeta.MapItem{
										// TODO should be here as well
										// Comments: []*yamlmeta.Comment{
										// 	&yamlmeta.Comment{Data: "@annotation", Position: filepos.NewPosition(5)},
										// },
										Key: "args",
										Value: &yamlmeta.Array{
											Items: []*yamlmeta.ArrayItem{
												&yamlmeta.ArrayItem{
													Value:    1,
													Position: filepos.NewPosition(7),
												},
												&yamlmeta.ArrayItem{
													Value:    2,
													Position: filepos.NewPosition(8),
												},
											},
											Position: filepos.NewPosition(6),
										},
										Position: filepos.NewPosition(6),
									},
								},
								Position: filepos.NewPosition(9),
							},
							Position: filepos.NewPosition(9),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	// TODO annotations are not properly assigned
	parserExamples{{Description: "with seq inside anchored data", Data: data, Expected: expectedVal}}.Check(t)
}

func TestParserMergeOp(t *testing.T) {
	data := `
#@ variable = 123
value: &value
  path: #@ variable
  #@annotation
  args:
  - 1
  - 2
merged_value:
  <<: *value
  other: true
`

	expectedVal := &yamlmeta.DocumentSet{
		Items: []*yamlmeta.Document{
			&yamlmeta.Document{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						&yamlmeta.MapItem{
							Key: "value",
							Comments: []*yamlmeta.Comment{
								&yamlmeta.Comment{Data: "@ variable = 123", Position: filepos.NewPosition(2)},
							},
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										// TODO should be here as well
										// Comments: []*yamlmeta.Comment{
										// 	&yamlmeta.Comment{Data: "@ variable", Position: filepos.NewPosition(4)},
										// },
										Key:      "path",
										Value:    nil,
										Position: filepos.NewPosition(4),
									},
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: "@annotation", Position: filepos.NewPosition(5)},
										},
										Key: "args",
										Value: &yamlmeta.Array{
											Items: []*yamlmeta.ArrayItem{
												&yamlmeta.ArrayItem{
													Value:    1,
													Position: filepos.NewPosition(7),
												},
												&yamlmeta.ArrayItem{
													Value:    2,
													Position: filepos.NewPosition(8),
												},
											},
											Position: filepos.NewPosition(6),
										},
										Position: filepos.NewPosition(6),
									},
								},
								Position: filepos.NewPosition(3),
							},
							Position: filepos.NewPosition(3),
						},
						&yamlmeta.MapItem{
							Key: "merged_value",
							Value: &yamlmeta.Map{
								Items: []*yamlmeta.MapItem{
									&yamlmeta.MapItem{
										Comments: []*yamlmeta.Comment{
											&yamlmeta.Comment{Data: "@ variable", Position: filepos.NewPosition(4)},
										},
										Key:      "path",
										Value:    nil,
										Position: filepos.NewPosition(4),
									},
									&yamlmeta.MapItem{
										// TODO should be here as well
										// Comments: []*yamlmeta.Comment{
										// 	&yamlmeta.Comment{Data: "@annotation", Position: filepos.NewPosition(5)},
										// },
										Key: "args",
										Value: &yamlmeta.Array{
											Items: []*yamlmeta.ArrayItem{
												&yamlmeta.ArrayItem{
													Value:    1,
													Position: filepos.NewPosition(7),
												},
												&yamlmeta.ArrayItem{
													Value:    2,
													Position: filepos.NewPosition(8),
												},
											},
											Position: filepos.NewPosition(6),
										},
										Position: filepos.NewPosition(6),
									},
									&yamlmeta.MapItem{
										Key:      "other",
										Value:    true,
										Position: filepos.NewPosition(11),
									},
								},
								Position: filepos.NewPosition(9),
							},
							Position: filepos.NewPosition(9),
						},
					},
					Position: filepos.NewPosition(1),
				},
				Position: filepos.NewPosition(1),
			},
		},
		Position: filepos.NewUnknownPosition(),
	}

	// TODO annotations are not properly assigned
	parserExamples{{Description: "merge", Data: data, Expected: expectedVal}}.Check(t)
}

func TestParserDocWithoutDashesPosition(t *testing.T) {
	const data = "key: 1\n"

	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(data), "data.yml")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	parsedPosStr := parsedVal.Items[0].Position.AsString()
	expectedPosStr := "line data.yml:1"

	if parsedPosStr != expectedPosStr {
		t.Fatalf("not equal\nparsed...: %s\nexpected.: %s\n", parsedPosStr, expectedPosStr)
	}
}

type parserExamples []parserExample

func (exs parserExamples) Check(t *testing.T) {
	for _, ex := range exs {
		ex.Check(t)
	}
}

type parserExample struct {
	Description string
	Data        string
	Expected    *yamlmeta.DocumentSet
	ExpectedErr string
}

func (ex parserExample) Check(t *testing.T) {
	parsedVal, err := yamlmeta.NewParser(yamlmeta.ParserOpts{WithoutComments: false}).ParseBytes([]byte(ex.Data), "")
	if len(ex.ExpectedErr) == 0 {
		ex.checkDocSet(t, parsedVal, err)
	} else {
		ex.checkErr(t, err)
	}
}

func (ex parserExample) checkDocSet(t *testing.T, parsedVal *yamlmeta.DocumentSet, err error) {
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	printer := yamlmeta.NewPrinterWithOpts(os.Stdout, yamlmeta.PrinterOpts{ExcludeRefs: true})

	parsedValStr := printer.PrintStr(parsedVal)
	expectedValStr := printer.PrintStr(ex.Expected)

	assertEqual(t, parsedValStr, expectedValStr)
}

func (ex parserExample) checkErr(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("expected error")
	}

	parsedValStr := err.Error()
	expectedValStr := ex.ExpectedErr

	assertEqual(t, parsedValStr, expectedValStr)
}

func assertEqual(t *testing.T, parsedValStr string, expectedValStr string) {
	if parsedValStr != expectedValStr {
		t.Fatalf("Not equal; diff expected...actual:\n%v\n", difflib.PPDiff(strings.Split(expectedValStr, "\n"), strings.Split(parsedValStr, "\n")))
	}
}
