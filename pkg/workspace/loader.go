package workspace

import (
	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
)

type Loader struct {
	Package *PackageInLibrary
	File    *files.File
	UI      files.UI
	Opts    TemplateLoaderOpts
}

func NewLoader(pkg *PackageInLibrary, file *files.File, ui files.UI, opts TemplateLoaderOpts) *Loader {
	return &Loader{pkg, file, ui, opts}
}

func (loader *Loader) LoadInternalLibrary(path string, astValues ...eval.ValuesAst) (*eval.Result, error) {
	fileInLib, pkgInLib, err := loader.Package.FindLoadPath(path)
	if err != nil {
		return nil, err
	}

	var files []*FileInLibrary
	if fileInLib != nil {
		files = append(files, fileInLib)
	} else if pkgInLib.Package() != pkgInLib.Library() {
		files = pkgInLib.ListAccessibleFiles()
	}

	ll := NewLibraryLoader(pkgInLib.Library(), files, loader.UI, loader.Opts)

	vals, err := ll.Values(astValues...)
	if err != nil {
		return nil, err
	}

	return ll.Eval(vals)
}
