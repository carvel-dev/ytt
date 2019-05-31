package workspace

import (
	"fmt"
	"sort"

	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/texttemplate"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

type LoadedLibrary struct {
	*Library
	forOutputFiles []FileInLibrary
	valuesFile     *files.File
	UI             files.UI
	Opts           eval.TemplateLoaderOpts
	Values         interface{}
}

func LoadLibrary(lib *Library, o eval.TemplateLoaderOpts) (load *LoadedLibrary, err error) {
	load = &LoadedLibrary{
		Opts:    o,
		Library: lib,
	}

	for _, file := range lib.ListAccessibleFiles() {
		if file.File.Type() == files.TypeYAML && file.File.IsTemplate() {
			fileBs, err := file.File.Bytes()
			if err != nil {
				return nil, err
			}

			docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, file.File.RelativePath())
			if err != nil {
				return nil, fmt.Errorf("Unmarshaling YAML template '%s': %s", file.File.RelativePath(), err)
			}

			tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: o.IgnoreUnknownComments}

			values, found, err := yttlibrary.DataValues{docSet, tplOpts}.Find()
			if err != nil {
				return nil, err
			}

			if found {
				if load.valuesFile != nil {
					// TODO until overlays are here
					return nil, fmt.Errorf(
						"Template values could only be specified once, but found multiple (%s, %s)",
						load.valuesFile.RelativePath(), file.File.RelativePath())
				}
				load.valuesFile = file.File
				load.Values = values
				continue
			}
		}
		if file.File.IsForOutput() {
			load.forOutputFiles = append(load.forOutputFiles, file)
		}
	}

	if load.Values == nil {
		load.Values = map[interface{}]interface{}{}
	}

	return load, nil
}

func sortedOutputDocSetPaths(outputDocSets map[string]*yamlmeta.DocumentSet) []string {
	var paths []string
	for relPath, _ := range outputDocSets {
		paths = append(paths, relPath)
	}
	sort.Strings(paths)
	return paths
}

func (l *LoadedLibrary) FilterValues(filter eval.ValuesFilter) error {
	val, err := filter(l.Values)
	if err != nil {
		return err
	}

	l.Values = val
	return nil
}

func (l *LoadedLibrary) Eval() (*eval.Result, error) {
	loader := NewTemplateLoader(l.Values, l.UI, l.Opts)
	res := &eval.Result{
		DocSet: &yamlmeta.DocumentSet{},
	}
	overlayProcess := OverlayPostProcessing{
		docSets: map[string]*yamlmeta.DocumentSet{},
	}

	for _, file := range l.forOutputFiles {
		// TODO find more generic way
		switch file.File.Type() {
		case files.TypeYAML:
			_, resultVal, err := loader.EvalYAML(file.Library, file.File)
			if err != nil {
				return nil, err
			}

			resultDocSet := resultVal.(*yamlmeta.DocumentSet)
			overlayProcess.docSets[file.RelativePath(l.Library)] = resultDocSet

		case files.TypeText:
			_, resultVal, err := loader.EvalText(file.Library, file.File)
			if err != nil {
				return nil, err
			}

			resultStr := resultVal.(*texttemplate.NodeRoot).AsString()

			l.UI.Debugf("### %s result\n%s", file.RelativePath(l.Library), resultStr)
			res.OutputFiles = append(res.OutputFiles, files.NewOutputFile(file.RelativePath(l.Library), []byte(resultStr)))

		default:
			return nil, fmt.Errorf("Unknown file type")
		}
	}

	var err error
	res.DocSets, err = overlayProcess.Apply()
	if err != nil {
		return nil, err
	}

	for _, relPath := range sortedOutputDocSetPaths(res.DocSets) {
		docSet := res.DocSets[relPath]
		res.DocSet.Items = append(res.DocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return nil, fmt.Errorf("Marshaling template result: %s", err)
		}

		l.UI.Debugf("### %s result\n%s", relPath, resultDocBytes)
		res.OutputFiles = append(res.OutputFiles, files.NewOutputFile(relPath, resultDocBytes))
	}

	return res, nil
}
