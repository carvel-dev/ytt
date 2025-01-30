// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSortedFilesFromPaths(t *testing.T) {
	inputPaths := []string{
		"../yamltemplate/filetests/def.tpltest",
		"https://example.org/test.yaml?hello=world&foo=bar",
		"an/overwritten/path/to/file.yaml=https://example.org/test.yaml?hello=world&foo=bar"}
	filesToProcess, err := files.NewSortedFilesFromPaths(inputPaths, files.SymlinkAllowOpts{true, []string{}})

	require.NoError(t, err)

	assert.Equal(t, filesToProcess[0].RelativePath(), "def.tpltest")
	assert.Equal(t, filesToProcess[1].RelativePath(), "test.yaml")
	assert.Equal(t, filesToProcess[1].Description(), "HTTP URL 'https://example.org/test.yaml?hello=world&foo=bar'")
	assert.Equal(t, filesToProcess[2].RelativePath(), "an/overwritten/path/to/file.yaml")
	assert.Equal(t, filesToProcess[2].Description(), "HTTP URL 'https://example.org/test.yaml?hello=world&foo=bar'")
}
