// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"sync"
	"testing"

	cmdtpl "carvel.dev/ytt/pkg/cmd/template"
	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The library deliberately uses the Starlark features that depend on the
// resolver flags this package enables (lambda, float, bitwise ops, recursion),
// so that dropping any of those flags fails this test rather than passing
// silently.
var concurrencyLibData = []byte(`
def factorial(n):
  return 1 if n <= 1 else n * factorial(n - 1)
end

def build():
  double = lambda x: x * 2
  return {
    "total": double(6),
    "ratio": 1.5,
    "masked": 12 & 10,
    "shifted": 1 << 4,
    "factorial": factorial(5),
  }
end
`)

var concurrencyTplData = []byte(`
#@ load("lib.star", "build")
#@ v = build()
values:
  total: #@ v["total"]
  ratio: #@ v["ratio"]
  masked: #@ v["masked"]
  shifted: #@ v["shifted"]
  factorial: #@ v["factorial"]
`)

const concurrencyExpected = `values:
  total: 12
  ratio: 1.5
  masked: 8
  shifted: 16
  factorial: 120
`

// TestConcurrentRendersProduceIdenticalOutput renders the same input from many
// goroutines at once. Embedders using ytt as a Go module render concurrently
// (one render per request, or per reconcile loop), so RunWithFiles has to be
// safe to call from multiple goroutines. Two pieces of package-level state used
// to make it unsafe: the instruction set ID counter in pkg/template, and the
// starlark-go resolver flags that were assigned on every compile.
//
// This should be run with -race to catch the races directly; the output
// comparison catches the functional consequence of an ID collision between a
// template and the library it loads.
func TestConcurrentRendersProduceIdenticalOutput(t *testing.T) {
	// A serial render first, so a failure in the concurrent phase is
	// unambiguously about concurrency rather than about the template.
	baseline, err := renderWithLib()
	require.NoError(t, err)
	require.Equal(t, concurrencyExpected, baseline)

	const goroutines = 16

	// Every goroutine gets its own output and error, so that we can report
	// which one failed. Each one is separate so that we don't need to lock on
	// anything
	outs := make([]string, goroutines)
	errs := make([]error, goroutines)

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Go(func() {
			outs[i], errs[i] = renderWithLib()
		})
	}
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "concurrent render %d failed", i)
		assert.Equal(t, concurrencyExpected, outs[i],
			"concurrent render %d diverged from the serial render", i)
	}
}

// renderWithLib evaluates the template with the library available to load(),
// returning the rendered YAML. Every call builds its own files and Options, as
// each render is expected to do.
func renderWithLib() (string, error) {
	lib := files.NewBytesSource("lib.star", concurrencyLibData)
	tpl := files.NewBytesSource("tpl.yaml", concurrencyTplData)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(lib),
		files.MustNewFileFromSource(tpl),
	}

	out := cmdtpl.NewOptions().RunWithFiles(
		cmdtpl.Input{Files: filesToProcess},
		ui.NewTTY(false),
	)
	if out.Err != nil {
		return "", out.Err
	}

	bs, err := out.DocSet.AsBytes()
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
