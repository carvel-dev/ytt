package template_test

import (
	"testing"

	"github.com/k14s/ytt/pkg/template"
	"go.starlark.net/starlark"
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

	resultGlobals, resultVal, _, err := compiledTemplate.Eval(thread, loader)
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
