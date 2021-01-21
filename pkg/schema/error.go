// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type errorWithContent struct {
	contents string
	Position *filepos.Position
}

func (e errorWithContent) Error() string {
	panic("do not use directly")
}

//contents?
func (e *errorWithContent) SetContext(doc *yamlmeta.Document) error {
	contents, err := doc.RawDataAtLine(e.Position)

	if err == nil {
		e.contents = contents
	}

	return err
}

func NewInvalidError(found, expected, hint string, position *filepos.Position) error {
	return &invalidError{
		errorWithContent: errorWithContent{
			Position: position,
		},
		Found:    found,
		Expected: expected,
		Hint:     hint,
	}
}

type invalidError struct {
	errorWithContent
	Found    string
	Expected string
	Message  string
	Hint     string
}

func (i invalidError) Error() string {
	var result string
	position := i.Position.AsCompactString()
	lineContent := strings.TrimSpace(i.contents)
	colonPosition := strings.Index(lineContent, ":")
	result += "\n"
	result += fmt.Sprintf("%s | %s", position, lineContent)
	result += beginningOfLine(len(position), leftPadding(colonPosition)+" ^^^")
	if i.Found != "" {
		result += beginningOfLine(len(position), fmt.Sprintf("found: %s", i.Found))
	}

	if i.Expected != "" {
		result += beginningOfLine(len(position), fmt.Sprintf("expected: %s", i.Expected))
	}

	if i.Message != "" {
		result += beginningOfLine(len(position), i.Message)
	}

	if i.Hint != "" {
		result += beginningOfLine(len(position), fmt.Sprintf("(hint: %s)", i.Hint))
	}
	return result
}

func NewTypeError(foundType, expectedType string, dataPosition, schemaPosition *filepos.Position) error {
	return &typeError{
		errorWithContent: errorWithContent{
			Position: dataPosition,
		},
		Found:          foundType,
		Expected:       expectedType,
		DataPosition:   dataPosition,
		SchemaPosition: schemaPosition,
	}
}

type typeError struct {
	errorWithContent
	Found          string
	Expected       string
	DataPosition   *filepos.Position
	SchemaPosition *filepos.Position
}

func (t typeError) Error() string {
	var result string
	position := t.DataPosition.AsCompactString()
	lineContent := strings.TrimSpace(t.contents)
	colonPosition := strings.Index(lineContent, ":")
	result += "\n"
	result += fmt.Sprintf("%s | %s", position, lineContent)
	result += beginningOfLine(len(position), leftPadding(colonPosition)+" ^^^")
	if t.Found != "" {
		result += beginningOfLine(len(position), fmt.Sprintf("found: %s", t.Found))
	}

	if t.Expected != "" {
		result += beginningOfLine(len(position),
			fmt.Sprintf("expected: %s (by %s)", t.Expected, t.SchemaPosition.AsCompactString()))
	}

	return result
}

func NewUnexpectedKeyError(keyName string, dataPosition, schemaPosition *filepos.Position) error {
	return &unexpectedKeyError{
		errorWithContent: errorWithContent{
			Position: dataPosition,
			contents: keyName,
		},
		DataPosition:   dataPosition,
		SchemaPosition: schemaPosition,
	}
}

type unexpectedKeyError struct {
	errorWithContent
	DataPosition   *filepos.Position
	SchemaPosition *filepos.Position
}

func (t unexpectedKeyError) Error() string {
	var result string
	position := t.DataPosition.AsCompactString()
	lineContent := strings.TrimSpace(t.contents)
	result += "\n"
	result += fmt.Sprintf("%s | %s", position, lineContent)
	result += "\n" + leftPadding(len(position)) + " | ^^^"
	result += beginningOfLine(len(position),
		fmt.Sprintf("unexpected key in map (as defined at %s)", t.SchemaPosition.AsCompactString()))

	return result
}

func beginningOfLine(size int, text string) string {
	return "\n" + leftPadding(size) + " |  " + text
}

func leftPadding(size int) string {
	result := ""
	for i := 0; i < size; i++ {
		result += " "
	}
	return result
}
