// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	yamlExts     = []string{".yaml", ".yml"}
	starlarkExts = []string{".star"}
	textExts     = []string{".txt"}
	libraryExt   = "lib" // eg .lib.yaml
)

type Type int

const (
	TypeUnknown Type = iota
	TypeYAML
	TypeText
	TypeStarlark
)

type File struct {
	src     Source
	relPath string

	markedRelPath   *string
	markedType      *Type
	markedTemplate  *bool
	markedForOutput *bool

	order int // lowest comes first; 0 is used to indicate unsorted
}

func NewSortedFilesFromPaths(paths []string, opts SymlinkAllowOpts) ([]*File, error) {
	var groupedFiles [][]*File

	for _, path := range paths {
		var files []*File

		relativePath := ""
		pathPieces := strings.Split(path, "=")

		switch len(pathPieces) {
		case 1:
			// do nothing
		case 2:
			relativePath = pathPieces[0]
			path = pathPieces[1]
		default:
			return nil, fmt.Errorf("Expected file '%s' to only have single '=' sign to for relative path assignment", path)
		}

		switch {
		case path == "-":
			file, err := NewFileFromSource(NewCachedSource(NewStdinSource()))
			if err != nil {
				return nil, err
			}
			if len(relativePath) > 0 {
				file.MarkRelativePath(relativePath)
			}
			files = append(files, file)

		case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
			file, err := NewFileFromSource(NewCachedSource(NewHTTPSource(path)))
			if err != nil {
				return nil, err
			}
			if len(relativePath) > 0 {
				file.MarkRelativePath(relativePath)
			}
			files = append(files, file)

		default:
			fileInfo, err := os.Lstat(path)
			if err != nil {
				return nil, fmt.Errorf("Checking file '%s': %s", path, err)
			}

			if fileInfo.IsDir() {
				err := filepath.Walk(path, func(walkedPath string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() {
						return err
					}
					regLocalSource, err := NewRegularFileLocalSource(walkedPath, path, fi, opts)
					if err != nil {
						return err
					}
					file, err := NewFileFromSource(NewCachedSource(regLocalSource))
					if err != nil {
						return err
					}
					// TODO relative path for directories?
					files = append(files, file)
					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("Listing files '%s': %s", path, err)
				}
			} else {
				regLocalSource, err := NewRegularFileLocalSource(path, "", fileInfo, opts)
				if err != nil {
					return nil, err
				}
				file, err := NewFileFromSource(NewCachedSource(regLocalSource))
				if err != nil {
					return nil, err
				}
				if len(relativePath) > 0 {
					file.MarkRelativePath(relativePath)
				}
				files = append(files, file)
			}
		}

		groupedFiles = append(groupedFiles, files)
	}

	var allFiles []*File
	currOrder := 1

	for _, files := range groupedFiles {
		// Only sort files alphanum within a group
		sort.Slice(files, func(i, j int) bool {
			return files[i].RelativePath() < files[j].RelativePath()
		})

		for _, file := range files {
			file.order = currOrder
			currOrder++
		}

		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

func NewSortedFiles(files []*File) []*File {
	currOrder := 1
	for _, file := range files {
		file.order = currOrder
		currOrder++
	}
	return files
}

func NewFileFromSource(fileSrc Source) (*File, error) {
	relPath, err := fileSrc.RelativePath()
	if err != nil {
		return nil, fmt.Errorf("Calculating relative path for '%s': %s", fileSrc, err)
	}

	return &File{src: fileSrc, relPath: filepath.ToSlash(relPath)}, nil
}

func MustNewFileFromSource(fileSrc Source) *File {
	file, err := NewFileFromSource(fileSrc)
	if err != nil {
		panic(err)
	}
	return file
}

func (r *File) Description() string { return r.src.Description() }

func (r *File) OriginalRelativePath() string { return r.relPath }

func (r *File) MarkRelativePath(relPath string) { r.markedRelPath = &relPath }

func (r *File) RelativePath() string {
	if r.markedRelPath != nil {
		return *r.markedRelPath
	}
	return r.relPath
}

func (r *File) Bytes() ([]byte, error) { return r.src.Bytes() }

func (r *File) MarkType(t Type) { r.markedType = &t }

func (r *File) Type() Type {
	if r.markedType != nil {
		return *r.markedType
	}

	switch {
	case r.matchesExt(yamlExts):
		return TypeYAML
	case r.matchesExt(starlarkExts):
		return TypeStarlark
	case r.matchesExt(textExts):
		return TypeText
	default:
		return TypeUnknown
	}
}

func (r *File) MarkForOutput(forOutput bool) { r.markedForOutput = &forOutput }

func (r *File) IsForOutput() bool {
	if r.markedForOutput != nil {
		return *r.markedForOutput
	}
	if r.markedTemplate != nil {
		// it may still be for output, even though it's not a template
		if *r.markedTemplate {
			return true
		}
	}
	return r.isTemplate()
}

func (r *File) MarkTemplate(template bool) { r.markedTemplate = &template }

func (r *File) IsTemplate() bool {
	if r.markedTemplate != nil {
		return *r.markedTemplate
	}
	return r.isTemplate()
}

func (r *File) isTemplate() bool {
	t := r.Type()
	return !r.IsLibrary() && (t == TypeYAML || t == TypeText)
}

func (r *File) IsLibrary() bool {
	exts := strings.Split(filepath.Base(r.RelativePath()), ".")

	if len(exts) > 2 && exts[len(exts)-2] == libraryExt {
		return true
	}

	// make exception for starlark files as they are just pure code
	return r.matchesExt(starlarkExts)
}

func (r *File) matchesExt(exts []string) bool {
	filename := filepath.Base(r.RelativePath())
	for _, ext := range exts {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func (r *File) OrderLess(otherFile *File) bool {
	if r.order == 0 || otherFile.order == 0 {
		panic("Missing file order assignment")
	}
	return r.order < otherFile.order
}

func NewRegularFileLocalSource(path, dir string, fi os.FileInfo, opts SymlinkAllowOpts) (LocalSource, error) {
	isRegFile := (fi.Mode() & os.ModeType) == 0
	isSymlink := (fi.Mode() & os.ModeSymlink) != 0
	isNamedPipe := (fi.Mode() & os.ModeNamedPipe) != 0 // allow pipes (`ytt -f <(echo "---")`)

	switch {
	case isRegFile || isSymlink || isNamedPipe:
		// do nothing
	default:
		return LocalSource{}, fmt.Errorf("Expected file '%s' to be a regular file, but was not", path)
	}

	if isSymlink {
		err := Symlink{path}.IsAllowed(opts)
		if err != nil {
			return LocalSource{}, fmt.Errorf("Checking symlink file '%s': %s", path, err)
		}
	}

	return NewLocalSource(path, dir), nil
}

const (
	pathSeparator = "/"
)

func SplitPath(path string) ([]string, string) {
	pieces := strings.Split(path, pathSeparator)
	if len(pieces) == 1 {
		return nil, pieces[0]
	}
	return pieces[:len(pieces)-1], pieces[len(pieces)-1]
}

func JoinPath(pieces []string) string {
	return strings.Join(pieces, pathSeparator)
}

func IsRootPath(path string) bool {
	return strings.HasPrefix(path, pathSeparator)
}

func StripRootPath(path string) string {
	return path[len(pathSeparator):]
}

func MakeRootPath(path string) string {
	return pathSeparator + path
}
