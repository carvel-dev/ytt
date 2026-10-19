// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// CustomHostVariable is the name of the environment variable that contains
// the custom hostname for the request. If this variable is not set the framework
// reverts to `DefaultServerAddress`. The value for a custom host should include
// a protocol: http://my-custom.host.com
const CustomHostVariable = "GO_API_HOST"

// DefaultServerAddress is prepended to the path of each incoming request
const DefaultServerAddress = "https://aws-serverless-go-api.com"

type RequestAccessor struct {
	stripBasePath string
}

func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.Path
	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	serverAddress := DefaultServerAddress
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}
	path = serverAddress + path

	if len(req.MultiValueQueryStringParameters) > 0 {
		queryString := ""
		for q, l := range req.MultiValueQueryStringParameters {
			for _, v := range l {
				if queryString != "" {
					queryString += "&"
				}
				queryString += url.QueryEscape(q) + "=" + url.QueryEscape(v)
			}
		}
		path += "?" + queryString
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.HTTPMethod),
		path,
		bytes.NewReader(decodedBody),
	)
	if err != nil {
		sanitizedErr := sanitizeRequestError(req.HTTPMethod, err)
		log.Println(sanitizedErr)
		return nil, sanitizedErr
	}

	for h := range req.Headers {
		httpRequest.Header.Add(h, req.Headers[h])
	}

	for hk, hvs := range req.MultiValueHeaders {
		for _, hv := range hvs {
			httpRequest.Header.Add(hk, hv)
		}
	}

	return httpRequest, nil
}

// sanitizeRequestError keeps the cause of a request construction failure but
// drops the URL, which http.NewRequest embeds in its error and which may
// contain secrets in the query string.
func sanitizeRequestError(method string, err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}
	return fmt.Errorf("could not convert %s request to http.Request: %w",
		method, err)
}
