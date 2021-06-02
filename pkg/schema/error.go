// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const schemaErrorReportTemplate = `
{{.Title}}

{{range .AssertionFailures}}{{.FileName}}:
{{pad "|" ""}}
{{pad "|" .FilePos}} {{.Source}}
{{pad "|" ""}}

{{with .Found}}{{pad "=" ""}} found: {{.}}{{end}}
{{with .Expected}}{{pad "=" ""}} expected: {{.}}{{end}}
{{range .Hints}}{{pad "=" ""}} hint: {{.}}
{{end}}
{{end}}
{{.MiscErrorMessage}}
`

func NewSchemaError(err error) error {
	if typeCheckError, ok := err.(yamlmeta.TypeCheck); ok {
		var failures []assertionFailure
		var miscErrorMessage string
		for _, checkErr := range typeCheckError.Violations {
			if typeCheckAssertionErr, ok := checkErr.(schemaAssertionError); ok {
				failures = append(failures, assertionFailure{
					FileName: typeCheckAssertionErr.position.GetFile(),
					FilePos:  typeCheckAssertionErr.position.AsIntString(),
					Source:   typeCheckAssertionErr.position.GetLine(),
					Expected: typeCheckAssertionErr.expected,
					Found:    typeCheckAssertionErr.found,
					Hints:    typeCheckAssertionErr.hints,
				})
			} else {
				miscErrorMessage += checkErr.Error()
			}
		}
		return &schemaError{
			Title:             fmt.Sprintf("Schema Typecheck - Value is of wrong type"),
			AssertionFailures: failures,
			MiscErrorMessage:  miscErrorMessage,
		}
	}

	if schemaErrorInfo, ok := err.(schemaAssertionError); ok {
		return &schemaError{
			Title: fmt.Sprintf("Invalid schema — %s", schemaErrorInfo.description),
			AssertionFailures: []assertionFailure{{
				FileName: schemaErrorInfo.position.GetFile(),
				FilePos:  schemaErrorInfo.position.AsIntString(),
				Source:   schemaErrorInfo.position.GetLine(),
				Expected: schemaErrorInfo.expected,
				Found:    schemaErrorInfo.found,
				Hints:    schemaErrorInfo.hints,
			}},
		}
	}

	return &schemaError{
		Title:            "Schema Error",
		MiscErrorMessage: err.Error(),
	}
}

func NewMismatchedTypeAssertionError(foundType yamlmeta.TypeWithValues, expectedType yamlmeta.Type) error {
	var expectedTypeString string
	if expectedType.PositionOfDefinition().IsKnown() {
		switch expectedType.(type) {
		case *MapItemType, *ArrayItemType:
			expectedTypeString = expectedType.GetValueType().String()
		default:
			expectedTypeString = expectedType.String()
		}
	}

	return schemaAssertionError{
		position: foundType.GetPosition(),
		expected: fmt.Sprintf("%s (by %s)", expectedTypeString, expectedType.PositionOfDefinition().AsCompactString()),
		found:    foundType.ValueTypeAsString(),
	}
}

func NewUnexpectedKeyAssertionError(found *yamlmeta.MapItem, definition *filepos.Position) error {
	return schemaAssertionError{
		position: definition,
		expected: fmt.Sprintf("(a key defined in map) (by %s)", definition.AsCompactString()),
		found:    fmt.Sprintf("%v", found.Key),
		hints:    []string{"declare data values in schema and override them in a data values document"},
	}
}

type schemaAssertionError struct {
	error
	position    *filepos.Position
	description string
	expected    string
	found       string
	hints       []string
}

type schemaError struct {
	Title             string
	AssertionFailures []assertionFailure
	MiscErrorMessage  string
}

type assertionFailure struct {
	FileName string
	Source   string
	FilePos  string
	Expected string
	Found    string
	Hints    []string
}

func (e schemaError) Error() string {
	maxFilePos := 0
	for _, hunk := range e.AssertionFailures {
		if len(hunk.FilePos) > maxFilePos {
			maxFilePos = len(hunk.FilePos)
		}
	}

	funcMap := template.FuncMap{
		"pad": func(delim string, filePos string) string {
			padding := "  "
			rightAlignedFilePos := fmt.Sprintf("%*s", maxFilePos, filePos)
			return padding + rightAlignedFilePos + " " + delim
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).Parse(schemaErrorReportTemplate)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	output := bytes.NewBufferString("")

	err = tmpl.Execute(output, e)
	if err != nil {
		panic(err.Error())
	}

	return output.String()
}
