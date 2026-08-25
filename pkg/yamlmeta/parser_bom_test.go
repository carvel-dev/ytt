// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"testing"

	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/stretchr/testify/require"
)

const utf8BOM = "\xEF\xBB\xBF"

func parseForBOMTest(t *testing.T, data string) *yamlmeta.DocumentSet {
	t.Helper()

	opts := yamlmeta.ParserOpts{WithoutComments: false}
	docSet, err := yamlmeta.NewParser(opts).ParseBytes([]byte(data), "t.yml")
	require.NoError(t, err)

	return docSet
}

func printForBOMTest(docSet *yamlmeta.DocumentSet) string {
	opts := yamlmeta.PrinterOpts{ExcludeRefs: true}
	return yamlmeta.NewPrinterWithOpts(nil, opts).PrintStr(docSet)
}

// A UTF-8 BOM marks the encoding of a stream; it is not part of the first
// key. Left in place it becomes a leading U+FEFF on that key, which then
// renders as a quoted, escaped key in the output.
func TestParserSkipsUTF8BOM(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "plain mapping", data: "test: null\n"},
		{name: "document marker", data: "---\ntest: null\n"},
		{name: "leading comment", data: "#! a comment\ntest: null\n"},
		{name: "two documents", data: "test: null\n---\nsecond: null\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withBOM := parseForBOMTest(t, utf8BOM+tc.data)
			withoutBOM := parseForBOMTest(t, tc.data)

			require.Equal(t,
				printForBOMTest(withoutBOM), printForBOMTest(withBOM),
				"a leading UTF-8 BOM changed how the document parsed")
		})
	}
}

// The document marker check runs on the raw input, so a BOM in front of
// "---" hides the marker. The parser then prepends its own marker, which
// shifts reported positions and can make the input fail to parse at all.
func TestParserPositionsAreCorrectAfterUTF8BOM(t *testing.T) {
	const data = "---\ntest: null\n"

	withBOM := parseForBOMTest(t, utf8BOM+data)
	withoutBOM := parseForBOMTest(t, data)

	require.Equal(t, len(withoutBOM.Items), len(withBOM.Items),
		"a leading UTF-8 BOM changed the number of parsed documents")

	wantPos := withoutBOM.Items[0].Position.AsIntString()
	require.Equal(t, wantPos, withBOM.Items[0].Position.AsIntString(),
		"a leading UTF-8 BOM shifted the reported line number")
}
