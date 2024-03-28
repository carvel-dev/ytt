// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
)

type CompiledTemplateLoader interface {
	FindCompiledTemplate(string) *CompiledTemplate
	Load(*starlark.Thread, string) (starlark.StringDict, error)
}

type NoopCompiledTemplateLoader struct {
	tpl *CompiledTemplate
}

func NewNoopCompiledTemplateLoader(tpl *CompiledTemplate) NoopCompiledTemplateLoader {
	return NoopCompiledTemplateLoader{tpl}
}

var _ CompiledTemplateLoader = NoopCompiledTemplateLoader{}

// FindCompiledTemplate returns the CompiledTemplate this CompiledTemplateLoader was constructed with.
func (l NoopCompiledTemplateLoader) FindCompiledTemplate(_ string) *CompiledTemplate {
	return l.tpl
}

func (l NoopCompiledTemplateLoader) Load(
	thread *starlark.Thread, module string) (starlark.StringDict, error) {

	return nil, fmt.Errorf("Load is not supported")
}
