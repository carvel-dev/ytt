// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"testing"
)

// TestMainWorks ensures that the example provided actually works.
func TestMainWorks(t *testing.T) {
	var b bytes.Buffer
	run(&b)

	if b.String() != ("greeting: Hello, world\n") {
		t.FailNow()
	}
}
