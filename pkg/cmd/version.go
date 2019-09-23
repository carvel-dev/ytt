package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version = "0.21.0"
)

type VersionOptions struct{}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func NewVersionCmd(o *VersionOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	return cmd
}

func (o *VersionOptions) Run() error {
	fmt.Printf("Version: %s\n", Version)

	return nil
}
