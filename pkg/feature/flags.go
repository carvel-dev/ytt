// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

/*
Package feature provides a global "Feature Flag" facility used to toggle entire features on or off.

To "register" a new feature,
1. add a new constant
2. add that constant to the `allFeatures` slice

Initialize the flags, enabling desired features:

    feature.Flags().Enable(<feature-constant>)

To then circuit-break functionality behind a feature flag.

    if feature.Flags().IsEnabled(<feature-constant>) {
        ...
    }
*/
package feature

import "fmt"

// Names of features that can be toggled
const (
	Noop        = "noop"
	Validations = "validations"
)

// allFeatures is the total list of features. It must contain all the constants defined, above.
var allFeatures = []string{Noop, Validations}

// Flags returns the singleton instance of feature flags.
func Flags() *Flagset {
	// NOT thread-safe.
	if instance == nil {
		instance = newFlagSet()
	}
	return instance
}

// Flagset is a collection of flags.
type Flagset struct {
	flags map[string]bool
}

// Enable toggles the named feature on.
// Subsequent calls to IsEnabled() for that same named feature will return true.
//
// Once this Flagset has been frozen, panics.
func (f *Flagset) Enable(name string) *Flagset {
	f.ensureExists(name)
	f.flags[name] = true
	return f
}

// IsEnabled reports whether the named feature has been enabled.
//
// If the feature has been explicitly enabled, returns true.
// Otherwise, returns false.
func (f *Flagset) IsEnabled(name string) bool {
	f.ensureExists(name)
	return f.flags[name]
}

func (f *Flagset) ensureExists(name string) {
	_, ok := f.flags[name]
	if !ok {
		panic(fmt.Sprintf("Unknown feature %q; known features are %q", name, allFeatures))
	}
}

func newFlagSet() *Flagset {
	flagset := &Flagset{flags: make(map[string]bool, len(allFeatures))}
	for _, feature := range allFeatures {
		flagset.flags[feature] = false
	}
	return flagset
}

var instance = newFlagSet()
