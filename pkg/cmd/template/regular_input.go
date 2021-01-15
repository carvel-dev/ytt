// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"io"

	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

const (
	regularFilesOutputTypeYAML = "yaml"
	regularFilesOutputTypeJSON = "json"
	regularFilesOutputTypePos  = "pos"
)

type RegularFilesSourceOpts struct {
	files []string

	outputDir   string
	OutputFiles string
	OutputType  string

	files.SymlinkAllowOpts
}

func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")

	cmd.Flags().StringVar(&s.outputDir, "dangerous-emptied-output-directory", "",
		"Delete given directory, and then create it with output files")
	cmd.Flags().StringVar(&s.OutputFiles, "output-files", "", "Add output files to given directory")

	cmd.Flags().StringVarP(&s.OutputType, "output", "o", regularFilesOutputTypeYAML, "Output type (yaml, json, pos)")

	cmd.Flags().BoolVar(&s.SymlinkAllowOpts.AllowAll, "dangerous-allow-all-symlink-destinations", false,
		"Symlinks to all destinations are allowed")
	cmd.Flags().StringSliceVar(&s.SymlinkAllowOpts.AllowedDstPaths, "allow-symlink-destination", nil,
		"File paths to which symlinks are allowed (can be specified multiple times)")
}

type RegularFilesSource struct {
	opts RegularFilesSourceOpts
	ui   ui.UI
}

func NewRegularFilesSource(opts RegularFilesSourceOpts, ui ui.UI) *RegularFilesSource {
	return &RegularFilesSource{opts, ui}
}

func (s *RegularFilesSource) HasInput() bool  { return len(s.opts.files) > 0 }
func (s *RegularFilesSource) HasOutput() bool { return true }

func (s *RegularFilesSource) Input() (Input, error) {
	filesToProcess, err := files.NewSortedFilesFromPaths(s.opts.files, s.opts.SymlinkAllowOpts)
	if err != nil {
		return Input{}, err
	}

	return Input{Files: filesToProcess}, nil
}

func (s *RegularFilesSource) Output(out Output) error {
	if out.Err != nil {
		return out.Err
	}

	nonYamlFileNames := []string{}
	switch {
	case len(s.opts.outputDir) > 0:
		return files.NewOutputDirectory(s.opts.outputDir, out.Files, s.ui).Write()
	case len(s.opts.OutputFiles) > 0:
		return files.NewOutputDirectory(s.opts.OutputFiles, out.Files, s.ui).WriteFiles()
	default:
		for _, file := range out.Files {
			if file.Type() != files.TypeYAML {
				nonYamlFileNames = append(nonYamlFileNames, file.RelativePath())
			}
		}
	}

	var printerFunc func(io.Writer) yamlmeta.DocumentPrinter

	switch s.opts.OutputType {
	case regularFilesOutputTypeYAML:
		printerFunc = nil
	case regularFilesOutputTypeJSON:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter { return yamlmeta.NewJSONPrinter(w) }
	case regularFilesOutputTypePos:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter {
			return yamlmeta.WrappedFilePositionPrinter{yamlmeta.NewFilePositionPrinter(w)}
		}
	default:
		return fmt.Errorf("Unknown output type '%s'", s.opts.OutputType)
	}

	combinedDocBytes, err := out.DocSet.AsBytesWithPrinter(printerFunc)
	if err != nil {
		return fmt.Errorf("Marshaling combined template result: %s", err)
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", combinedDocBytes) // no newline

	if len(nonYamlFileNames) > 0 {
		s.ui.Warnf("\n" + `Warning: Found Non-YAML templates in input. Non-YAML templates are not rendered to standard output.
If you want to include those results, use the --output-files or --dangerous-emptied-output-directory flag.` + "\n")
	}
	return nil
}
