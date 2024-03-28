// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package template provides the core templating engine for ytt.

At its heart, a template is text (possibly of a specific format like YAML) that
is parsed into a set of template.EvaluationNode's which is then compiled into a
Starlark program whose objective is to build an output structure that reflects
the evaluated expressions that were annotated in the original text.
*/
package template
