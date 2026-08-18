// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

const utf8BOM = "\xEF\xBB\xBF"

// A file saved with a UTF-8 BOM should render exactly like the same file
// without one. Before this was handled the BOM was read as part of the first
// key, so `test: #@ None` rendered as a quoted, escaped key instead of `test`.
func TestUTF8BOMRendersTheSameAsNoBOM(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "annotated value",
			template: "test: #@ None\n",
			expected: "test: null\n",
		},
		{
			name:     "explicit document marker",
			template: "---\ntest: 1\n",
			expected: "test: 1\n",
		},
		{
			name:     "ytt comment first",
			template: "#! a comment\ntest: 1\n",
			expected: "test: 1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withoutBOM := renderOneFile(t, []byte(tc.template))
			require.Equal(t, tc.expected, withoutBOM,
				"the control case rendered unexpectedly")

			withBOM := renderOneFile(t, []byte(utf8BOM+tc.template))
			require.Equal(t, withoutBOM, withBOM,
				"a leading UTF-8 BOM changed the output")
		})
	}
}

// Only a leading BOM is a stream marker. One part way through a file is
// ordinary content, and the emitter keeps it by quoting and escaping it.
func TestUTF8BOMInsideAFileIsLeftAlone(t *testing.T) {
	rendered := renderOneFile(t, []byte("test: \"a"+utf8BOM+"b\"\n"))
	require.Equal(t, "test: \"a\\uFEFFb\"\n", rendered)
}

func TestUTF8BOMStarlarkLibraryLoadsTheSameAsNoBOM(t *testing.T) {
	renderWithLibrary := func(t *testing.T, libraryData []byte) string {
		t.Helper()

		filesToProcess := []*files.File{
			files.MustNewFileFromSource(files.NewBytesSource(
				"data.yml", []byte("#@ load(\"values.star\", \"value\")\nresult: #@ value\n"))),
			files.MustNewFileFromSource(files.NewBytesSource("values.star", libraryData)),
		}

		out := cmdtpl.NewOptions().RunWithFiles(
			cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
		require.NoError(t, out.Err)
		require.Len(t, out.Files, 1, "unexpected number of output files")

		return string(out.Files[0].Bytes())
	}

	const library = "value = \"from starlark\"\n"
	withoutBOM := renderWithLibrary(t, []byte(library))
	require.Equal(t, "result: from starlark\n", withoutBOM,
		"the control case rendered unexpectedly")

	withBOM := renderWithLibrary(t, []byte(utf8BOM+library))
	require.Equal(t, withoutBOM, withBOM,
		"a leading UTF-8 BOM changed how a Starlark library loaded")
}

func TestUTF8BOMTextTemplateRendersTheSameAsNoBOM(t *testing.T) {
	const textTemplate = "result: (@= \"from text template\" @)\n"

	withoutBOM := renderOneNamedFile(t, "data.txt", []byte(textTemplate))
	require.Equal(t, "result: from text template\n", withoutBOM,
		"the control case rendered unexpectedly")

	withBOM := renderOneNamedFile(t, "data.txt", []byte(utf8BOM+textTemplate))
	require.Equal(t, withoutBOM, withBOM,
		"a leading UTF-8 BOM changed the text-template output")
}

func renderOneFile(t *testing.T, data []byte) string {
	return renderOneNamedFile(t, "data.yml", data)
}

func renderOneNamedFile(t *testing.T, name string, data []byte) string {
	t.Helper()

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource(name, data)),
	}

	input := cmdtpl.Input{Files: filesToProcess}
	out := cmdtpl.NewOptions().RunWithFiles(input, ui.NewTTY(false))
	require.NoError(t, out.Err)
	require.Len(t, out.Files, 1, "unexpected number of output files")

	return string(out.Files[0].Bytes())
}
