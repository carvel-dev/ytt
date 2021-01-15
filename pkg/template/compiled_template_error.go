// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/resolve"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/syntax"
	"github.com/k14s/ytt/pkg/filepos"
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
	TemplateLine *Line

	BeforeTemplateLine *Line
	AfterTemplateLine  *Line
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

		result = append(result, fmt.Sprintf("- %s%s", topicLine, e.hintMsg(topicLine)))

		for _, pos := range err.Positions {
			// TODO do better
			if pos.TemplateLine == nil {
				continue
			}

			linePad := "    "

			if len(pos.ContextName) > 0 {
				result = append(result, linePad+"in "+pos.ContextName)
				linePad += "  "
			}

			if pos.TemplateLine.SourceLine != nil {
				if pos.TemplateLine.SourceLine.Selection != nil {
					result = append(result, fmt.Sprintf("%s%s%s",
						linePad, e.posPrefixStr(pos.TemplateLine.SourceLine.Selection), pos.TemplateLine.SourceLine.Selection.Content))
				} else {
					result = append(result, fmt.Sprintf("%s%s%s",
						linePad, e.posPrefixStr(pos.TemplateLine.SourceLine), pos.TemplateLine.SourceLine.Content))
				}
			} else {
				if pos.BeforeTemplateLine != nil && pos.BeforeTemplateLine.SourceLine != nil {
					result = append(result, fmt.Sprintf("%s%s%s",
						linePad, e.posPrefixStr(pos.BeforeTemplateLine.SourceLine), pos.BeforeTemplateLine.SourceLine.Content))
				}

				result = append(result, fmt.Sprintf("%s%s:? | %s (generated)",
					linePad, pos.Filename, pos.TemplateLine.Instruction.AsString()))

				if pos.AfterTemplateLine != nil && pos.AfterTemplateLine.SourceLine != nil {
					result = append(result, fmt.Sprintf("%s%s%s",
						linePad, e.posPrefixStr(pos.AfterTemplateLine.SourceLine), pos.AfterTemplateLine.SourceLine.Content))
				}
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

func (e CompiledTemplateMultiError) posPrefixStr(srcLine *SourceLine) string {
	// TODO show column information
	return fmt.Sprintf("%s | ", srcLine.Position.AsCompactString())
}

func (e CompiledTemplateMultiError) buildEvalErr(err *starlark.EvalError) CompiledTemplateError {
	// fmt.Printf("frame:\n%s\n", err.Backtrace())
	result := CompiledTemplateError{Msg: err.Msg}
	for i := len(err.CallStack) - 1; i >= 0; i-- {
		pos := e.buildPos(err.CallStack[i].Pos)
		pos.ContextName = err.CallStack[i].Name
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
	if line == nil {
		panic(fmt.Errorf("Expected to find compiled template line %d", pos.Line))
	}

	return CompiledTemplateErrorPosition{
		Filename:           pos.Filename(),
		TemplateLine:       line,
		BeforeTemplateLine: e.findClosestLine(ct, int(pos.Line), -1),
		AfterTemplateLine:  e.findClosestLine(ct, int(pos.Line), 1),
	}
}

func (CompiledTemplateMultiError) findClosestLine(ct *CompiledTemplate, posLine int, lineInc int) *Line {
	for {
		posLine += lineInc
		if posLine < 1 {
			return nil
		}

		line := ct.CodeAtLine(filepos.NewPosition(posLine))
		if line == nil || line.SourceLine != nil {
			return line
		}
	}
}

func (CompiledTemplateMultiError) hintMsg(errMsg string) string {
	hintMsg := ""
	switch errMsg {
	case "undefined: true":
		hintMsg = "use 'True' instead of 'true' for boolean assignment"
	case "undefined: false":
		hintMsg = "use 'False' instead of 'false' for boolean assignment"
	case "got newline, want ':'":
		hintMsg = "missing colon at the end of 'if/for/def' statement?"
	case "undefined: null":
		hintMsg = "use 'None' instead of 'null' to indicate no value"
	case "undefined: nil":
		hintMsg = "use 'None' instead of 'nil' to indicate no value"
	case "undefined: none":
		hintMsg = "use 'None' instead of 'none' to indicate no value"
	case "unhandled index operation struct[string]":
		hintMsg = "use getattr(...) to access struct field programmatically"
	}

	if len(hintMsg) > 0 {
		hintMsg = fmt.Sprintf(" (hint: %s)", hintMsg)
	}
	return hintMsg
}
