package template

import (
	"fmt"

	cmdcore "github.com/get-ytt/ytt/pkg/cmd/core"
	"github.com/get-ytt/ytt/pkg/files"
	"github.com/spf13/cobra"
)

type RegularFilesSourceOpts struct {
	files               []string
	filterTemplateFiles []string
	recursive           bool
	output              string
}

func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&s.filterTemplateFiles, "filter-template-file", nil, "Specify which file to template (can be specified multiple times)")
	cmd.Flags().BoolVarP(&s.recursive, "recursive", "R", false, "Interpret file as directory")
	cmd.Flags().StringVarP(&s.output, "output", "o", "", "Directory for output")
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
	filesToProcess, err := files.NewFiles(s.opts.files, s.opts.recursive)
	if err != nil {
		return TemplateInput{}, err
	}

	// Mark some files as non template files
	if len(s.opts.filterTemplateFiles) > 0 {
		for _, file := range filesToProcess {
			var isTemplate bool
			for _, filteredFile := range s.opts.filterTemplateFiles {
				if filteredFile == file.RelativePath() {
					isTemplate = true
					break
				}
			}
			if !isTemplate {
				file.MarkNonTemplate()
			}
		}
	}

	return TemplateInput{Files: filesToProcess}, nil
}

func (s *RegularFilesSource) Output(out TemplateOutput) error {
	if out.Err != nil {
		return out.Err
	}

	if len(s.opts.output) > 0 {
		return files.NewOutputDirectory(s.opts.output, out.Files, s.ui).Write()
	}

	combinedDocBytes, err := out.DocSet.AsBytes()
	if err != nil {
		return fmt.Errorf("Marshaling combined template result: %s", err)
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", combinedDocBytes) // no newline

	return nil
}
