package files

import (
	"os"
	"path/filepath"
)

type OutputFile struct {
	relativePath string
	data         []byte
}

func NewOutputFile(relativePath string, data []byte) OutputFile {
	return OutputFile{relativePath, data}
}

func (f OutputFile) RelativePath() string { return f.relativePath }
func (f OutputFile) Bytes() []byte        { return f.data }

func (f OutputFile) Path(dirPath string) string {
	return filepath.Join(dirPath, f.relativePath)
}

func (f OutputFile) Create(dirPath string) error {
	resultPath := f.Path(dirPath)

	err := os.MkdirAll(filepath.Dir(resultPath), 0700)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(resultPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}

	defer fd.Close()

	_, err = fd.Write(f.data)
	return err
}
