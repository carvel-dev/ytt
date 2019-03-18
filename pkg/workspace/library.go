package workspace

import (
	"fmt"
	"io"
	"strings"

	"github.com/k14s/ytt/pkg/files"
)

const (
	privateName   = "_ytt_lib"
	pathSeparator = "/"
)

type Library struct {
	name     string
	private  bool // in _ytt_lib
	children []*Library
	files    []*files.File
}

func NewRootLibrary(fs []*files.File) *Library {
	rootLibrary := &Library{}

	for _, file := range fs {
		dirPieces, _ := files.SplitPath(file.RelativePath())

		var currLibrary *Library = rootLibrary
		for _, piece := range dirPieces {
			lib, found := currLibrary.FindLibrary(piece)
			if !found {
				currLibrary = currLibrary.CreateLibrary(piece)
			} else {
				currLibrary = lib
			}
		}

		currLibrary.files = append(currLibrary.files, file)
	}

	return rootLibrary
}

func (l *Library) FindLibrary(name string) (*Library, bool) {
	for _, lib := range l.children {
		if lib.name == name {
			return lib, true
		}
	}
	return nil, false
}

func (l *Library) CreateLibrary(name string) *Library {
	lib := &Library{name: name, private: name == privateName}
	l.children = append(l.children, lib)
	return lib
}

func (l *Library) FindAccessibleLibrary(path string) (*Library, error) {
	dirPieces, namePiece := files.SplitPath(path)
	pieces := append(dirPieces, namePiece)

	privateLib, found := l.findPrivateLibrary()
	if !found {
		return nil, fmt.Errorf("Could not find private library (directory '%s' missing?)", privateName)
	}

	var currLibrary *Library = privateLib
	for i, piece := range pieces {
		lib, found := currLibrary.FindLibrary(piece)
		if !found {
			return nil, fmt.Errorf("Expected to find library '%s', but did not find '%s'",
				path, files.JoinPath(pieces[:i]))
		}
		if lib.private {
			return nil, fmt.Errorf("Could not load private library '%s'",
				files.JoinPath(pieces[:i]))
		}
		currLibrary = lib
	}

	return currLibrary, nil
}

func (l *Library) findPrivateLibrary() (*Library, bool) {
	for _, lib := range l.children {
		if lib.private {
			return lib, true
		}
	}
	return nil, false
}

func (l *Library) FindFile(path string) (*files.File, error) {
	dirPieces, namePiece := files.SplitPath(path)

	var currLibrary *Library = l
	for i, piece := range dirPieces {
		lib, found := currLibrary.FindLibrary(piece)
		if !found {
			return nil, fmt.Errorf("Expected to find file '%s', but did not find '%s'",
				path, files.JoinPath(dirPieces[:i]))
		}
		if lib.private {
			return nil, fmt.Errorf("Could not load file '%s' because it's contained in private library '%s' "+
				"(use load(\"@lib:file\", \"symbol\") where 'lib' is library name under %s, for example, 'github.com/k14s/test')",
				path, files.JoinPath(dirPieces[:i]), privateName)
		}
		currLibrary = lib
	}

	for _, file := range currLibrary.files {
		_, fileNamePiece := files.SplitPath(file.RelativePath())
		if fileNamePiece == namePiece {
			return file, nil
		}
	}
	return nil, fmt.Errorf("Expected to find file %s", path)
}

func (l *Library) ListAccessibleFiles() []*files.File {
	var result []*files.File
	result = append(result, l.files...)
	for _, lib := range l.children {
		if !lib.private {
			result = append(result, lib.ListAccessibleFiles()...)
		}
	}
	return result
}

func (l *Library) Print(out io.Writer) {
	l.print(out, 0)
}

func (l *Library) print(out io.Writer, indent int) {
	indentStr := strings.Repeat("  ", indent)
	fmt.Fprintf(out, "%s- %s (private %t)\n", indentStr, l.name, l.private)

	fmt.Fprintf(out, "%s  files:\n", indentStr)
	if len(l.files) == 0 {
		fmt.Fprintf(out, "%s    <none>\n", indentStr)
	}
	for _, file := range l.files {
		fmt.Fprintf(out, "%s  - %s\n", indentStr, file.RelativePath())
	}

	fmt.Fprintf(out, "%s  libraries:\n", indentStr)
	if len(l.children) == 0 {
		fmt.Fprintf(out, "%s    <none>\n", indentStr)
	}
	for _, lib := range l.children {
		lib.print(out, indent+1)
	}
}
