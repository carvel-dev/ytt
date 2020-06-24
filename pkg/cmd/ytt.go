package cmd

import (
	"github.com/cppforlife/cobrautil"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/version"
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
	cmd := cmdtpl.NewCmd(cmdtpl.NewOptions())

	cmd.Use = "ytt"
	cmd.Aliases = nil
	cmd.Version = version.Version
	cmd.Short = "ytt performs YAML templating"
	cmd.Long = `ytt performs YAML templating.

Docs: https://github.com/k14s/ytt/tree/develop/docs
Docs for data values: https://github.com/k14s/ytt/blob/develop/docs/ytt-data-values.md`

	// Affects children as well
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Disable docs header
	cmd.DisableAutoGenTag = true

	// TODO bash completion

	cmd.AddCommand(NewVersionCmd(NewVersionOptions()))
	cmd.AddCommand(cmdtpl.NewCmd(cmdtpl.NewOptions())) // for backwards compat
	cmd.AddCommand(NewFmtCmd(NewFmtOptions()))
	cmd.AddCommand(NewWebsiteCmd(NewWebsiteOptions()))

	// Reconfigure Commands
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureCmdWithSubcmd,
		cobrautil.DisallowExtraArgs, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}
