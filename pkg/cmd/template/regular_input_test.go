// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/cmd/template"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
	"github.com/spf13/cobra"
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
	rfsOpts := template.RegularFilesSourceOpts{}
	rfsOpts.Set(&cobra.Command{})
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("An unexpected error occurred: %v", out.Err)
	}

	err := rfs.Output(out)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}

	err = assertOnStdOutAndStdErr(stdout, stderr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n %s", err)
	}
}

func assertOnStdOutAndStdErr(stdout *bytes.Buffer, stderr *bytes.Buffer, expectedStdOut string, expectedStdErr string) error {
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

func redirectStdOutAndErr() (*os.File, *os.File, error) {
	fakeStdErr, err := ioutil.TempFile(os.TempDir(), "fakedev")
	if err != nil {
		return nil, nil, err
	}

	fakeStdOut, err := ioutil.TempFile(os.TempDir(), "fakedev")
	if err != nil {
		return nil, nil, err
	}

	os.Stdout = fakeStdOut
	os.Stderr = fakeStdErr

	return fakeStdOut, fakeStdErr, nil
}

func restoreStdOutAndErr() {
	os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
}

func assertStdOutAndStdErr(stdout, stderr *os.File, expectedOut, expectedErr string) error {
	stdErrContents, err := ioutil.ReadFile(stderr.Name())
	if err != nil {
		return err
	}

	if string(stdErrContents) != expectedErr {
		return fmt.Errorf("Expected std err to be >>>%s<<<\nBut was: >>>%s<<<", expectedErr, stdErrContents)
	}

	stdOutContents, err := ioutil.ReadFile(stdout.Name())
	if err != nil {
		return fmt.Errorf("Expected reading stdout to not fail: %v", err)
	}

	if string(stdOutContents) != expectedOut {
		return fmt.Errorf("Expected std out to be >>>%s<<<, but was: >>>%s<<<", expectedOut, stdOutContents)
	}
	return nil
}
