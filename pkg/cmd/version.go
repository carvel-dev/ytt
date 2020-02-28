package cmd

import (
	"fmt"

	"github.com/k14s/ytt/pkg/version"
	"github.com/spf13/cobra"
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
	fmt.Printf("ytt version %s\n", version.Version)

	return nil
}
