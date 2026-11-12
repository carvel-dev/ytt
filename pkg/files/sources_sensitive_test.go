// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPSourceDiagnosticsAndAuthentication(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "password-canary", pass)
		require.Equal(t, "query-canary", r.URL.Query().Get("token"))
		w.WriteHeader(http.StatusForbidden)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	raw := strings.Replace(server.URL,
		"http://", "http://user:password-canary@", 1) +
		"/input?token=query-canary"
	source := NewHTTPSource(raw)
	require.NotContains(t, source.Description(), "canary")
	_, _ = source.Bytes()
	server.Close()
	_, err := source.Bytes()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "canary")
}
