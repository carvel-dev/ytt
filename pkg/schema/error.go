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
{{.Title}}

{{.FileName}}:
{{pad "|" false}}
{{pad "|" true}} {{.Diff}}
{{pad "|" false}}

{{with .Found}}{{pad "=" false}} found: {{.}}{{end}}
{{with .Expected}}{{pad "=" false}} expected: {{.}}{{end}}
{{range .Hints}}{{pad "=" false}} hint: {{.}}
{{end}}`

func NewSchemaError(err error, node yamlmeta.Node) error {
	if schemaErrorInfo, ok := err.(schemaAssertionError); ok {
		return &schemaError{
			Title:    fmt.Sprintf("Invalid schema — %s", schemaErrorInfo.description),
			FileName: node.GetPosition().GetFile(),
			filePos:  node.GetPosition().AsIntString(),
			Diff:     node.GetPosition().GetLine(),
			Expected: schemaErrorInfo.expected,
			Found:    schemaErrorInfo.found,
			Hints:    schemaErrorInfo.hints,
		}
	}

	return &schemaError{
		Title:    "Schema Error",
		FileName: node.GetPosition().GetFile(),
		filePos:  node.GetPosition().AsIntString(),
		Diff:     node.GetPosition().GetLine(),
	}
}

func NewMismatchedTypeError(foundType yamlmeta.TypeWithValues, expectedType yamlmeta.Type) error {
	return &mismatchedTypeError{
		Found:    foundType,
		Expected: expectedType,
	}
}

func NewUnexpectedKeyError(found *yamlmeta.MapItem, definition *filepos.Position) error {
	return &unexpectedKeyError{
		Found:                 found,
		MapDefinitionPosition: definition,
	}
}

type schemaAssertionError struct {
	error
	description string
	expected    string
	found       string
	hints       []string
}

type schemaError struct {
	Title    string
	FileName string
	Diff     string
	Expected string
	Found    string
	Hints    []string

	filePos string
}

func (e schemaError) Error() string {
	funcMap := template.FuncMap{
		"pad": func(delim string, includeLineNumber bool) string {
			padding := "  "
			if includeLineNumber {
				return padding + e.filePos + " " + delim
			}
			return padding + strings.Repeat(" ", len(e.filePos)) + " " + delim
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

type mismatchedTypeError struct {
	Found    yamlmeta.TypeWithValues
	Expected yamlmeta.Type
}

func (t mismatchedTypeError) Error() string {
	position := t.Found.GetPosition().AsCompactString()
	lineContent := t.Found.GetPosition().GetLine()

	leftPadLength := len(position) + 1
	msg := "\n"
	msg += formatLine(leftPadLength, position, lineContent)
	msg += formatLine(leftPadLength, "", "")
	msg += formatLine(leftPadLength, "", "TYPE MISMATCH - the value of this item is not what schema expected:")
	msg += formatLine(leftPadLength, "", fmt.Sprintf("     found: %s", t.Found.ValueTypeAsString()))

	if t.Expected.PositionOfDefinition().IsKnown() {
		expectedTypeString := ""
		switch t.Expected.(type) {
		case *MapItemType, *ArrayItemType:
			expectedTypeString = t.Expected.GetValueType().String()
		default:
			expectedTypeString = t.Expected.String()
		}

		msg += formatLine(leftPadLength, "", fmt.Sprintf("  expected: %s (by %s)", expectedTypeString, t.Expected.PositionOfDefinition().AsCompactString()))
	}

	return msg
}

type unexpectedKeyError struct {
	Found                 *yamlmeta.MapItem
	MapDefinitionPosition *filepos.Position
}

func (t unexpectedKeyError) Error() string {
	position := t.Found.Position.AsCompactString()
	leftColumnSize := len(position) + 1
	lineContent := strings.TrimSpace(t.Found.Position.GetLine())
	keyAsString := fmt.Sprintf("%v", t.Found.Key)

	msg := "\n"
	msg += formatLine(leftColumnSize, position, lineContent)
	msg += formatLine(leftColumnSize, "", "")
	msg += formatLine(leftColumnSize, "", "UNEXPECTED KEY - the key of this item was not found in the schema's corresponding map:")
	msg += formatLine(leftColumnSize, "", fmt.Sprintf("     found: %s", keyAsString))
	msg += formatLine(leftColumnSize, "", fmt.Sprintf("  expected: (a key defined in map) (by %s)", t.MapDefinitionPosition.AsCompactString()))
	msg += formatLine(leftColumnSize, "", "  (hint: declare data values in schema and override them in a data values document)")
	return msg
}

func leftPadding(size int) string {
	result := ""
	for i := 0; i < size; i++ {
		result += " "
	}
	return result
}

func formatLine(leftColumnSize int, left, right string) string {
	if len(right) > 0 {
		right = " " + right
	}
	return fmt.Sprintf("%s%s|%s\n", left, leftPadding(leftColumnSize-len(left)), right)
}
