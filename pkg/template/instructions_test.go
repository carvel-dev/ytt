// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"sync"
	"testing"

	"carvel.dev/ytt/pkg/template"
	"github.com/stretchr/testify/require"
)

// TestNewInstructionSetConcurrency tests that all identifiers NewInstructionSet
// generates are unique when used concurrently.
func TestNewInstructionSetConcurrency(t *testing.T) {
	const (
		goroutines     = 16
		setsPerRoutine = 64
		wantIDs        = goroutines * setsPerRoutine
	)

	// Each goroutine owns one slot, so collecting results needs no locking
	// and cannot itself be a source of races.
	generated := make([][]string, goroutines)

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Go(func() {
			names := make([]string, setsPerRoutine)
			for j := range names {
				// Every identifier in a set derives from that set's ID, so one
				// field is enough to spot a collision.
				names[j] = template.NewInstructionSet().SetNode.Name
			}
			generated[i] = names
		})
	}
	wg.Wait()

	// Comparing counts rather than using require.Len keeps the failure
	// message from dumping every generated identifier.
	require.Equal(t, wantIDs, countUnique(generated),
		"duplicate identifiers mean the ID counter lost updates")
}

func countUnique(groups [][]string) int {
	unique := map[string]struct{}{}
	for _, group := range groups {
		for _, name := range group {
			unique[name] = struct{}{}
		}
	}
	return len(unique)
}
