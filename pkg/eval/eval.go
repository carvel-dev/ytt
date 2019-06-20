package eval

import (
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type ValuesFilter func(interface{}) (interface{}, error)

type Loader interface {
	LoadExternalLibrary(path []string, astValues interface{}, recursive bool, opts *TemplateLoaderOpts) (*Result, error)
	LoadInternalLibrary(path string, astValues interface{}, opts *TemplateLoaderOpts) (*Result, error)
}

type Result struct {
	DocSet      *yamlmeta.DocumentSet
	DocSets     map[string]*yamlmeta.DocumentSet
	OutputFiles []files.OutputFile
}

type TemplateLoaderOpts struct {
	IgnoreUnknownComments bool
}
