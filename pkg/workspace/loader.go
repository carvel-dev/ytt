package workspace

import (
	"path/filepath"
	"strings"

	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
)

type Loader struct {
	Library *Library
	File    *files.File
	UI      files.UI
	Opts    TemplateLoaderOpts
}

func NewLoader(lib *Library, file *files.File, ui files.UI, opts TemplateLoaderOpts) *Loader {
	return &Loader{lib, file, ui, opts}
}

func (loader *Loader) LoadInternalLibrary(path string, astValues eval.ValuesAst) (*eval.Result, error) {
	if strings.HasPrefix(path, "@") {
		path = privateName + pathSeparator + path[1:]
	}

	absolutePath := filepath.Join(filepath.Dir(loader.File.RelativePath()), path)
	rootLibrary, err := loader.Library.FindRecursiveLibrary(absolutePath, true)
	if err != nil {
		return nil, err
	}

	ll := NewLibraryLoader(rootLibrary, loader.UI, loader.Opts)

	astValues, err = ll.Values(astValues)
	if err != nil {
		return nil, err
	}

	return ll.Eval(astValues)
}
