package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
)

type Loader struct {
	Library *Library
	File    *files.File
	UI      files.UI
	Opts    eval.TemplateLoaderOpts
}

func NewLoader(lib *Library, file *files.File, ui files.UI, opts eval.TemplateLoaderOpts) *Loader {
	return &Loader{lib, file, ui, opts}
}

func (loader *Loader) LoadInternalLibrary(path string, opts *eval.TemplateLoaderOpts) (eval.Library, error) {
	if opts == nil {
		opts = &loader.Opts
	}

	if strings.HasPrefix(path, "@") {
		path = privateName + pathSeparator + path[1:]
	}

	absolutePath := filepath.Join(filepath.Dir(loader.File.RelativePath()), path)
	rootLibrary, err := loader.Library.FindRecursiveLibrary(absolutePath, true)
	if err != nil {
		return nil, err
	}

	library, err := LoadLibrary(rootLibrary, *opts)
	if err != nil {
		return nil, err
	}

	library.UI = loader.UI

	return library, nil
}

func (loader *Loader) LoadExternalLibrary(paths []string, recursive bool, opts *eval.TemplateLoaderOpts) (eval.Library, error) {
	if opts == nil {
		opts = &loader.Opts
	}

	sourcePath := loader.File.AbsolutePath()
	if sourcePath == "" {
		return nil, fmt.Errorf("Impossible to load external library with no context")
	}

	var absolutePaths []string
	for _, path := range paths {
		absolutePaths = append(absolutePaths, filepath.Join(filepath.Dir(sourcePath), path))
	}

	return LoadRootLibrary(absolutePaths, recursive, loader.UI, *opts)
}

func LoadRootLibrary(absolutePaths []string, recursive bool, ui files.UI, opts eval.TemplateLoaderOpts) (eval.Library, error) {
	filesToProcess, err := files.NewFiles(absolutePaths, true)
	if err != nil {
		return nil, err
	}

	rootLibrary := NewRootLibrary(filesToProcess, recursive)

	library, err := LoadLibrary(rootLibrary, opts)
	if err != nil {
		return nil, err
	}

	library.UI = ui

	return library, nil
}
