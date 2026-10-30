// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"cmp"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"carvel.dev/ytt/pkg/files"
	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/yamlmeta"
)

type diagnosticValues []string

func diagnosticValuesFromFiles(inputs []*files.File) diagnosticValues {
	var result diagnosticValues
	for _, file := range inputs {
		result = append(result, diagnosticFileStrings(file)...)
	}
	return result
}

func diagnosticFileStrings(file *files.File) []string {
	bs, err := file.Bytes()
	if err != nil {
		return nil
	}
	content := string(bs)
	if !strings.Contains(content, "#@data/values") ||
		strings.Contains(content, "#@data/values-schema") {
		return nil
	}
	docs, err := yamlmeta.NewDocumentSetFromBytes(bs, yamlmeta.DocSetOpts{})
	if err != nil {
		return nil
	}
	var result []string
	for _, doc := range docs.Items {
		result = append(result, diagnosticStrings(doc.Value)...)
	}
	return result
}

func (d *diagnosticValues) add(envelopes ...*datavalues.Envelope) {
	for _, envelope := range envelopes {
		*d = append(*d, diagnosticStrings(envelope.Doc.Value)...)
	}
}

func (d *diagnosticValues) redactOutput(out *Output) {
	if out.Err != nil {
		out.Err = redactDataValuesError(out.Err, *d)
	}
}

// Input values are opaque; a key name cannot prove that a value is safe.
func diagnosticStrings(value any) []string {
	var result []string
	switch value := value.(type) {
	case *yamlmeta.Map:
		for _, item := range value.Items {
			result = append(result, diagnosticStrings(item.Value)...)
		}
	case *yamlmeta.Array:
		for _, item := range value.Items {
			result = append(result, diagnosticStrings(item.Value)...)
		}
	case string:
		result = append(result, value)
	default:
		// Scalar non-string values cannot contain string credentials.
	}
	return result
}

func diagnosticPatterns(values []string) []string {
	variants := map[string]bool{}
	for _, value := range values {
		if value == "" {
			continue
		}
		variants[value] = true
		quoted := strconv.Quote(value)
		variants[quoted[1:len(quoted)-1]] = true
	}
	var alternatives []string
	for value := range variants {
		alternatives = append(alternatives, value)
	}
	slices.SortFunc(alternatives, compareDiagnosticValues)
	for i, value := range alternatives {
		alternatives[i] = regexp.QuoteMeta(value)
	}
	return alternatives
}

func compareDiagnosticValues(a, b string) int {
	// Prefer a complete value over an overlapping substring.
	if len(a) != len(b) {
		return cmp.Compare(len(b), len(a))
	}
	return strings.Compare(a, b)
}

func redactDataValuesError(err error, values []string) error {
	alternatives := diagnosticPatterns(values)
	if len(alternatives) == 0 {
		return err
	}
	matcher, compileErr := regexp.Compile(strings.Join(alternatives, "|"))
	if compileErr != nil {
		return errors.New("Template processing failed; " +
			"check data value types " +
			"and template expressions (input values omitted)")
	}
	// Replace values even when a template embeds them in a larger message.
	// Redaction never changes the values supplied to evaluation itself.
	safe := matcher.ReplaceAllLiteralString(err.Error(), "[REDACTED]")
	return errors.New(safe)
}
