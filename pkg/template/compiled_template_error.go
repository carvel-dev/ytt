package template

import (
	"fmt"
	"strings"

	"github.com/get-ytt/ytt/pkg/filepos"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type CompiledTemplateMultiError struct {
	errs   []CompiledTemplateError
	loader CompiledTemplateLoader
}

var _ error = CompiledTemplateMultiError{}

type CompiledTemplateError struct {
	Positions []CompiledTemplateErrorPosition
	Msg       string
}

type CompiledTemplateErrorPosition struct {
	Filename     string
	ContextName  string
	TemplateLine *TemplateLine
}

func NewCompiledTemplateMultiError(err error, loader CompiledTemplateLoader) error {
	e := CompiledTemplateMultiError{loader: loader}

	switch typedErr := err.(type) {
	case syntax.Error:
		e.errs = append(e.errs, CompiledTemplateError{
			Positions: []CompiledTemplateErrorPosition{e.buildPos(typedErr.Pos)},
			Msg:       typedErr.Msg,
		})

	case resolve.ErrorList:
		for _, resolveErr := range typedErr {
			e.errs = append(e.errs, CompiledTemplateError{
				Positions: []CompiledTemplateErrorPosition{e.buildPos(resolveErr.Pos)},
				Msg:       resolveErr.Msg,
			})
		}

	case *starlark.EvalError:
		e.errs = append(e.errs, e.buildEvalErr(typedErr))

	default:
		e.errs = append(e.errs, CompiledTemplateError{Msg: err.Error()})
	}

	return e
}

func (e CompiledTemplateMultiError) Error() string {
	result := []string{""}

	for _, err := range e.errs {
		var topicLine string
		var otherLines []string

		if !strings.Contains(err.Msg, "\n") {
			topicLine = err.Msg
		} else {
			for i, line := range strings.Split(err.Msg, "\n") {
				if i == 0 {
					topicLine = line
				} else {
					otherLines = append(otherLines, line)
				}
			}
		}

		result = append(result, fmt.Sprintf("- %s", topicLine))

		for _, pos := range err.Positions {
			// TODO do better
			if pos.TemplateLine == nil {
				continue
			}

			ctx := ""
			if len(pos.ContextName) > 0 {
				ctx = " in " + pos.ContextName
			}

			// TODO show column information
			result = append(result, fmt.Sprintf("    %s:%s%s",
				pos.Filename, pos.TemplateLine.Position().AsIntString(), ctx))

			if pos.TemplateLine.SourceLine != nil {
				result = append(result, fmt.Sprintf("     L %s",
					pos.TemplateLine.SourceLine.Content))
			} else {
				result = append(result, fmt.Sprintf("     L %s (generated)",
					pos.TemplateLine.Instruction.AsString()))
			}
		}

		if len(otherLines) > 0 {
			result = append(result, []string{"", fmt.Sprintf("    reason:")}...)
			for _, line := range otherLines {
				result = append(result, fmt.Sprintf("     %s", line))
			}
		}
	}

	return strings.Join(result, "\n")
}

func (e CompiledTemplateMultiError) buildEvalErr(err *starlark.EvalError) CompiledTemplateError {
	// fmt.Printf("frame:\n%s\n", err.Backtrace())
	result := CompiledTemplateError{Msg: err.Msg}
	for _, frame := range err.Stack() {
		pos := e.buildPos(frame.Position())
		pos.ContextName = frame.Callable().Name()
		result.Positions = append(result.Positions, pos)
	}
	return result
}

func (e CompiledTemplateMultiError) buildPos(pos syntax.Position) CompiledTemplateErrorPosition {
	// TODO seems to be a bug in starlark where, for example,
	// "function call2 takes exactly 1 positional argument (0 given)"
	// error has 0 line number position (even though its 1 based)
	if pos.Line == 0 {
		return CompiledTemplateErrorPosition{}
	}

	ct, err := e.loader.FindCompiledTemplate(pos.Filename())
	if err != nil {
		panic(fmt.Errorf("Expected to find compiled template: %s", err))
	}

	line := ct.CodeAtLine(filepos.NewPosition(int(pos.Line)))
	if line != nil {
		return CompiledTemplateErrorPosition{
			Filename:     pos.Filename(),
			TemplateLine: line,
		}
	}

	panic(fmt.Errorf("Expected to find compiled template line %d", pos.Line))
}
