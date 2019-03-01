package template

import (
	"fmt"

	"go.starlark.net/starlark"
)

type CompiledTemplateLoader interface {
	FindCompiledTemplate(string) (*CompiledTemplate, error)
	Load(*starlark.Thread, string) (starlark.StringDict, error)
	LoadData(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
	ListData(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
}

type NoopCompiledTemplateLoader struct{}

var _ CompiledTemplateLoader = NoopCompiledTemplateLoader{}

func (l NoopCompiledTemplateLoader) FindCompiledTemplate(_ string) (*CompiledTemplate, error) {
	return nil, fmt.Errorf("FindCompiledTemplate is not supported")
}

func (l NoopCompiledTemplateLoader) Load(
	thread *starlark.Thread, module string) (starlark.StringDict, error) {

	return nil, fmt.Errorf("Load is not supported")
}

func (l NoopCompiledTemplateLoader) LoadData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("LoadData is not supported")
}

func (l NoopCompiledTemplateLoader) ListData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("ListData is not supported")
}
