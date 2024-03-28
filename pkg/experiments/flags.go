// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package experiments

import (
	"os"
	"strings"
)

/*
Registering a New Experiment

1. implement a getter on this package `Is<experiment-name>Enabled()` and add the experiment to GetEnabled()

2. circuit-break functionality behind that check:

    if experiments.Is<experiment-name>Enabled() {
        ...
    }

3. in tests, enable experiment(s) by setting the environment variable:

    experiments.ResetForTesting()
    os.Setenv(experiments.Env, "<experiment-name>,<other-experiment-name>,...")
*/

// Env is the OS environment variable with comma-separated names of experiments to enable.
const Env = "YTTEXPERIMENTS"

// GetEnabled reports the name of all enabled experiments.
//
// An experiment is enabled by including its name in the OS environment variable named Env.
func GetEnabled() []string {
	experiments := []string{}
	return experiments
}

func isSet(flag string) bool {
	for _, setting := range getSettings() {
		if setting == flag {
			return true
		}
	}
	return false
}

func getSettings() []string {
	if settings == nil {
		for _, setting := range strings.Split(os.Getenv(Env), ",") {
			settings = append(settings, strings.ToLower(strings.TrimSpace(setting)))
		}
	}
	return settings
}

// settings cached copy of name of experiments that are enabled (cleaned up).
var settings []string

// isNoopEnabled reports whether the "noop" experiment was enabled.
//
// This is for testing purposes only.
func isNoopEnabled() bool {
	return isSet("noop")
}

// ResetForTesting clears the experiment flag settings, forcing reload from the Env on next use.
//
// This is for testing purposes only.
func ResetForTesting() {
	settings = nil
}
