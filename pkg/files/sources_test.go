// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestHTTPFileSources(t *testing.T) {
	url := "http://example.com/some/path"

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		require.Equal(t, req.URL.String(), url)
		return &http.Response{
			StatusCode: http.StatusOK,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	fileSource := files.NewHTTPSource(url)
	fileSource.Client = client
	body, err := fileSource.Bytes()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), body)

	// 2xx Status Codes
	client = NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		require.Equal(t, req.URL.String(), url)
		return &http.Response{
			StatusCode: http.StatusIMUsed,
			Body:       io.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	})

	fileSource = files.NewHTTPSource(url)
	fileSource.Client = client
	body, err = fileSource.Bytes()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), body)

	// Non-OK HTTP Status Code
	status := "404 Not Found"
	client = NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		require.Equal(t, req.URL.String(), url)
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     status,
			Header:     make(http.Header),
		}
	})

	fileSource = files.NewHTTPSource(url)
	fileSource.Client = client
	_, err = fileSource.Bytes()
	require.EqualError(t, err, fmt.Sprintf("Requesting URL '%s': %s", url, status))
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
