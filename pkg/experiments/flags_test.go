// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package experiments

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
At runtime, there is a singleton instance of experiments flags.
This test suite verifies the behavior of of such an instance.
To avoid test pollution, a fresh instance is created in each test.
*/

func TestFeaturesAreDisabledByDefault(t *testing.T) {
	ResetForTesting()
	os.Setenv(Env, "")
	assert.False(t, isNoopEnabled())
}

func TestFeaturesCanBeEnabled(t *testing.T) {
	ResetForTesting()
	os.Setenv(Env, "noop")
	assert.True(t, isNoopEnabled())
}
