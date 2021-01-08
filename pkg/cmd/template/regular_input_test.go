// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
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

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()
	rfsOpts := template.RegularFilesSourceOpts{}
	rfsOpts.Set(&cobra.Command{})
	rfs := template.NewRegularFilesSource(rfsOpts, ui)

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("An unexpected error occurred: %v", out.Err)
	}

	fakeStdOut, fakeStdErr, err := redirectStdOutAndErr()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}
	defer os.Remove(fakeStdErr.Name())
	defer os.Remove(fakeStdOut.Name())
	defer restoreStdOutAndErr()

	err = rfs.Output(out)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %v", err)
	}
	restoreStdOutAndErr()

	err = assertStdOutAndStdErr(fakeStdOut, fakeStdErr, expectedStdOut, expectedStdErr)
	if err != nil {
		t.Fatalf("Assertion failed:\n%v", err)
	}
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
