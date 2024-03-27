// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/version"
	"github.com/cppforlife/cobrautil"
	"github.com/spf13/cobra"
)

type YttOptions struct{}

func NewDefaultYttOptions() *YttOptions {
	return &YttOptions{}
}

func NewDefaultYttCmd() *cobra.Command {
	return NewYttCmd(NewDefaultYttOptions())
}

func NewYttCmd(o *YttOptions) *cobra.Command {
	cmd := NewCmd(cmdtpl.NewOptions())

	cmd.Use = "ytt"
	cmd.Aliases = nil
	cmd.Version = version.Version
	cmd.Short = "ytt performs YAML templating"
	cmd.Long = `ytt performs YAML templating.

Docs: https://carvel.dev/ytt/docs/latest/
Docs for data values: https://carvel.dev/ytt/docs/latest/ytt-data-values/`

	// Affects children as well
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Disable docs header
	cmd.DisableAutoGenTag = true

	// TODO bash completion

	cmd.AddCommand(NewVersionCmd(NewVersionOptions()))
	cmd.AddCommand(NewCmd(cmdtpl.NewOptions())) // for backwards compat
	cmd.AddCommand(NewFmtCmd(NewFmtOptions()))
	cmd.AddCommand(NewWebsiteCmd(NewWebsiteOptions()))

	// Reconfigure Commands
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureCmdWithSubcmd,
		cobrautil.DisallowExtraArgs, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}
