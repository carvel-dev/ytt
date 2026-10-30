// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

type diagnosticCase struct {
	name, template, schema string
	fails, rendersSecret   bool
}

func TestRuntimeDataValuesStayOutOfDiagnostics(t *testing.T) {
	templates := []diagnosticCase{
		{name: "schema", template: "value: safe\n",
			schema: "#@data/values-schema\n---\npassword: 0\n", fails: true},
		{name: "runtime", template: `#@ load("@ytt:data", "data")
value: #@ int(data.values.password)
`, fails: true},
		{name: "embedded runtime", template: `#@ load("@ytt:data", "data")
value: #@ int("prefix" + data.values.password + "suffix")
`, fails: true},
		{name: "debug", template: "value: safe\n"},
		{name: "output", template: `#@ load("@ytt:data", "data")
value: #@ data.values.password
`, rendersSecret: true},
	}
	for _, tc := range templates {
		t.Run(tc.name, func(t *testing.T) {
			runDiagnosticCase(t, tc)
		})
	}
}

func runDiagnosticCase(t *testing.T, tc diagnosticCase) {
	t.Helper()
	const secret = "password-canary"
	opts := cmdtpl.NewOptions()
	opts.DataValuesFlags.KVsFromStrings = []string{"password=" + secret}
	inputs := []*files.File{
		diagnosticInput("template.yml", tc.template),
	}
	if tc.schema != "" {
		schemaInput := diagnosticInput("schema.yml", tc.schema)
		inputs = append(inputs, schemaInput)
	}
	var stdout, stderr bytes.Buffer
	tty := ui.NewCustomWriterTTY(true, &stdout, &stderr)
	out := opts.RunWithFiles(cmdtpl.Input{
		Files: files.NewSortedFiles(inputs),
	}, tty)
	if tc.fails {
		require.Error(t, out.Err)
		require.NotContains(t, out.Err.Error(), secret)
	} else {
		require.NoError(t, out.Err)
	}
	if tc.rendersSecret {
		require.Contains(t, string(out.Files[0].Bytes()), secret)
	}
	require.NotContains(t, stderr.String(), secret)
}

func diagnosticInput(name, content string) *files.File {
	return files.MustNewFileFromSource(
		files.NewBytesSource(name, []byte(content)))
}
