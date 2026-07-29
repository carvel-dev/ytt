// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package ui_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"carvel.dev/ytt/pkg/cmd/ui"
)

func TestIsDebugReportsDebugMode(t *testing.T) {
	require.True(t, ui.NewTTY(true).IsDebug())
	require.False(t, ui.NewTTY(false).IsDebug())
}
