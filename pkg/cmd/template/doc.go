// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package template implements the "template" command (not to be confused with
"pkg/template" home of the templating mechanism itself).

Front-and-center is template.Options. This is both the host of ytt settings
parsed from the command-line through Cobra AND the top-level logic that
implements the command.
*/
package template
