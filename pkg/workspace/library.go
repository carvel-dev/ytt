// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/k14s/ytt/pkg/files"
)

const (
	privateName = "_ytt_lib"
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
			lib, found := currLibrary.findLibrary(piece)
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

func (l *Library) findLibrary(name string) (*Library, bool) {
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
		lib, found := currLibrary.findLibrary(piece)
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

func (l *Library) FindLibrary(path string) (*Library, error) {
	dirPieces, namePiece := files.SplitPath(path)

	var currLibrary *Library = l
	for i, piece := range append(dirPieces, namePiece) {
		lib, found := currLibrary.findLibrary(piece)
		if !found {
			return nil, fmt.Errorf("Did not find '%s'", files.JoinPath(dirPieces[:i]))
		}
		if lib.private {
			return nil, fmt.Errorf("Encountered private library '%s'", privateName)
		}
		currLibrary = lib
	}

	return currLibrary, nil
}

func (l *Library) FindFile(path string) (FileInLibrary, error) {
	dirPieces, namePiece := files.SplitPath(path)

	var currLibrary *Library = l

	if len(dirPieces) > 0 {
		lib, err := l.FindLibrary(files.JoinPath(dirPieces))
		if err != nil {
			return FileInLibrary{}, fmt.Errorf("Expected to find file '%s', but did not: %s", path, err)
		}
		currLibrary = lib
	}

	for _, file := range currLibrary.files {
		_, fileNamePiece := files.SplitPath(file.RelativePath())
		if fileNamePiece == namePiece {
			return FileInLibrary{File: file, Library: currLibrary}, nil
		}
	}

	return FileInLibrary{}, fmt.Errorf(
		"Expected to find file '%s' (hint: only files included via -f flag are available)", path)
}

type FileInLibrary struct {
	File            *files.File
	Library         *Library
	parentLibraries []*Library
}

func (fileInLib *FileInLibrary) RelativePath() string {
	var components []string
	for _, lib := range fileInLib.parentLibraries {
		components = append(components, lib.name)
	}
	_, fileName := files.SplitPath(fileInLib.File.RelativePath())
	components = append(components, fileName)
	return files.JoinPath(components)
}

func (l *Library) ListAccessibleFiles() []*FileInLibrary {
	return l.listAccessibleFiles(nil)
}

func (l *Library) listAccessibleFiles(parents []*Library) []*FileInLibrary {
	var result []*FileInLibrary
	for _, file := range l.files {
		result = append(result, &FileInLibrary{
			File:            file,
			Library:         l,
			parentLibraries: parents,
		})
	}
	for _, lib := range l.children {
		if !lib.private {
			newParents := append(append([]*Library{}, parents...), lib)
			result = append(result, lib.listAccessibleFiles(newParents)...)
		}
	}
	return result
}

func (l *Library) Print(out io.Writer) {
	l.print(out, 0)
}

func (l *Library) PrintStr() string {
	var buf bytes.Buffer
	l.print(&buf, 0)
	return buf.String()
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

func SortFilesInLibrary(files []*FileInLibrary) {
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].File.OrderLess(files[j].File)
	})
}
