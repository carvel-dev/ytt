// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"

	"syscall/js"
)

type jsFunc func(js.Value, []js.Value) interface{}

func registerFunc(name string, fn jsFunc) {
	js.Global().Set(name, js.FuncOf(fn))
	fmt.Printf("Registered \"%s\" with Global.\n", name)
}

func template(this js.Value, args []js.Value) interface{} {
	yamlTplData := []byte(`
#@ load("@ytt:data", "data")
data_int: #@ data.values.int
data_str: #@ data.values.str`)

	yamlData := []byte(`
#@data/values
---
int: 123
str: str`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("values/data.yml", yamlData)),
	})

	tty := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, tty)
	file := out.Files[0]

	return string(file.Bytes())
}

func main() {
	registerFunc("ytt", template)

	// Go-based WASM modules must remain running to be available to the runtime.
	<-make(chan int)
}
