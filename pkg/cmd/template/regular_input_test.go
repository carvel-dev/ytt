// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/k14s/ytt/pkg/cmd/template"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
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
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml"}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

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

	outputDir, err := ioutil.TempDir(os.TempDir(), "fakedir")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	expectedStdOut := `creating: ` + outputDir + "/ini.txt" + "\n" + `creating: ` + outputDir + "/foo.yaml" + "\n"

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml", OutputFiles: outputDir}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)

	err = rfs.Output(out)
	require.NoError(t, err)

	assertStdoutAndStderr(t, stdout, stderr, expectedStdOut, expectedStdErr)
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
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml"}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	require.NoError(t, out.Err)

	err := rfs.Output(out)
	require.NoError(t, err)

	assertStdoutAndStderr(t, stdout, stderr, expectedStdOut, expectedStdErr)
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
	stdoutOutput, err := ioutil.ReadAll(stdout)
	require.NoError(t, err, "reading stdout")

	require.Equal(t, expectedStdOut, string(stdoutOutput), "comparing stdout")

	stderrOutput, err := ioutil.ReadAll(stderr)
	require.NoError(t, err, "reading stdout")

	require.Equal(t, expectedStdErr, string(stderrOutput), "comparing stderr")
	return
}
