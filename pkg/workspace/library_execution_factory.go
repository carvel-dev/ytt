// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/k14s/ytt/pkg/cmd/ui"
)

type LibraryExecutionContext struct {
	Current *Library
	Root    *Library
}

type LibraryExecutionFactory struct {
	ui                 ui.UI
	templateLoaderOpts TemplateLoaderOpts
}

func NewLibraryExecutionFactory(ui ui.UI, templateLoaderOpts TemplateLoaderOpts) *LibraryExecutionFactory {
	return &LibraryExecutionFactory{ui, templateLoaderOpts}
}

func (f *LibraryExecutionFactory) WithTemplateLoaderOptsOverrides(overrides TemplateLoaderOptsOverrides) *LibraryExecutionFactory {
	return NewLibraryExecutionFactory(f.ui, f.templateLoaderOpts.Merge(overrides))
}

func (f *LibraryExecutionFactory) New(ctx LibraryExecutionContext) *LibraryExecution {
	return NewLibraryExecution(ctx, f.ui, f.templateLoaderOpts, f)
}
