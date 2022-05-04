// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
)

// LibraryExecutionContext holds the total set of inputs that are involved in a LibraryExecution.
type LibraryExecutionContext struct {
	Current *Library // the target library that will be executed/evaluated.
	Root    *Library // reference to the root library to support accessing "absolute path" loading of files.
}

// LibraryExecutionFactory holds configuration for and produces instances of LibraryExecution's.
type LibraryExecutionFactory struct {
	ui                 ui.UI
	templateLoaderOpts TemplateLoaderOpts

	skipDataValuesValidation bool
}

// NewLibraryExecutionFactory configures a new instance of a LibraryExecutionFactory.
func NewLibraryExecutionFactory(ui ui.UI, templateLoaderOpts TemplateLoaderOpts, skipDataValuesValidation bool) *LibraryExecutionFactory {
	return &LibraryExecutionFactory{ui, templateLoaderOpts, skipDataValuesValidation}
}

// WithTemplateLoaderOptsOverrides produces a new LibraryExecutionFactory identical to this one, except it configures
// its TemplateLoader with the merge of the supplied TemplateLoaderOpts over this factory's configuration.
func (f *LibraryExecutionFactory) WithTemplateLoaderOptsOverrides(overrides TemplateLoaderOptsOverrides) *LibraryExecutionFactory {
	return NewLibraryExecutionFactory(f.ui, f.templateLoaderOpts.Merge(overrides), f.skipDataValuesValidation)
}

// ThatSkipsDataValuesValidations produces a new LibraryExecutionFactory identical to this one, except it might also
// skip running validation rules over Data Values.
func (f *LibraryExecutionFactory) ThatSkipsDataValuesValidations(skipDataValuesValidation bool) *LibraryExecutionFactory {
	return NewLibraryExecutionFactory(f.ui, f.templateLoaderOpts, f.skipDataValuesValidation || skipDataValuesValidation)
}

// New produces a new instance of a LibraryExecution, set with the configuration and dependencies of this factory.
func (f *LibraryExecutionFactory) New(ctx LibraryExecutionContext) *LibraryExecution {
	return NewLibraryExecution(ctx, f.ui, f.templateLoaderOpts, f, f.skipDataValuesValidation)
}
