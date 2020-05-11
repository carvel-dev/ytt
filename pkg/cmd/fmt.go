package cmd

import (
	"os"
	"time"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlfmt"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

type FmtOptions struct {
	Files      []string
	StrictYAML bool
	Debug      bool
}

func NewFmtOptions() *FmtOptions {
	return &FmtOptions{}
}

func NewFmtCmd(o *FmtOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "Format YAML templates",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().StringArrayVarP(&o.Files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")
	cmd.Flags().BoolVarP(&o.StrictYAML, "strict", "s", false, "Configure to use _strict_ YAML subset")
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Enable debug output")
	return cmd
}

func (o *FmtOptions) Run() error {
	ui := cmdcore.NewPlainUI(o.Debug)
	t1 := time.Now()

	defer func() {
		ui.Debugf("total: %s\n", time.Now().Sub(t1))
	}()

	filesToProcess, err := files.NewSortedFilesFromPaths(o.Files, files.SymlinkAllowOpts{})
	if err != nil {
		return err
	}

	for _, file := range filesToProcess {
		if file.Type() == files.TypeYAML {
			data, err := file.Bytes()
			if err != nil {
				return err
			}

			docSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{Strict: o.StrictYAML}).ParseBytes([]byte(data), "")
			if err != nil {
				return err
			}

			yamlfmt.NewPrinter(os.Stdout).Print(docSet)
		}
	}

	return nil
}
