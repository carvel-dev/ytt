// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"os"
	"strings"

	"github.com/k14s/ytt/pkg/cmd/ui"
)

var (
	suspiciousOutputDirectoryPaths = []string{"/", ".", "./", ""}
)

type OutputDirectory struct {
	path  string
	files []OutputFile
	ui    ui.UI
}

func NewOutputDirectory(path string, files []OutputFile, ui ui.UI) *OutputDirectory {
	return &OutputDirectory{path, files, ui}
}

func (d *OutputDirectory) Files() []OutputFile { return d.files }

func (d *OutputDirectory) Write() error {
	filePaths := map[string]struct{}{}

	for _, file := range d.files {
		path := file.RelativePath()
		if _, found := filePaths[path]; found {
			return fmt.Errorf("Multiple files have same output destination paths: %s", path)
		}
		filePaths[path] = struct{}{}
	}

	for _, path := range suspiciousOutputDirectoryPaths {
		if d.path == path {
			return fmt.Errorf("Expected output directory path to not be one of '%s'",
				strings.Join(suspiciousOutputDirectoryPaths, "', '"))
		}
	}

	err := os.RemoveAll(d.path)
	if err != nil {
		return err
	}

	return d.WriteFiles()
}

func (d *OutputDirectory) WriteFiles() error {
	err := os.MkdirAll(d.path, 0700)
	if err != nil {
		return err
	}

	for _, file := range d.files {
		d.ui.Printf("creating: %s\n", file.Path(d.path))

		err := file.Create(d.path)
		if err != nil {
			return err
		}
	}

	return nil
}
