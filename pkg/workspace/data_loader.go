// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/k14s/ytt/pkg/files"
)

type DataLoader struct {
	libraryCtx LibraryExecutionContext
}

func (l DataLoader) FilePaths(path string) ([]string, error) {
	library := l.libraryCtx.Current
	fromRoot := false

	if files.IsRootPath(path) {
		fromRoot = true
		library = l.libraryCtx.Root
		path = files.StripRootPath(path)
	}

	if len(path) > 0 {
		foundLibrary, err := library.FindLibrary(path)
		if err != nil {
			return nil, err
		}
		library = foundLibrary
	}

	result := []string{}
	for _, fileInLib := range library.ListAccessibleFiles() {
		path := fileInLib.RelativePath()
		// Make it compatible with FileData()
		if fromRoot {
			path = files.MakeRootPath(path)
		}
		result = append(result, path)
	}
	return result, nil
}

func (l DataLoader) FileData(path string) ([]byte, error) {
	library := l.libraryCtx.Current

	if files.IsRootPath(path) {
		library = l.libraryCtx.Root
		path = files.StripRootPath(path)
	}

	fileInLib, err := library.FindFile(path)
	if err != nil {
		return nil, err
	}

	fileBs, err := fileInLib.File.Bytes()
	if err != nil {
		return nil, err
	}

	return fileBs, nil
}
