// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package website

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Example struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Files       []File `json:"files,omitempty"`
}

type exampleSet struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Examples    []Example `json:"examples"`
}

// Files map is modified by ./generated.go created during ./hack/build.sh
var Files = map[string]File{}
var exampleSets = []exampleSet{}
