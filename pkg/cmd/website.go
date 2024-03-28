// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/website"
	"github.com/spf13/cobra"
)

type WebsiteOptions struct {
	ListenAddr      string
	RedirectToHTTPS bool
	BinaryPath      string
	CheckCookie     bool
}

func NewWebsiteOptions() *WebsiteOptions {
	return &WebsiteOptions{
		BinaryPath: os.Args[0],
	}
}

func NewWebsiteCmd(o *WebsiteOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "website",
		Short: "Starts website HTTP server",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().StringVar(&o.ListenAddr, "listen-addr", "localhost:8080", "Listen address")
	cmd.Flags().BoolVar(&o.RedirectToHTTPS, "redirect-to-https", true, "Redirect to HTTPs address")
	return cmd
}

func (o *WebsiteOptions) Server() *website.Server {
	opts := website.ServerOpts{
		ListenAddr:      o.ListenAddr,
		RedirectToHTTPS: o.RedirectToHTTPS,
		TemplateFunc:    o.execBinary,
		ErrorFunc:       o.bulkOutErr,
		CheckCookie:     o.CheckCookie,
	}
	return website.NewServer(opts)
}

func (o *WebsiteOptions) Run() error {
	return o.Server().Run()
}

func (o *WebsiteOptions) execBinary(data []byte) ([]byte, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command(o.BinaryPath, "--bulk-in", string(data), "--bulk-out")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error: %s", stderr.String())
	}

	return out.Bytes(), nil
}

func (*WebsiteOptions) bulkOutErr(err error) ([]byte, error) {
	return json.Marshal(cmdtpl.BulkFiles{Errors: err.Error()})
}
