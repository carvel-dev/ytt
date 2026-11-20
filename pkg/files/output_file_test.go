// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestOutputFileCreatedWithoutExecutePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission bits are not meaningful on Windows")
	}

	dir := t.TempDir()

	out := files.NewOutputFile(
		"config.yml", []byte("foo: bar\n"), files.TypeYAML)
	err := out.Create(dir)
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(dir, "config.yml"))
	require.NoError(t, err)

	require.Equal(t, "-rw-------", info.Mode().String())
}
