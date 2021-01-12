// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/cmd/template"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
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
	ui := cmdcore.NewFakeUI(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml"}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("An unexpected error occurred: %v", out.Err)
	}

	err := rfs.Output(out)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	err = assertStdoutAndStderr(stdout, stderr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n %s", err)
	}
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
	if err != nil {
		t.Fatalf("Expected creating a temp dir to not fail: %v", err)
	}
	defer os.RemoveAll(outputDir)

	expectedStdOut := `creating: ` + outputDir + "/ini.txt" + "\n" + `creating: ` + outputDir + "/foo.yaml" + "\n"

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := cmdcore.NewFakeUI(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml", OutputFiles: outputDir}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("An unexpected error occurred: %v", out.Err)
	}

	err = rfs.Output(out)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	err = assertStdoutAndStderr(stdout, stderr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n %s", err)
	}
}

func Test_FileMark_YAML_With_Shows_No_Warning(t *testing.T) {
	yamlData := []byte(`foo: bar`)

	expectedStdErr := ""
	expectedStdOut := "foo: bar\n"

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("yaml.txt", yamlData)),
	}

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	ui := cmdcore.NewFakeUI(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	opts.FileMarksOpts.FileMarks = []string{"yaml.txt:type=yaml-plain"}
	rfsOpts := template.RegularFilesSourceOpts{OutputType: "yaml"}
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("An unexpected error occurred: %v", out.Err)
	}

	err := rfs.Output(out)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	err = assertStdoutAndStderr(stdout, stderr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n %s", err)
	}
}

func assertStdoutAndStderr(stdout *bytes.Buffer, stderr *bytes.Buffer, expectedStdOut string, expectedStdErr string) error {
	stdoutOutput, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("Error reading stdout buffer: %s", err)
	}

	if string(stdoutOutput) != expectedStdOut {
		return fmt.Errorf("Expected stdout to be >>>%s<<<\nBut was: >>>%s<<<", expectedStdOut, string(stdoutOutput))
	}

	stderrOutput, err := ioutil.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("Error reading stderr buffer: %s", err)
	}

	if string(stderrOutput) != expectedStdErr {
		return fmt.Errorf("Expected stderr to be >>>%s<<<\nBut was: >>>%s<<<", expectedStdErr, string(stderrOutput))
	}
	return nil
}
