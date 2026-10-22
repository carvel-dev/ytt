// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diagnostics

import (
	"errors"
	"net/url"
	"testing"
)

func TestURLPresentation(t *testing.T) {
	raw := "https://user:password-canary@example.com/input" +
		"?access_token=query-canary#fragment-canary"
	if got := URL(raw); got != "https://example.com/input" {
		t.Fatalf("unsafe URL: %s", got)
	}
	invalid := "https://user:password-canary@example.com:bad"
	if URL(invalid) != "[invalid URL]" {
		t.Fatal("invalid URL exposed")
	}
	nested := &url.Error{
		Op: "Get", URL: raw,
		Err: &url.Error{
			Op: "proxy", URL: raw, Err: errors.New("connection refused"),
		},
	}
	if got := Error(nested).Error(); got != "connection refused" {
		t.Fatal("nested URL leaked")
	}
}
