package cmd

import (
	"fmt"
	"strings"

	"github.com/cppforlife/cobrautil"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
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
	cmd.Short = "ytt performs YAML templating"
	cmd.Long = `ytt performs YAML templating.

Docs: https://github.com/k14s/ytt/tree/master/docs
Docs for data values: https://github.com/k14s/ytt/blob/master/docs/ytt-data-values.md`

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

	// Last one runs first
	cobrautil.VisitCommands(cmd, reconfigureCmdWithSubcmd)
	cobrautil.VisitCommands(cmd, reconfigureLeafCmd)

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}

func reconfigureCmdWithSubcmd(cmd *cobra.Command) {
	if len(cmd.Commands()) == 0 {
		return
	}

	if cmd.Args == nil {
		cmd.Args = cobra.ArbitraryArgs
	}
	if cmd.RunE == nil {
		cmd.RunE = ShowSubcommands
	}

	var strs []string
	for _, subcmd := range cmd.Commands() {
		strs = append(strs, subcmd.Use)
	}

	cmd.Short += " (" + strings.Join(strs, ", ") + ")"
}

func reconfigureLeafCmd(cmd *cobra.Command) {
	if len(cmd.Commands()) > 0 {
		return
	}

	if cmd.RunE == nil {
		panic(fmt.Sprintf("Internal: Command '%s' does not set RunE", cmd.CommandPath()))
	}

	if cmd.Args == nil {
		origRunE := cmd.RunE
		cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("command '%s' does not accept extra arguments '%s'", args[0], cmd2.CommandPath())
			}
			return origRunE(cmd2, args)
		}
		cmd.Args = cobra.ArbitraryArgs
	}
}

func ShowSubcommands(cmd *cobra.Command, args []string) error {
	var strs []string
	for _, subcmd := range cmd.Commands() {
		strs = append(strs, subcmd.Use)
	}
	return fmt.Errorf("Use one of available subcommands: %s", strings.Join(strs, ", "))
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return fmt.Errorf("Invalid command - see available commands/subcommands above")
}
