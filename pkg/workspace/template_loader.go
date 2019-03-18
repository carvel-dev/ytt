package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/texttemplate"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"go.starlark.net/starlark"
)

type TemplateLoader struct {
	ui                files.UI
	values            interface{}
	compiledTemplates map[string]*template.CompiledTemplate
}

func NewTemplateLoader(values interface{}, ui files.UI) *TemplateLoader {
	return &TemplateLoader{
		ui:                ui,
		values:            values,
		compiledTemplates: map[string]*template.CompiledTemplate{},
	}
}

func (l *TemplateLoader) FindCompiledTemplate(path string) (*template.CompiledTemplate, error) {
	ct, found := l.compiledTemplates[path]
	if !found {
		return nil, fmt.Errorf("Expected to find '%s' compiled template", path)
	}
	return ct, nil
}

func (l *TemplateLoader) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	if api, found := l.getYTTLibrary(thread)[module]; found {
		return api, nil
	}

	library := l.getLibrary(thread)
	filePath := module

	if strings.HasPrefix(module, "@") {
		pieces := strings.SplitN(module[1:], ":", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected library path to be in format '@name:path', " +
				" for example, '@github.com/k14s/test:test.star'")
		}

		foundLib, err := library.FindAccessibleLibrary(pieces[0])
		if err != nil {
			return nil, err
		}

		library = foundLib
		filePath = pieces[1]
	}

	file, err := library.FindFile(filePath)
	if err != nil {
		return nil, err
	}

	if !file.IsLibrary() {
		return nil, fmt.Errorf("File '%s' is not a library file "+
			"(use data.read(...) for loading non-templated file contents into a variable)", file.RelativePath())
	}

	switch file.Type() {
	case files.TypeYAML:
		globals, _, err := l.EvalYAML(library, file)
		return globals, err

	case files.TypeStarlark:
		return l.EvalStarlark(library, file)

	case files.TypeText:
		globals, _, err := l.EvalText(library, file)
		return globals, err

	default:
		return nil, fmt.Errorf("File '%s' type is not a known", file.RelativePath())
	}
}

func (l *TemplateLoader) ListData(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected exactly zero arguments")
	}

	result := []starlark.Value{}
	for _, file := range l.getLibrary(thread).ListAccessibleFiles() {
		result = append(result, starlark.String(file.RelativePath()))
	}
	return starlark.NewList(result), nil
}

func (l *TemplateLoader) LoadData(thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	path, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	file, err := l.getLibrary(thread).FindFile(path)
	if err != nil {
		return nil, err
	}

	fileBs, err := file.Bytes()
	if err != nil {
		return nil, err
	}

	return starlark.String(string(fileBs)), nil
}

func (l *TemplateLoader) EvalYAML(library *Library, file *files.File) (starlark.StringDict, interface{}, error) {
	fileBs, err := file.Bytes()
	if err != nil {
		return nil, nil, err
	}

	l.ui.Debugf("## file %s\n", file.RelativePath())

	docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, file.RelativePath())
	if err != nil {
		return nil, nil, fmt.Errorf("Unmarshaling YAML template: %s", err)
	}

	l.ui.Debugf("### ast\n")
	docSet.Print(l.ui.DebugWriter())

	compiledTemplate, err := yamltemplate.NewTemplate(file.RelativePath()).Compile(docSet)
	if err != nil {
		return nil, nil, err
	}

	l.addCompiledTemplate(file.RelativePath(), compiledTemplate)
	l.ui.Debugf("### template\n%s", compiledTemplate.DebugCodeAsString())

	yttLibrary := yttlibrary.NewAPI(compiledTemplate.TplReplaceNode, l.values, l)
	thread := l.newThread(library, yttLibrary, file)

	globals, resultVal, err := compiledTemplate.Eval(thread, l)
	if err != nil {
		return nil, nil, err
	}

	return globals, resultVal, nil
}

func (l *TemplateLoader) EvalText(library *Library, file *files.File) (starlark.StringDict, interface{}, error) {
	fileBs, err := file.Bytes()
	if err != nil {
		return nil, nil, err
	}

	l.ui.Debugf("## file %s\n", file.RelativePath())

	textRoot, err := texttemplate.NewParser().Parse(fileBs, file.RelativePath())
	if err != nil {
		return nil, nil, err
	}

	compiledTemplate, err := texttemplate.NewTemplate(file.RelativePath()).Compile(textRoot)
	if err != nil {
		return nil, nil, err
	}

	l.addCompiledTemplate(file.RelativePath(), compiledTemplate)
	l.ui.Debugf("### template\n%s", compiledTemplate.DebugCodeAsString())

	yttLibrary := yttlibrary.NewAPI(compiledTemplate.TplReplaceNode, l.values, l)
	thread := l.newThread(library, yttLibrary, file)

	globals, resultVal, err := compiledTemplate.Eval(thread, l)
	if err != nil {
		return nil, nil, fmt.Errorf("Evaluating text template: %s", err)
	}

	return globals, resultVal, nil
}

func (l *TemplateLoader) EvalStarlark(library *Library, file *files.File) (starlark.StringDict, error) {
	fileBs, err := file.Bytes()
	if err != nil {
		return nil, err
	}

	l.ui.Debugf("## file %s\n", file.RelativePath())

	instructions := template.NewInstructionSet()
	compiledTemplate := template.NewCompiledTemplate(
		file.RelativePath(), template.NewCodeFromBytes(fileBs, instructions),
		instructions, template.NewNodes(), template.EvaluationCtxDialects{})

	l.addCompiledTemplate(file.RelativePath(), compiledTemplate)
	l.ui.Debugf("### template\n%s", compiledTemplate.DebugCodeAsString())

	yttLibrary := yttlibrary.NewAPI(compiledTemplate.TplReplaceNode, l.values, l)
	thread := l.newThread(library, yttLibrary, file)

	globals, _, err := compiledTemplate.Eval(thread, l)
	if err != nil {
		return nil, fmt.Errorf("Evaluating starlark template: %s", err)
	}

	return globals, nil
}

const (
	threadLibraryKey    = "ytt.library_key"
	threadYTTLibraryKey = "ytt.ytt_library_key"
)

func (l *TemplateLoader) getLibrary(thread *starlark.Thread) *Library {
	lib, ok := thread.Local(threadLibraryKey).(*Library)
	if !ok {
		panic("Expected to find library associated with thread")
	}
	return lib
}

func (l *TemplateLoader) setLibrary(thread *starlark.Thread, library *Library) {
	thread.SetLocal(threadLibraryKey, library)
}

func (l *TemplateLoader) getYTTLibrary(thread *starlark.Thread) yttlibrary.API {
	yttLibrary, ok := thread.Local(threadYTTLibraryKey).(yttlibrary.API)
	if !ok {
		panic("Expected to find YTT library associated with thread")
	}
	return yttLibrary
}

func (l *TemplateLoader) setYTTLibrary(thread *starlark.Thread, yttLibrary yttlibrary.API) {
	thread.SetLocal(threadYTTLibraryKey, yttLibrary)
}

func (l *TemplateLoader) newThread(library *Library, yttLibrary yttlibrary.API, file *files.File) *starlark.Thread {
	thread := &starlark.Thread{Name: "template=" + file.RelativePath(), Load: l.Load}
	l.setLibrary(thread, library)
	l.setYTTLibrary(thread, yttLibrary)
	return thread
}

func (l *TemplateLoader) addCompiledTemplate(path string, ct *template.CompiledTemplate) {
	l.compiledTemplates[path] = ct
}
