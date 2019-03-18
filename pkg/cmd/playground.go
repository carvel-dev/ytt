package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/playground"
	"github.com/spf13/cobra"
)

type PlaygroundOptions struct {
	ListenAddr  string
	BinaryPath  string
	CheckCookie bool
}

func NewPlaygroundOptions() *PlaygroundOptions {
	return &PlaygroundOptions{
		BinaryPath: os.Args[0],
	}
}

func NewPlaygroundCmd(o *PlaygroundOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "playground",
		Aliases: []string{"pg"},
		Short:   "Starts playground HTTP server",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().StringVar(&o.ListenAddr, "listen-addr", "localhost:8080", "Listen address")
	return cmd
}

func (o *PlaygroundOptions) Server() *playground.Server {
	opts := playground.ServerOpts{
		ListenAddr:   o.ListenAddr,
		TemplateFunc: o.execBinary,
		ErrorFunc:    o.bulkOutErr,
		CheckCookie:  o.CheckCookie,
	}
	return playground.NewServer(opts)
}

func (o *PlaygroundOptions) Run() error {
	return o.Server().Run()
}

func (o *PlaygroundOptions) execBinary(data []byte) ([]byte, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command(o.BinaryPath, "tpl", "--bulk-in", string(data), "--bulk-out")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error: %s", stderr.String())
	}

	return out.Bytes(), nil
}

func (*PlaygroundOptions) bulkOutErr(err error) ([]byte, error) {
	return json.Marshal(cmdtpl.BulkFiles{Errors: err.Error()})
}
