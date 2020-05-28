package template

import (
	"fmt"
	"io"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
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
	files     []string
	fileMarks []string

	outputDir   string
	outputFiles string
	outputType  string

	files.SymlinkAllowOpts
}

func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")

	cmd.Flags().StringVar(&s.outputDir, "dangerous-emptied-output-directory", "",
		"Delete given directory, and then create it with output files")
	cmd.Flags().StringVar(&s.outputFiles, "output-files", "", "Add output files to given directory")

	cmd.Flags().StringVarP(&s.outputType, "output", "o", regularFilesOutputTypeYAML, "Output type (yaml, json, pos)")

	cmd.Flags().BoolVar(&s.SymlinkAllowOpts.AllowAll, "dangerous-allow-all-symlink-destinations", false,
		"Symlinks to all destinations are allowed")
	cmd.Flags().StringSliceVar(&s.SymlinkAllowOpts.AllowedDstPaths, "allow-symlink-destination", nil,
		"File paths to which symlinks are allowed (can be specified multiple times)")
}

type RegularFilesSource struct {
	opts RegularFilesSourceOpts
	ui   cmdcore.PlainUI
}

func NewRegularFilesSource(opts RegularFilesSourceOpts, ui cmdcore.PlainUI) *RegularFilesSource {
	return &RegularFilesSource{opts, ui}
}

func (s *RegularFilesSource) HasInput() bool  { return len(s.opts.files) > 0 }
func (s *RegularFilesSource) HasOutput() bool { return true }

func (s *RegularFilesSource) Input() (TemplateInput, error) {
	filesToProcess, err := files.NewSortedFilesFromPaths(s.opts.files, s.opts.SymlinkAllowOpts)
	if err != nil {
		return TemplateInput{}, err
	}

	return TemplateInput{Files: filesToProcess}, nil
}

func (s *RegularFilesSource) Output(out TemplateOutput) error {
	if out.Err != nil {
		return out.Err
	}

	switch {
	case len(s.opts.outputDir) > 0:
		return files.NewOutputDirectory(s.opts.outputDir, out.Files, s.ui).Write()
	case len(s.opts.outputFiles) > 0:
		return files.NewOutputDirectory(s.opts.outputFiles, out.Files, s.ui).WriteFiles()
	}

	var printerFunc func(io.Writer) yamlmeta.DocumentPrinter

	switch s.opts.outputType {
	case regularFilesOutputTypeYAML:
		printerFunc = nil
	case regularFilesOutputTypeJSON:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter { return yamlmeta.NewJSONPrinter(w) }
	case regularFilesOutputTypePos:
		printerFunc = func(w io.Writer) yamlmeta.DocumentPrinter {
			return yamlmeta.WrappedFilePositionPrinter{yamlmeta.NewFilePositionPrinter(w)}
		}
	default:
		return fmt.Errorf("Unknown output type '%s'", s.opts.outputType)
	}

	combinedDocBytes, err := out.DocSet.AsBytesWithPrinter(printerFunc)
	if err != nil {
		return fmt.Errorf("Marshaling combined template result: %s", err)
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", combinedDocBytes) // no newline

	return nil
}
