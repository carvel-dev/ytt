package workspace

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/k14s/ytt/pkg/files"
)

const (
	libLocation   = "_ytt_lib"
	privateName   = libLocation
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

func (l *Library) findRecursiveLibrary(path string, pieces []string, allowPrivate bool) (*Library, []*Library, error) {
	var currLibrary *Library = l
	var libPath []*Library
	for i, piece := range pieces {
		lib, found := currLibrary.FindLibrary(piece)
		if !found {
			return nil, nil, fmt.Errorf("Expected to find '%s', but did not find '%s'",
				path, files.JoinPath(pieces[:i+1]))
		}
		if lib.private && !allowPrivate {
			return nil, nil, fmt.Errorf("Could not load file '%s' because it's contained in private library '%s' "+
				"(use load(\"@lib:file\", \"symbol\") where 'lib' is library name under %s, for example, 'github.com/k14s/test')",
				path, files.JoinPath(pieces[:i+1]), privateName)
		}
		libPath = append(libPath, lib)
		currLibrary = lib
	}

	return currLibrary, libPath, nil
}

func (l *Library) FindFile(path string) (*files.File, error) {
	dirPieces, namePiece := files.SplitPath(path)

	currLibrary, _, err := l.findRecursiveLibrary(path, dirPieces, false)
	if err != nil {
		return nil, err
	}

	for _, file := range currLibrary.files {
		_, fileNamePiece := files.SplitPath(file.RelativePath())
		if fileNamePiece == namePiece {
			return file, nil
		}
	}
	return nil, fmt.Errorf("Expected to find file %s", path)
}

type PackageInLibrary struct {
	pkgPath []*Library
}

func (packageInLib *PackageInLibrary) Package() *Library {
	return packageInLib.pkgPath[len(packageInLib.pkgPath)-1]
}

func (packageInLib *PackageInLibrary) Library() *Library {
	return packageInLib.pkgPath[0]
}

var LoaderRegexp = regexp.MustCompile("^((@?)([^:]*):)?(.*)$")

func (packageInLib *PackageInLibrary) FindLoadPath(path string) (*FileInLibrary, *PackageInLibrary, error) {
	matches := LoaderRegexp.FindStringSubmatch(path)

	pkg := &PackageInLibrary{}

	if matches[1] == "" {
		pkg.pkgPath = packageInLib.pkgPath
	} else if matches[1] == "@:" {
		pkg.pkgPath = []*Library{packageInLib.Library()}
	} else {
		libPath, libPathLast := files.SplitPath(matches[3])
		libPath = append(libPath, libPathLast)
		if matches[2] == "@" {
			libPath = append([]string{libLocation}, libPath...)
		}
		lib, _, err := packageInLib.Library().findRecursiveLibrary(strings.Join(libPath, pathSeparator), libPath, true)
		if err != nil {
			return nil, nil, fmt.Errorf("While loading '%s', %v", path, err)
		}
		pkg.pkgPath = []*Library{lib}
	}

	dirPieces, namePiece := files.SplitPath(matches[4])
	_, packagePath, err := pkg.Package().findRecursiveLibrary(path, dirPieces, false)
	if err != nil {
		return nil, nil, fmt.Errorf("While loading '%s', %v", path, err)
	}
	pkg.pkgPath = append(pkg.pkgPath, packagePath...)

	if namePiece == "" {
		return nil, pkg, nil
	}

	for _, file := range pkg.Package().files {
		_, fileNamePiece := files.SplitPath(file.RelativePath())
		if fileNamePiece == namePiece {
			return &FileInLibrary{pkg, file}, pkg, nil
		}
	}

	return nil, nil, fmt.Errorf("While loading '%s' expected to find file %s", path, namePiece)
}

func (packageInLib *PackageInLibrary) packagePath() []string {
	var components []string
	for _, lib := range packageInLib.pkgPath[1:] {
		components = append(components, lib.name)
	}
	return components
}

func (packageInLib *PackageInLibrary) ListAccessibleFiles() []*FileInLibrary {
	return packageInLib.Package().listAccessibleFiles(packageInLib.pkgPath)
}

type FileInLibrary struct {
	*PackageInLibrary
	File *files.File
}

func (fileInLib *FileInLibrary) RelativePath() string {
	pkgPath := fileInLib.PackageInLibrary.packagePath()
	_, fileName := files.SplitPath(fileInLib.File.RelativePath())
	return strings.Join(append(pkgPath, fileName), pathSeparator)
}

func (l *Library) ListAccessibleFiles() []*FileInLibrary {
	return l.listAccessibleFiles([]*Library{l})
}

func (l *Library) listAccessibleFiles(parents []*Library) []*FileInLibrary {
	var result []*FileInLibrary
	for _, file := range l.files {
		result = append(result, &FileInLibrary{
			PackageInLibrary: &PackageInLibrary{append([]*Library{}, parents...)},
			File:             file,
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
