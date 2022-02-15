// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/carvel-ytt/pkg/version"
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
