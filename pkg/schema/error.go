// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"
	"text/template"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/spell"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const schemaErrorReportTemplate = `
{{- if .Summary}}
{{.Summary}}
{{addBreak .Summary}}
{{ end}}
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
{{- range .AnnPositions}}
{{pad "|" .Pos}} {{.Source}}
{{- if .SkipLines}}
{{pad "|" ""}} {{"..."}}
{{- end}}
{{- end}}
{{pad "|" .FilePos}} {{.Source}}
{{pad "|" ""}}
{{- end}}

{{with .Found}}{{pad "=" ""}} found: {{.}}{{end}}
{{with .Expected}}{{pad "=" ""}} expected: {{.}}{{end}}
{{- range .Hints}}
{{pad "=" ""}} hint: {{.}}
{{- end}}
{{end}}
{{- .MiscErrorMessage}}
`

func NewSchemaError(summary string, errs ...error) error {
	var failures []assertionFailure
	var miscErrorMessage string
	for _, err := range errs {
		if typeCheckAssertionErr, ok := err.(schemaAssertionError); ok {
			failures = append(failures, assertionFailure{
				Description:  typeCheckAssertionErr.description,
				FileName:     typeCheckAssertionErr.position.GetFile(),
				AnnPositions: createAnnInfo(typeCheckAssertionErr.annPositions, typeCheckAssertionErr.position),
				FilePos:      typeCheckAssertionErr.position.AsIntString(),
				FromMemory:   typeCheckAssertionErr.position.FromMemory(),
				SourceName:   "Data value calculated",
				Source:       typeCheckAssertionErr.position.GetLine(),
				Expected:     typeCheckAssertionErr.expected,
				Found:        typeCheckAssertionErr.found,
				Hints:        typeCheckAssertionErr.hints,
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
		// TODO: remove this hint once we can report if mistyped value came from annotation
		hints: []string{fmt.Sprintf("is the default value set using @%v?", AnnotationDefault)},
	}
}

// NewUnexpectedKeyAssertionError generates a schema assertion error including the context (and hints) needed to report it to the user
func NewUnexpectedKeyAssertionError(found *yamlmeta.MapItem, definition *filepos.Position, allowedKeys []string) error {
	key := fmt.Sprintf("%v", found.Key)
	err := schemaAssertionError{
		description: "Given data value is not declared in schema",
		position:    found.GetPosition(),
		found:       key,
	}
	sort.Strings(allowedKeys)
	switch numKeys := len(allowedKeys); {
	case numKeys == 1:
		err.expected = fmt.Sprintf(`a %s with the key named "%s" (from %s)`, found.DisplayName(), allowedKeys[0], definition.AsCompactString())
	case numKeys > 1 && numKeys <= 9: // Miller's Law
		err.expected = fmt.Sprintf("one of { %s } (from %s)", strings.Join(allowedKeys, ", "), definition.AsCompactString())
	default:
		err.expected = fmt.Sprintf("a key declared in map (from %s)", definition.AsCompactString())
	}
	mostSimilarKey := spell.Nearest(key, allowedKeys)
	if mostSimilarKey != "" {
		err.hints = append(err.hints, fmt.Sprintf(`did you mean "%s"?`, mostSimilarKey))
	}
	return err
}

type schemaError struct {
	Summary           string
	AssertionFailures []assertionFailure
	MiscErrorMessage  string
}

type assertionFailure struct {
	Description  string
	FileName     string
	AnnPositions []annErr
	Source       string
	FilePos      string
	FromMemory   bool
	SourceName   string
	Expected     string
	Found        string
	Hints        []string
}

type schemaAssertionError struct {
	error
	annPositions []*filepos.Position
	position     *filepos.Position
	description  string
	expected     string
	found        string
	hints        []string
}

type annErr struct {
	Pos       string
	Source    string
	SkipLines bool
}

func createAnnInfo(annPosList []*filepos.Position, nodePos *filepos.Position) []annErr {
	sort.SliceStable(annPosList, func(i, j int) bool {
		return annPosList[i].LineNum() < annPosList[j].LineNum()
	})

	var AnnPositions []annErr
	for i, p := range annPosList {
		//if last position in list, compare with node position
		if i+1 == len(annPosList) {
			skipLines := !p.IsNextTo(nodePos)
			AnnPositions = append(AnnPositions, annErr{Pos: p.AsIntString(), Source: p.GetLine(), SkipLines: skipLines})
		} else {
			skipLines := !p.IsNextTo(annPosList[i+1])
			AnnPositions = append(AnnPositions, annErr{Pos: p.AsIntString(), Source: p.GetLine(), SkipLines: skipLines})
		}
	}
	return AnnPositions
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
