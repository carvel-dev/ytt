package template_test

import (
	"testing"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
)

type ErrorHintTest struct {
	Input  string
	ErrMsg string
}

func TestErrorHints(t *testing.T) {
	cases := []ErrorHintTest{
		{
			Input: `v = true`,
			ErrMsg: `
- undefined: true (hint: use 'True' instead of 'true' for boolean assignment)
    2 | v = true`,
		},
		{
			Input: `v = false`,
			ErrMsg: `
- undefined: false (hint: use 'False' instead of 'false' for boolean assignment)
    2 | v = false`,
		},
		{
			Input: `v = null`,
			ErrMsg: `
- undefined: null (hint: use 'None' instead of 'null' to indicate no value)
    2 | v = null`,
		},
		{
			Input: `v = none`,
			ErrMsg: `
- undefined: none (hint: use 'None' instead of 'none' to indicate no value)
    2 | v = none`,
		},
		{
			Input: `v = nil`,
			ErrMsg: `
- undefined: nil (hint: use 'None' instead of 'nil' to indicate no value)
    2 | v = nil`,
		},
		{
			Input: `if True
  v = 123
end`,
			ErrMsg: `
- got newline, want ':' (hint: missing colon at the end of 'if/for/def' statement?)
    3 |   v = 123`,
		},
		{
			Input: `def foo()
  return 123
end`,
			ErrMsg: `
- got newline, want ':' (hint: missing colon at the end of 'if/for/def' statement?)
    3 |   return 123`,
		},
	}

	for _, cs := range cases {
		instructions := template.NewInstructionSet()
		compiledTemplate := template.NewCompiledTemplate(
			"stdin", template.NewCodeFromBytes([]byte("\n"+cs.Input), instructions),
			instructions, template.NewNodes(), template.EvaluationCtxDialects{})

		loader := template.NewNoopCompiledTemplateLoader(compiledTemplate)
		thread := &starlark.Thread{Name: "test", Load: loader.Load}

		_, _, err := compiledTemplate.Eval(thread, loader)
		if err == nil {
			t.Fatalf("Expected starlark template error, but was nil")
		}

		if err.Error() != cs.ErrMsg {
			t.Fatalf("Expected error to match but did not; expected >>>%s<<< vs actual >>>%s<<<", cs.ErrMsg, err)
		}
	}
}
