// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
)

func NewInvalidSchemaError(found, expected, hint string, position *filepos.Position) error {
	return &invalidSchemaError{
		Position: position,
		Found:    found,
		Expected: expected,
		Hint:     hint,
	}
}

type invalidSchemaError struct {
	Position *filepos.Position
	Found    string
	Expected string
	Message  string
	Hint     string
}

func (i invalidSchemaError) Error() string {
	var result string
	position := i.Position.AsCompactString()
	lineContent := strings.TrimSpace(i.Position.GetLine())
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
		Found:          foundType,
		Expected:       expectedType,
		DataPosition:   dataPosition,
		SchemaPosition: schemaPosition,
	}
}

type typeError struct {
	Found          string
	Expected       string
	DataPosition   *filepos.Position
	SchemaPosition *filepos.Position
}

func (t typeError) Error() string {
	var result string
	position := t.DataPosition.AsCompactString()
	lineContent := strings.TrimSpace(t.DataPosition.GetLine())
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

func NewUnexpectedKeyError(dataPosition, schemaPosition *filepos.Position) error {
	return &unexpectedKeyError{
		DataPosition:   dataPosition,
		SchemaPosition: schemaPosition,
	}
}

type unexpectedKeyError struct {
	DataPosition   *filepos.Position
	SchemaPosition *filepos.Position
}

func (t unexpectedKeyError) Error() string {
	var result string
	position := t.DataPosition.AsCompactString()
	lineContent := strings.TrimSpace(t.DataPosition.GetLine())
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
