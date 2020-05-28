package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	suspiciousOutputDirectoryPaths = []string{"/", ".", "./", ""}
)

type OutputDirectory struct {
	path  string
	files []OutputFile
	ui    UI
}

func NewOutputDirectory(path string, files []OutputFile, ui UI) *OutputDirectory {
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

// clean removes all files that may conflict with output files
// we don's just use os.RemoveAll to avoid accidently deleting
// files like .git if incorrect directory is specified.
func (d *OutputDirectory) removeOldFiles() error {
	fileInfo, err := os.Stat(d.path)
	if err != nil {
		return fmt.Errorf("Checking directory '%s'", d.path)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("Expected file '%s' to be a directory", d.path)
	}

	var selectedPaths []string

	err = filepath.Walk(d.path, func(walkedPath string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}

		// TODO does not work with filtering of template files
		if (&File{nil, walkedPath, nil, nil, nil, nil, 0}).IsForOutput() {
			selectedPaths = append(selectedPaths, walkedPath)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Listing files '%s'", d.path)
	}

	for _, selectedPath := range selectedPaths {
		d.ui.Printf("deleting: %s\n", selectedPath)

		err := os.Remove(selectedPath)
		if err != nil {
			return fmt.Errorf("Deleting file '%s'", selectedPath)
		}
	}

	return nil
}
