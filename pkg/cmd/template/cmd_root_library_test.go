package template_test

import (
	"testing"

	"github.com/k14s/ytt/pkg/files"
)

func TestRootLibraryLoad(t *testing.T) {
	configTplData := []byte(`
#@ load("@:config.lib.yml", "func1")
#@ load("@:dir/config.lib.yml", "func2")
#@ load("dir/config.lib.yml", func3="func2")
func1: #@ func1()
func2: #@ func2()
func3: #@ func3()`)

	configLibData := []byte(`
#@ def func1():
func1: true
#@ end`)

	dirConfigLibData := []byte(`
#@ load("@:config.lib.yml", "func1")
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

func TestRootLibraryLoadWithinYttLibDirectory(t *testing.T) {
	configTplData := []byte(`
#@ load("@:config.lib.yml", "func1")
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
#@ load("@:config.lib.yml", "func2")
#@ def func3():
func3: #@ func2()
#@ end
`)

	expectedYAMLTplData := `func1:
  func3:
    func3:
      func2: true
func3:
  func3:
    func2: true
`

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", configTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("config.lib.yml", configLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/config.lib.yml", dirConfigLibData)),
		files.MustNewFileFromSource(files.NewBytesSource("_ytt_lib/dir/dir2/config.lib.yml", dir2ConfigLibData)),
	})

	runAndCompare(t, filesToProcess, expectedYAMLTplData)
}
