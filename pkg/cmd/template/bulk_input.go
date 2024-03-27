// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"encoding/json"

	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
)

type BulkFilesSourceOpts struct {
	bulkIn  string
	bulkOut bool
}

// Set registers "bulk" flags and wires-up those flags up to this
// BulkFilesSourceOpts to be set when the corresponding cobra.Command is executed.
func (s *BulkFilesSourceOpts) Set(cmdFlags CmdFlags) {
	cmdFlags.StringVar(&s.bulkIn, "bulk-in", "", "Accept files in bulk format")
	cmdFlags.BoolVar(&s.bulkOut, "bulk-out", false, "Output files in bulk format")
}

type BulkFilesSource struct {
	opts BulkFilesSourceOpts
	ui   ui.UI
}

type BulkFiles struct {
	Files  []BulkFile `json:"files,omitempty"`
	Errors string     `json:"errors,omitempty"`
}

type BulkFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewBulkFilesSource(opts BulkFilesSourceOpts, ui ui.UI) *BulkFilesSource {
	return &BulkFilesSource{opts, ui}
}

func (s *BulkFilesSource) HasInput() bool  { return len(s.opts.bulkIn) > 0 }
func (s *BulkFilesSource) HasOutput() bool { return s.opts.bulkOut }

func (s BulkFilesSource) Input() (Input, error) {
	var fs BulkFiles
	err := json.Unmarshal([]byte(s.opts.bulkIn), &fs)
	if err != nil {
		return Input{}, err
	}

	var result []*files.File

	for _, f := range fs.Files {
		file, err := files.NewFileFromSource(files.NewBytesSource(f.Name, []byte(f.Data)))
		if err != nil {
			return Input{}, err
		}

		result = append(result, file)
	}

	return Input{files.NewSortedFiles(result)}, nil
}

func (s *BulkFilesSource) Output(out Output) error {
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
