// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/carvel-dev/ytt/pkg/cmd/template"
	"github.com/spf13/cobra"

	// Load ytt library extensions (should be available in ytt binary)
	_ "github.com/carvel-dev/ytt/pkg/yttlibraryext"
)

// NewCmd construct main ytt command. It has been moved out of "template" package
// so that "template" package does not carry dependency on cobra. This is desirable
// for users who use ytt as a library.
func NewCmd(o *template.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Process YAML templates (deprecated; use top-level command -- e.g. `ytt -f-` instead of `ytt template -f-`)",
		RunE:    func(c *cobra.Command, args []string) error { return o.Run() },
	}
	o.BindFlags(cmd.Flags())
	return cmd
}
