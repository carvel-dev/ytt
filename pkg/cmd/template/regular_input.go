// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"io"
	"strings"

	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/spf13/cobra"
)

// RegularFilesSourceOpts holds configuration for when regular files are the input/output
// of the execution (via the --files and --output... flags).
type RegularFilesSourceOpts struct {
	files []string

	outputDir   string
	OutputFiles string
	OutputType  OutputType

	files.SymlinkAllowOpts
}

// OutputType holds the user's desire for two (2) categories of output:
// - file format type :: yaml, json, pos
// - schema type :: OpenAPI V3, ytt Schema
type OutputType struct {
	Types []string
}

// Set registers flags related to sourcing ordinary files/directories and wires-up those flags up to this
// RegularFilesSourceOpts to be set when the corresponding cobra.Command is executed.
func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")

	cmd.Flags().StringVar(&s.outputDir, "dangerous-emptied-output-directory", "",
		"Delete given directory, and then create it with output files")
	cmd.Flags().StringVar(&s.OutputFiles, "output-files", "", "Add output files to given directory")

	cmd.Flags().StringSliceVarP(&s.OutputType.Types, "output", "o", []string{RegularFilesOutputTypeYAML},
		fmt.Sprintf("Configure output format. Can specify file format (%s) and/or schema type (%s) (can be specified multiple times)",
			strings.Join(RegularFilesOutputFormatTypes, ", "),
			strings.Join(RegularFilesOutputSchemaTypes, ", ")))

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

	outputType, err := s.opts.OutputType.Format()
	if err != nil {
		return err
	}

	switch outputType {
	case RegularFilesOutputTypeYAML:
		printerFunc = nil
	case RegularFilesOutputTypeJSON:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter { return yamlmeta.NewJSONPrinter(w) }
	case RegularFilesOutputTypePos:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter {
			return yamlmeta.WrappedFilePositionPrinter{yamlmeta.NewFilePositionPrinter(w)}
		}
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

// When the FileSource are RegularFilesSource, indicates which file format to use when rendering the output.
const (
	RegularFilesOutputTypeYAML = "yaml"
	RegularFilesOutputTypeJSON = "json"
	RegularFilesOutputTypePos  = "pos"
)

// When the FileSource are RegularFilesSource, indicates which schema type to use when rendering the output.
const (
	RegularFilesOutputTypeOpenAPI = "openapi-v3"
	RegularFilesOutputTypeNone    = ""
)

// Collections of each category of output type
var (
	RegularFilesOutputFormatTypes = []string{RegularFilesOutputTypeYAML, RegularFilesOutputTypeJSON, RegularFilesOutputTypePos}
	RegularFilesOutputSchemaTypes = []string{RegularFilesOutputTypeOpenAPI}
	RegularFilesOutputTypes       = append(RegularFilesOutputFormatTypes, RegularFilesOutputSchemaTypes...)
)

// Format returns which of the file format types is in effect
func (o *OutputType) Format() (string, error) {
	return o.typeFrom(RegularFilesOutputFormatTypes, RegularFilesOutputTypeYAML)
}

// Schema returns which of the schema types is in effect
func (o *OutputType) Schema() (string, error) {
	return o.typeFrom(RegularFilesOutputSchemaTypes, RegularFilesOutputTypeNone)
}

func (o *OutputType) typeFrom(types []string, defaultValue string) (string, error) {
	if err := o.validateTypes(); err != nil {
		return "", err
	}
	for _, t := range o.Types {
		if o.stringInSlice(t, types) {
			return t, nil
		}
	}
	return defaultValue, nil
}

func (o *OutputType) validateTypes() error {
	for _, t := range o.Types {
		if !o.stringInSlice(t, RegularFilesOutputTypes) {
			return fmt.Errorf("Unknown output type '%v'", t)
		}
	}
	return nil
}

func (o *OutputType) stringInSlice(target string, slice []string) bool {
	for _, s := range slice {
		if target == s {
			return true
		}
	}
	return false
}
