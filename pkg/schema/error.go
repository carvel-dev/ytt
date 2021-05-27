// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func NewInvalidSchemaError(found yamlmeta.Node, message, hint string) error {
	return &invalidSchemaError{
		Message: message,
		Found:   found,
		Hint:    hint,
	}
}

func NewInvalidArrayDefinitionError(found yamlmeta.Node, hint string) error {
	return &invalidArrayDefinitionError{
		Found: found,
		Hint:  hint,
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

type invalidSchemaError struct {
	Message string
	Found   yamlmeta.Node
	Hint    string
}

func (e invalidSchemaError) Error() string {
	position := e.Found.GetPosition().AsCompactString()
	leftColumnSize := len(position) + 1
	lineContent := e.Found.GetPosition().GetLine()

	msg := "\n"
	msg += formatLine(leftColumnSize, position, lineContent)
	msg += formatLine(leftColumnSize, "", "")
	msg += formatLine(leftColumnSize, "", "INVALID SCHEMA - "+e.Message)
	if e.Hint != "" {
		msg += formatLine(leftColumnSize, "", fmt.Sprintf("  (hint: %s)", e.Hint))
	}

	return msg
}

type invalidArrayDefinitionError struct {
	Found yamlmeta.Node
	Hint  string
}

func (i invalidArrayDefinitionError) Error() string {
	position := i.Found.GetPosition().AsCompactString()
	leftColumnSize := len(position) + 1
	lineContent := i.Found.GetPosition().GetLine()

	msg := "\n"
	msg += formatLine(leftColumnSize, position, lineContent)
	msg += formatLine(leftColumnSize, "", "")
	msg += formatLine(leftColumnSize, "", "INVALID ARRAY DEFINITION IN SCHEMA - unable to determine the desired type")
	msg += formatLine(leftColumnSize, "", fmt.Sprintf("     found: %d array items", len(i.Found.GetValues())))
	msg += formatLine(leftColumnSize, "", "  expected: exactly 1 array item, of the desired type")
	msg += formatLine(leftColumnSize, "", fmt.Sprintf("  (hint: %s)", i.Hint))

	return msg
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
