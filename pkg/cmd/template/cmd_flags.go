// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

// CmdFlags interface decouples this package from
// depending on cobra.Command/flags concrete types.
type CmdFlags interface {
	BoolVar(p *bool, name string, value bool, usage string)
	BoolVarP(p *bool, name, shorthand string, value bool, usage string)

	StringVar(p *string, name string, value string, usage string)
	StringVarP(p *string, name, shorthand string, value string, usage string)

	StringArrayVar(p *[]string, name string, value []string, usage string)
	StringArrayVarP(p *[]string, name, shorthand string, value []string, usage string)

	StringSliceVar(p *[]string, name string, value []string, usage string)
	StringSliceVarP(p *[]string, name, shorthand string, value []string, usage string)
}
