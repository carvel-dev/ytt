// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package cmd is home to the full set of ytt's "commands" -- instances of cobra.Command
(not to be confused with ./cmd which contains the bootstrapping for executing ytt in various environments).

A cobra.Command is the starting point of execution.

For a list of commands run:

	$ ytt help

The default command is "template".
*/
package cmd
