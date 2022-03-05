// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/website"
)

type WebsiteOptions struct {
	ListenAddr      string
	RedirectToHTTPS bool
	CheckCookie     bool
}

func NewWebsiteOptions() *WebsiteOptions {
	return &WebsiteOptions{}
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
		TemplateFunc:    o.runYtt,
		ErrorFunc:       o.bulkOutErr,
		CheckCookie:     o.CheckCookie,
	}
	return website.NewServer(opts)
}

func (o *WebsiteOptions) Run() error {
	return o.Server().Run()
}

func (o *WebsiteOptions) runYtt(data []byte) ([]byte, error) {
	//var out, stderr bytes.Buffer

	cmdOptions := cmdtpl.NewBulkOptions(string(data), true)
	bytes, err := cmdOptions.RunLambda()

	//cmd := exec.Command(o.BinaryPath, "--bulk-in", string(data), "--bulk-out")
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr
	//
	//err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error: %s", err.Error())
	}

	return bytes, nil
}

func (*WebsiteOptions) bulkOutErr(err error) ([]byte, error) {
	return json.Marshal(cmdtpl.BulkFiles{Errors: err.Error()})
}
