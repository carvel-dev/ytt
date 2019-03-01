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
	TypeStarlark
	TypeText
)

type File struct {
	src         Source
	relPath     string
	nonTemplate bool
}

func NewFiles(paths []string, recursive bool) ([]*File, error) {
	var fileSrcs []Source

	for _, path := range paths {
		switch {
		case path == "-":
			fileSrcs = append(fileSrcs, NewStdinSource())

		case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
			fileSrcs = append(fileSrcs, NewHTTPSource(path))

		default:
			fileInfo, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("Checking file '%s'", path)
			}

			if fileInfo.IsDir() {
				if !recursive {
					return nil, fmt.Errorf("Expected file '%s' to not be a directory", path)
				}

				var selectedPaths []string

				err := filepath.Walk(path, func(walkedPath string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() {
						return err
					}
					selectedPaths = append(selectedPaths, walkedPath)
					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("Listing files '%s'", path)
				}

				sort.Strings(selectedPaths)

				for _, selectedPath := range selectedPaths {
					fileSrcs = append(fileSrcs, NewLocalSource(selectedPath, path))
				}
			} else {
				fileSrcs = append(fileSrcs, NewLocalSource(path, ""))
			}
		}
	}

	var files []*File

	for _, fileSrc := range fileSrcs {
		file, err := NewFileFromSource(fileSrc)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func NewFileFromSource(fileSrc Source) (*File, error) {
	relPath, err := fileSrc.RelativePath()
	if err != nil {
		return nil, fmt.Errorf("Calculating relative path for '%s': %s", fileSrc, err)
	}

	return &File{src: fileSrc, relPath: relPath}, nil
}

func MustNewFileFromSource(fileSrc Source) *File {
	file, err := NewFileFromSource(fileSrc)
	if err != nil {
		panic(err)
	}
	return file
}

func (r *File) Description() string    { return r.src.Description() }
func (r *File) RelativePath() string   { return r.relPath }
func (r *File) Bytes() ([]byte, error) { return r.src.Bytes() }

func (r *File) Type() Type {
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

func (r *File) MarkNonTemplate() { r.nonTemplate = true }

func (r *File) IsTemplate() bool {
	return !r.nonTemplate && !r.IsLibrary() && (r.matchesExt(yamlExts) || r.matchesExt(textExts))
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

func SplitPath(path string) ([]string, string) {
	pieces := strings.Split(path, "/")
	if len(pieces) == 1 {
		return nil, pieces[0]
	}
	return pieces[:len(pieces)-1], pieces[len(pieces)-1]
}

func JoinPath(pieces []string) string {
	return strings.Join(pieces, "/")
}
