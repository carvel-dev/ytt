// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"carvel.dev/ytt/pkg/experiments"
	"carvel.dev/ytt/pkg/version"
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
	for _, experiment := range experiments.GetEnabled() {
		fmt.Printf("- experiment %q enabled.\n", experiment)
	}
	return nil
}
