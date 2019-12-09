package workspace

import (
	"github.com/k14s/ytt/pkg/files"
)

type LibraryExecutionFactory struct {
	ui                 files.UI
	templateLoaderOpts TemplateLoaderOpts
}

func NewLibraryExecutionFactory(ui files.UI, templateLoaderOpts TemplateLoaderOpts) *LibraryExecutionFactory {
	return &LibraryExecutionFactory{ui, templateLoaderOpts}
}

func (f *LibraryExecutionFactory) New(library *Library) *LibraryLoader {
	return NewLibraryLoader(library, f.ui, f.templateLoaderOpts, f)
}
