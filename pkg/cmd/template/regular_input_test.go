// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Non_YAML_Files_With_No_Output_Flag_Produces_Warning(t *testing.T) {
	iniData := []byte(`
[owner]
name=John Doe
organization=Acme Widgets Inc.`)
	yamlData := []byte(`foo: bar`)

	expectedStdErr := fmt.Sprintf("\n" + `Warning: Found Non-YAML templates in input. Non-YAML templates are not rendered to standard output.
If you want to include those results, use the --output-files or --dangerous-emptied-output-directory flag.
`)
	expectedStdOut := "foo: bar\n"

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("ini.txt", iniData)),
		files.MustNewFileFromSource(files.NewBytesSource("foo.yaml", yamlData)),
	}

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	rfsOpts := cmdtpl.RegularFilesSourceOpts{OutputType: cmdtpl.OutputType{Types: []string{"yaml"}}}
	rfs := cmdtpl.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)

	err := rfs.Output(out)
	require.NoError(t, err)

	assertStdoutAndStderr(t, stdout, stderr, expectedStdOut, expectedStdErr)
}

func Test_Non_YAML_With_Output_Flag_Shows_No_Warning(t *testing.T) {
	iniData := []byte(`
[owner]
name=John Doe
organization=Acme Widgets Inc.`)
	yamlData := []byte(`foo: bar`)

	expectedStdErr := ""

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("ini.txt", iniData)),
		files.MustNewFileFromSource(files.NewBytesSource("foo.yaml", yamlData)),
	}

	outputDir, err := os.MkdirTemp(os.TempDir(), "fakedir")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	expectedStdOut := `creating: ` + outputDir + "/ini.txt" + "\n" + `creating: ` + outputDir + "/foo.yaml" + "\n"

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	rfsOpts := cmdtpl.RegularFilesSourceOpts{OutputType: cmdtpl.OutputType{Types: []string{"yaml"}}, OutputFiles: outputDir}
	rfs := cmdtpl.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)

	err = rfs.Output(out)
	require.NoError(t, err)

	assertStdoutAndStderr(t, stdout, stderr, expectedStdOut, expectedStdErr)
}

func Test_Input_Marked_As_Starlark_Is_Evaluated_As_Starlark(t *testing.T) {
	starlarkData := []byte(`fail("hello")`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("print.star1", starlarkData)),
	}

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	testUI := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	opts.FileMarksOpts.FileMarks = []string{"print.star1:type=starlark"}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, testUI)
	require.Error(t, out.Err)
	assert.ErrorContains(t, out.Err, "fail: hello")
}

func Test_FileMark_YAML_Shows_No_Warning(t *testing.T) {
	yamlData := []byte(`foo: bar`)

	expectedStdErr := ""
	expectedStdOut := "foo: bar\n"

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("yaml.txt", yamlData)),
	}

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	opts.FileMarksOpts.FileMarks = []string{"yaml.txt:type=yaml-plain"}
	rfsOpts := cmdtpl.RegularFilesSourceOpts{OutputType: cmdtpl.OutputType{Types: []string{"yaml"}}}
	rfs := cmdtpl.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)

	err := rfs.Output(out)
	require.NoError(t, err)

	assertStdoutAndStderr(t, stdout, stderr, expectedStdOut, expectedStdErr)
}

func Test_OutputType_Flag(t *testing.T) {
	type example struct {
		desc   string
		input  []string
		format interface{}
		schema interface{}
	}
	successExamples := []example{
		{
			desc:   "no_input",
			input:  []string{},
			format: "yaml",
			schema: "",
		},
		{
			desc:   "explicitly_YAML",
			input:  []string{"yaml"},
			format: "yaml",
			schema: "",
		},
		{
			desc:   "explicitly_JSON",
			input:  []string{"json"},
			format: "json",
			schema: "",
		},
		{
			desc:   "explicitly_POS",
			input:  []string{"pos"},
			format: "pos",
			schema: "",
		},
		{
			desc:   "explicitly_YAML,_OpenAPI v3",
			input:  []string{"yaml", "openapi-v3"},
			format: "yaml",
			schema: "openapi-v3",
		},
		{
			desc:   "explicitly_JSON,_OpenAPI_v3",
			input:  []string{"json", "openapi-v3"},
			format: "json",
			schema: "openapi-v3",
		},
		{
			desc:   "explicitly_POS,_OpenAPI_v3",
			input:  []string{"pos", "openapi-v3"},
			format: "pos",
			schema: "openapi-v3",
		},
	}
	for _, eg := range successExamples {
		t.Run(eg.desc, func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.RegularFilesSourceOpts.OutputType.Types = eg.input

			format, formatErr := opts.RegularFilesSourceOpts.OutputType.Format()
			schema, schemaErr := opts.RegularFilesSourceOpts.OutputType.Schema()

			t.Run("format", func(t *testing.T) {
				require.NoError(t, formatErr)
				require.Equal(t, eg.format, format)
			})
			t.Run("schema", func(t *testing.T) {
				require.NoError(t, schemaErr)
				require.Equal(t, eg.schema, schema)
			})
		})
	}

	errorExamples := []example{
		{
			desc:   "invalid_input",
			input:  []string{"ymal"},
			format: errors.New("Unknown output type 'ymal'"),
			schema: errors.New("Unknown output type 'ymal'"),
		},
		{
			desc:   "empty_input",
			input:  []string{""},
			format: errors.New("Unknown output type ''"),
			schema: errors.New("Unknown output type ''"),
		},
	}
	for _, eg := range errorExamples {
		t.Run(eg.desc, func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.RegularFilesSourceOpts.OutputType.Types = eg.input

			_, formatErr := opts.RegularFilesSourceOpts.OutputType.Format()
			_, schemaErr := opts.RegularFilesSourceOpts.OutputType.Schema()

			t.Run("format", func(t *testing.T) {
				require.Error(t, formatErr)
				require.Equal(t, eg.format, formatErr)
			})
			t.Run("schema", func(t *testing.T) {
				require.Error(t, schemaErr)
				require.Equal(t, eg.schema, schemaErr)
			})
		})
	}
}

func TestFileMarkMultipleExcludes(t *testing.T) {
	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("a.yml", []byte(`a: bar`))),
		files.MustNewFileFromSource(files.NewBytesSource("b.yml", []byte(`b: bar`))),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", []byte(`config: bar`))),
		files.MustNewFileFromSource(files.NewBytesSource("d1.yml", []byte(`d: bar`))),
	}

	opts := cmdtpl.NewOptions()
	opts.FileMarksOpts.FileMarks = []string{"a.yml:exclude=true", "b.yml:exclude=true", "d*.yml:exclude=true"}

	runAndCompareWithOpts(t, opts, filesToProcess, "config: bar\n")
}

func assertStdoutAndStderr(t *testing.T, stdout *bytes.Buffer, stderr *bytes.Buffer, expectedStdOut string, expectedStdErr string) {
	stdoutOutput, err := io.ReadAll(stdout)
	require.NoError(t, err, "reading stdout")

	require.Equal(t, expectedStdOut, string(stdoutOutput), "comparing stdout")

	stderrOutput, err := io.ReadAll(stderr)
	require.NoError(t, err, "reading stdout")

	require.Equal(t, expectedStdErr, string(stderrOutput), "comparing stderr")
	return
}
