// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	"carvel.dev/ytt/pkg/files"
)

func TestLoadAbs(t *testing.T) {
	configTplData := []byte(`
#@ load("/config.lib.yml", "func1")
#@ load("/dir/config.lib.yml", "func2")
#@ load("dir/config.lib.yml", func3="func2")
func1: #@ func1()
func2: #@ func2()
func3: #@ func3()`)

	configLibData := []byte(`
#@ def func1():
func1: true
#@ end`)

	dirConfigLibData := []byte(`
#@ load("/config.lib.yml", "func1")
#@ def func2():
func2: #@ func1()
#@ end`)

	expectedYAMLTplData := `func1:
  func1: true
func2:
  func2:
    func1: true
func3:
  func2:
    func1: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("config.lib.yml", configLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("dir/config.lib.yml", dirConfigLibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}

func TestLoadAbsWithinYttLibDirectory(t *testing.T) {
	configTplData := []byte(`
#@ load("/config.lib.yml", "func1")
#@ load("@dir:dir2/config.lib.yml", "func3")
func1: #@ func1()
func3: #@ func3()`)

	configLibData := []byte(`
#@ load("@dir:dir2/config.lib.yml", "func3")
#@ def func1():
func3: #@ func3()
#@ end`)

	dirConfigLibData := []byte(`
#@ def func2():
func2: true
#@ end`)

	dir2ConfigLibData := []byte(`
#@ load("/config.lib.yml", "func2")
#@ load("@dir3:/dir4/config.lib.yml", "func4")
#@ def func3():
func2: #@ func2()
func4: #@ func4()
#@ end
`)

	dir4ConfigLibData := []byte(`
#@ load("/funcs.lib.yml", "dir3_funcs")
#@ def func4():
dir3_funcs: #@ dir3_funcs()
#@ end
`)

	dir3FuncsLibData := []byte(`
#@ def dir3_funcs():
dir3_funcs: true
#@ end
`)

	expectedYAMLTplData := `func1:
  func3:
    func2:
      func2: true
    func4:
      dir3_funcs:
        dir3_funcs: true
func3:
  func2:
    func2: true
  func4:
    dir3_funcs:
      dir3_funcs: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("config.lib.yml", configLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/config.lib.yml", dirConfigLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/config.lib.yml", dir2ConfigLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/_ytt_lib/dir3/dir4/config.lib.yml", dir4ConfigLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/_ytt_lib/dir3/funcs.lib.yml", dir3FuncsLibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}
