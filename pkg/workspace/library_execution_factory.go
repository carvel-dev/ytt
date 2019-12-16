package workspace

import (
	"github.com/k14s/ytt/pkg/files"
)

type LibraryExecutionContext struct {
	Current *Library
	Root    *Library
}

type LibraryExecutionFactory struct {
	ui                 files.UI
	templateLoaderOpts TemplateLoaderOpts
}

func NewLibraryExecutionFactory(ui files.UI, templateLoaderOpts TemplateLoaderOpts) *LibraryExecutionFactory {
	return &LibraryExecutionFactory{ui, templateLoaderOpts}
}

func (f *LibraryExecutionFactory) New(ctx LibraryExecutionContext) *LibraryLoader {
	return NewLibraryLoader(ctx, f.ui, f.templateLoaderOpts, f)
}
