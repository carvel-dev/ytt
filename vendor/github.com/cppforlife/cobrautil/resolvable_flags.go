package cobrautil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ResolvableFlag interface {
	Resolve() error
}

func ResolveFlagsForCmd(cmd *cobra.Command, args []string) error {
	var lastFlagErr error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value == nil {
			return
		}
		if resolvableVal, ok := flag.Value.(ResolvableFlag); ok {
			err := resolvableVal.Resolve()
			if err != nil {
				lastFlagErr = err
			}
		}
	})
	return lastFlagErr
}
