package template

import (
	"encoding/json"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/spf13/cobra"
)

type BulkFilesSourceOpts struct {
	bulkIn  string
	bulkOut bool
}

func (s *BulkFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.bulkIn, "bulk-in", "", "Accept files in bulk format")
	cmd.Flags().BoolVar(&s.bulkOut, "bulk-out", false, "Output files in bulk format")
}

type BulkFilesSource struct {
	opts BulkFilesSourceOpts
	ui   cmdcore.PlainUI
}

type BulkFiles struct {
	Files  []BulkFile `json:"files,omitempty"`
	Errors string     `json:"errors,omitempty"`
}

type BulkFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewBulkFilesSource(opts BulkFilesSourceOpts, ui cmdcore.PlainUI) *BulkFilesSource {
	return &BulkFilesSource{opts, ui}
}

func (s *BulkFilesSource) HasInput() bool  { return len(s.opts.bulkIn) > 0 }
func (s *BulkFilesSource) HasOutput() bool { return s.opts.bulkOut }

func (s BulkFilesSource) Input() (TemplateInput, error) {
	var fs BulkFiles
	err := json.Unmarshal([]byte(s.opts.bulkIn), &fs)
	if err != nil {
		return TemplateInput{}, err
	}

	var result []*files.File

	for _, f := range fs.Files {
		file, err := files.NewFileFromSource(files.NewBytesSource(f.Name, []byte(f.Data)))
		if err != nil {
			return TemplateInput{}, err
		}

		result = append(result, file)
	}

	return TemplateInput{files.NewSortedFiles(result)}, nil
}

func (s *BulkFilesSource) Output(out TemplateOutput) error {
	fs := BulkFiles{}

	if out.Err != nil {
		fs.Errors = out.Err.Error()
	}

	for _, outputFile := range out.Files {
		fs.Files = append(fs.Files, BulkFile{
			Name: outputFile.RelativePath(),
			Data: string(outputFile.Bytes()),
		})
	}

	resultBytes, err := json.Marshal(fs)
	if err != nil {
		return err
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", resultBytes)

	return nil
}
