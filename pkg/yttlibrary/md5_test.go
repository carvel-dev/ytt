// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"bytes"
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/cmd/ui"
	"github.com/k14s/starlark-go/starlark"
)

func TestMD5SumEmitsDeprecationWarningOnce(t *testing.T) {
	hasWarnedMD5Deprecated.Store(false)
	defer hasWarnedMD5Deprecated.Store(false)

	var stderr bytes.Buffer
	mod := NewMD5Module(ui.NewCustomWriterTTY(false, nil, &stderr))
	sumFunc := mod.warnOnCall(mod.sum)
	args := starlark.Tuple{starlark.String("data")}

	result, err := sumFunc(nil, nil, args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSum := starlark.String("8d777f385d3dfec8815d20f7496026dc")
	if result != expectedSum {
		t.Errorf("expected sum %q, got %q", expectedSum, result)
	}

	if !strings.Contains(stderr.String(), "@ytt:md5 module is deprecated") {
		t.Errorf("expected deprecation warning on stderr, got: %q",
			stderr.String())
	}

	// A second call must not repeat the warning.
	stderr.Reset()
	if _, err := sumFunc(nil, nil, args, nil); err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no warning on second call, got: %q", stderr.String())
	}
}
