package template_test

import (
	"strings"
	"testing"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
)

func TestTemplatePlainStarlark(t *testing.T) {
	data := `
def hello():
  return "hello"
end
a = hello()
`
	instructions := template.NewInstructionSet()
	compiledTemplate := template.NewCompiledTemplate(
		"stdin", template.NewCodeFromBytes([]byte(data), instructions),
		instructions, template.NewNodes(), template.EvaluationCtxDialects{})

	loader := template.NoopCompiledTemplateLoader{}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	resultGlobals, resultVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		t.Fatalf("Evaluating starlark template: %s", err)
	}

	if resultGlobals["a"] != starlark.String("hello") {
		t.Fatalf("Expected global to be set, but was %#v", resultGlobals["a"])
	}

	if resultVal != nil {
		t.Fatalf("Expected result val to be nil, but was %#v", resultVal)
	}
}

func TestTemplateIgnoreIndentationExceptContinuedLines(t *testing.T) {
	data := `
		if True:
	multiline = "[\
   ]"
		end
`
	instructions := template.NewInstructionSet()
	compiledTemplate := template.NewCompiledTemplate(
		"stdin", template.NewCodeFromBytes([]byte(data), instructions),
		instructions, template.NewNodes(), template.EvaluationCtxDialects{})

	loader := template.NewNoopCompiledTemplateLoader(compiledTemplate)
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	resultGlobals, _, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		t.Fatalf("Evaluating starlark template: %s", err)
	}

	if resultGlobals["multiline"] != starlark.String("[   ]") {
		t.Fatalf("Expected multiline to contain the entire multiline string, but was \"%#v\"", resultGlobals["multiline"])
	}
}

func TestLine1StarlarkError(t *testing.T) {
	data := `badsyntax`
	instructions := template.NewInstructionSet()
	compiledTemplate := template.NewCompiledTemplate(
		"stdin", template.NewCodeFromBytes([]byte(data), instructions),
		instructions, template.NewNodes(), template.EvaluationCtxDialects{})

	loader := template.NewNoopCompiledTemplateLoader(compiledTemplate)
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, _, err := compiledTemplate.Eval(thread, loader)
	if err == nil {
		t.Fatalf("Expected eval to return err")
	}

	matchString := "undefined: badsyntax"
	if !strings.Contains(err.Error(), matchString) {
		t.Fatalf("\nExpected:\n%s\n\nto contain:\n\n%s\n", err.Error(), matchString)
	}
}
