// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestProxyEventToHTTPRequestDoesNotLeakQueryToken(t *testing.T) {
	const token = "unit-test-secret"

	tests := []struct {
		name   string
		host   string
		path   string
		reason string
	}{
		{"invalid path escape", DefaultServerAddress, "/%zz",
			"invalid URL escape"},
		{"invalid configured host port", "https://example.com:badport", "/",
			"invalid port"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTokenNotLeaked(t, tt.host, tt.path, token, tt.reason)
		})
	}
}

func checkTokenNotLeaked(t *testing.T, host, path, token, reason string) {
	// Do not run in parallel: this changes the default logger.
	t.Setenv(CustomHostVariable, host)

	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousOutput) })

	event := events.ALBTargetGroupRequest{
		HTTPMethod: http.MethodGet,
		Path:       path,
		MultiValueQueryStringParameters: map[string][]string{
			"token": {token},
		},
	}

	_, err := (&RequestAccessor{}).ProxyEventToHTTPRequest(event)
	if err == nil {
		t.Fatal("expected request conversion to fail")
	}
	if strings.Contains(logs.String(), token) {
		t.Errorf("query token leaked in logs: %s", logs.String())
	}
	if strings.Contains(err.Error(), token) {
		t.Errorf("query token leaked in returned error: %v", err)
	}
	if !strings.Contains(err.Error(), reason) {
		t.Errorf("returned error lost failure reason %q: %v", reason, err)
	}
}
