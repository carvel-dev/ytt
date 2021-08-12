// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const schemaErrorReportTemplate = `
{{- if .Summary}}
{{.Summary}}
{{addBreak .Summary}}
{{- end}}
{{- range .AssertionFailures}}
{{- if .Description}}
{{.Description}}
{{- end}}

{{- if .FromMemory}}

{{.SourceName}}:
{{pad "#" ""}}
{{pad "#" ""}} {{.Source}}
{{pad "#" ""}}
{{- else}}

{{.FileName}}:
{{pad "|" ""}}
{{pad "|" .FilePos}} {{.Source}}
{{pad "|" ""}}
{{- end}}

{{with .Found}}{{pad "=" ""}} found: {{.}}{{end}}
{{with .Expected}}{{pad "=" ""}} expected: {{.}}{{end}}
{{- range .Hints}}
{{pad "=" ""}} hint: {{.}}
{{- end}}
{{- end}}
{{.MiscErrorMessage}}
`

func NewSchemaError(summary string, errs ...error) error {
	var failures []assertionFailure
	var miscErrorMessage string
	for _, err := range errs {
		if typeCheckAssertionErr, ok := err.(schemaAssertionError); ok {
			failures = append(failures, assertionFailure{
				Description: typeCheckAssertionErr.description,
				FileName:    typeCheckAssertionErr.position.GetFile(),
				FilePos:     typeCheckAssertionErr.position.AsIntString(),
				FromMemory:  typeCheckAssertionErr.position.FromMemory(),
				SourceName:  "Data value calculated",
				Source:      typeCheckAssertionErr.position.GetLine(),
				Expected:    typeCheckAssertionErr.expected,
				Found:       typeCheckAssertionErr.found,
				Hints:       typeCheckAssertionErr.hints,
			})
		} else {
			miscErrorMessage += fmt.Sprintf("%s \n", err.Error())
		}
	}

	return &schemaError{
		Summary:           summary,
		AssertionFailures: failures,
		MiscErrorMessage:  miscErrorMessage,
	}
}

func NewMismatchedTypeAssertionError(foundType yamlmeta.TypeWithValues, expectedType yamlmeta.Type) error {
	var expectedTypeString string
	if expectedType.GetDefinitionPosition().IsKnown() {
		switch expectedType.(type) {
		case *MapItemType, *ArrayItemType:
			expectedTypeString = expectedType.GetValueType().String()
		default:
			expectedTypeString = expectedType.String()
		}
	}

	return schemaAssertionError{
		position: foundType.GetPosition(),
		expected: fmt.Sprintf("%s (by %s)", expectedTypeString, expectedType.GetDefinitionPosition().AsCompactString()),
		found:    foundType.ValueTypeAsString(),
	}
}

func NewUnexpectedKeyAssertionError(found *yamlmeta.MapItem, definition *filepos.Position) error {
	return schemaAssertionError{
		position: found.GetPosition(),
		expected: fmt.Sprintf("(a key defined in map) (by %s)", definition.AsCompactString()),
		found:    fmt.Sprintf("%v", found.Key),
		hints:    []string{"declare data values in schema and override them in a data values document"},
	}
}

type schemaError struct {
	Summary           string
	AssertionFailures []assertionFailure
	MiscErrorMessage  string
}

type assertionFailure struct {
	Description string
	FileName    string
	Source      string
	FilePos     string
	FromMemory  bool
	SourceName  string
	Expected    string
	Found       string
	Hints       []string
}

type schemaAssertionError struct {
	error
	position    *filepos.Position
	description string
	expected    string
	found       string
	hints       []string
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
		"addBreak": func(title string) string {
			return strings.Repeat("=", len(title))
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
