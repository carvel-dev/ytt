package cobrautil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func VisitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		VisitCommands(child, f)
	}
}

func WrapRunEForCmd(additionalRunE func(*cobra.Command, []string) error) func(cmd *cobra.Command) {
	return func(cmd *cobra.Command) {
		origRunE := cmd.RunE
		cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
			err := additionalRunE(cmd2, args)
			if err != nil {
				return err
			}
			return origRunE(cmd2, args)
		}
	}
}

func ReconfigureCmdWithSubcmd(cmd *cobra.Command) {
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
		if !subcmd.Hidden {
			strs = append(strs, subcmd.Use)
		}
	}

	cmd.Short += " (" + strings.Join(strs, ", ") + ")"
}

func ReconfigureLeafCmd(cmd *cobra.Command) {
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
		if !subcmd.Hidden {
			strs = append(strs, subcmd.Use)
		}
	}
	return fmt.Errorf("Use one of available subcommands: %s", strings.Join(strs, ", "))
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return fmt.Errorf("Invalid command - see available commands/subcommands above")
}
