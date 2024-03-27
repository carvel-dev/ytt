// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"

	yttcmd "carvel.dev/ytt/pkg/cmd/template"
	yttui "carvel.dev/ytt/pkg/cmd/ui"
	yttfiles "carvel.dev/ytt/pkg/files"
)

func main() {
	run(os.Stdout)
}

func run(out io.Writer) {
	// inputs are YAML-formatted strings
	tpl := []string{`
#@ load("@ytt:data", "data")
---
greeting: #@ "Hello, " + data.values.name
`}
	// in the format "(data-value-name)=value"
	dvs := []string{"name=world"}

	results, err := ytt(tpl, dvs)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(out, "%s", results)
}

func ytt(tpl, dvs []string) (string, error) {
	// create and invoke ytt "template" command
	templatingOptions := yttcmd.NewOptions()

	input, err := templatesAsInput(tpl...)
	if err != nil {
		return "", err
	}

	// equivalent to `--data-value-yaml`
	templatingOptions.DataValuesFlags.KVsFromYAML = dvs

	// for in-memory use, pipe output to "/dev/null"
	noopUI := yttui.NewCustomWriterTTY(false, noopWriter{}, noopWriter{})

	// Evaluate the template given the configured data values...
	output := templatingOptions.RunWithFiles(input, noopUI)
	if output.Err != nil {
		return "", output.Err
	}

	// output.DocSet contains the full set of resulting YAML documents, in order.
	bs, err := output.DocSet.AsBytes()
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// templatesAsInput conveniently wraps one or more strings, each in a files.File, into a template.Input.
func templatesAsInput(tpl ...string) (yttcmd.Input, error) {
	var files []*yttfiles.File
	for i, t := range tpl {
		// to make this less brittle, you'll probably want to use well-defined names for `path`, here, for each input.
		// this matters when you're processing errors which report based on these paths.
		file, err := yttfiles.NewFileFromSource(yttfiles.NewBytesSource(fmt.Sprintf("tpl%d.yml", i), []byte(t)))
		if err != nil {
			return yttcmd.Input{}, err
		}

		files = append(files, file)
	}

	return yttcmd.Input{Files: files}, nil
}

type noopWriter struct{}

func (w noopWriter) Write(data []byte) (int, error) { return len(data), nil }
