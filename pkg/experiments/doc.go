// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package experiments provides a global "Feature Flag" facility for
circuit-breaking pre-GA code.

The intent is to provide a means of selecting a flavor of executable at
boot; not to toggle experiments on and off. Once settings are loaded from
the environment variable experiments.Env, they are fixed.
*/
package experiments
