package eval

import (
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"go.starlark.net/starlark"
)

type Loader interface {
	LoadInternalLibrary(path string, astValues ...ValuesAst) (*Result, error)
}

// Result represents the output of an evaluation.
type Result struct {
	Files   []files.OutputFile
	DocSet  *yamlmeta.DocumentSet
	DocSets map[string]*yamlmeta.DocumentSet
	Exports []Export
}

type Export struct {
	RelativePath string
	Values       starlark.StringDict
}

// ValuesAst represents in AST form the data.values to be fed to the
// evaluation process.
type ValuesAst interface{}
