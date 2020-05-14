package cobrautil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type ReconfigureFunc func(cmd *cobra.Command)

func VisitCommands(cmd *cobra.Command, fns ...ReconfigureFunc) {
	for _, f := range fns {
		f(cmd)
	}
	for _, child := range cmd.Commands() {
		VisitCommands(child, fns...)
	}
}

func ReconfigureLeafCmds(fs ...func(cmd *cobra.Command)) ReconfigureFunc {
	return func(cmd *cobra.Command) {
		if len(cmd.Commands()) > 0 {
			return
		}

		for _, f := range fs {
			f(cmd)
		}
	}
}

func WrapRunEForCmd(additionalRunE func(*cobra.Command, []string) error) ReconfigureFunc {
	return func(cmd *cobra.Command) {
		if cmd.RunE == nil {
			panic(fmt.Sprintf("Internal: Command '%s' does not set RunE", cmd.CommandPath()))
		}

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

// ReconfigureFuncs

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

func DisallowExtraArgs(cmd *cobra.Command) {
	WrapRunEForCmd(func(cmd2 *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("command '%s' does not accept extra arguments '%s'", cmd2.CommandPath(), args[0])
		}
		return nil
	})(cmd)
	cmd.Args = cobra.ArbitraryArgs
}

// New RunE's

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
