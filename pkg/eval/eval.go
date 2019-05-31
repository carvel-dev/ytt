package eval

import (
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type ValuesFilter func(interface{}) (interface{}, error)

type Loader interface {
	LoadExternalLibrary(path []string, recursive bool, opts *TemplateLoaderOpts) (Library, error)
	LoadInternalLibrary(path string, opts *TemplateLoaderOpts) (Library, error)
}

type Library interface {
	FilterValues(filter ValuesFilter) error
	Eval() (*Result, error)
}

type Result struct {
	DocSet      *yamlmeta.DocumentSet
	DocSets     map[string]*yamlmeta.DocumentSet
	OutputFiles []files.OutputFile
}

type TemplateLoaderOpts struct {
	IgnoreUnknownComments bool
}
