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
}

func NewFiles(paths []string) ([]*File, error) {
	var fileSrcs []Source

	for _, path := range paths {
		switch {
		case path == "-":
			fileSrcs = append(fileSrcs, NewStdinSource())

		case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
			fileSrcs = append(fileSrcs, NewHTTPSource(path))

		default:
			fileInfo, err := os.Lstat(path)
			if err != nil {
				return nil, fmt.Errorf("Checking file '%s'", path)
			}

			if fileInfo.IsDir() {
				err := filepath.Walk(path, func(walkedPath string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() {
						return err
					}
					regLocalSource, err := NewRegularFileLocalSource(walkedPath, path, fi)
					if err != nil {
						return err
					}
					fileSrcs = append(fileSrcs, regLocalSource)
					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("Listing files '%s': %s", path, err)
				}
			} else {
				regLocalSource, err := NewRegularFileLocalSource(path, "", fileInfo)
				if err != nil {
					return nil, err
				}
				fileSrcs = append(fileSrcs, regLocalSource)
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

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelativePath() < files[j].RelativePath()
	})

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

func NewRegularFileLocalSource(path, dir string, fi os.FileInfo) (LocalSource, error) {
	// support pipes (`ytt template -f <(echo "---")`)
	namedPipe := (fi.Mode() & os.ModeNamedPipe) != 0

	if !namedPipe && (fi.Mode()&os.ModeType) != 0 {
		return LocalSource{}, fmt.Errorf("Expected file '%s' to be a regular file, but was not", path)
	}
	return NewLocalSource(path, dir), nil
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
