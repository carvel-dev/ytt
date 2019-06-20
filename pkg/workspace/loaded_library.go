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
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type LoadedLibrary struct {
	*Library
	UI     files.UI
	Opts   eval.TemplateLoaderOpts
	Values interface{}
}

func LoadLibrary(lib *Library, o eval.TemplateLoaderOpts, parentAstValues interface{}) (load *LoadedLibrary, err error) {
	load = &LoadedLibrary{
		Opts:    o,
		Library: lib,
		Values:  parentAstValues,
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

func (l *LoadedLibrary) filterValues(astValues interface{}) error {
	if l.Values == nil {
		l.Values = astValues
		return nil
	}

	op := yttoverlay.OverlayOp{
		Left:   yamlmeta.NewASTFromInterface(astValues),
		Right:  l.Values,
		Thread: &starlark.Thread{Name: "data-values-overlay-pre-processing"},
	}

	newLeft, err := op.Apply()
	if err != nil {
		return fmt.Errorf("Overlaying data values from eval() provided values on top of data values (marked as @data/values): %s", err)
	}

	l.Values = (&yamlmeta.Document{Value: newLeft}).AsInterface(yamlmeta.InterfaceConvertOpts{})
	return nil
}

func (l *LoadedLibrary) loadValues(lib *Library, recurse bool) (values interface{}, valuesFile *files.File, err error) {
	for _, file := range lib.files {
		if file.Type() == files.TypeYAML && file.IsTemplate() {
			fileBs, err := file.Bytes()
			if err != nil {
				return nil, nil, err
			}

			docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, file.RelativePath())
			if err != nil {
				return nil, nil, fmt.Errorf("Unmarshaling YAML template '%s': %s", file.RelativePath(), err)
			}

			tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: l.Opts.IgnoreUnknownComments}

			vals, found, err := yttlibrary.DataValues{docSet, tplOpts}.Find()
			if err != nil {
				return nil, nil, err
			}

			if found {
				if valuesFile != nil {
					// TODO until overlays are here
					return nil, nil, fmt.Errorf(
						"Template values could only be specified once, but found multiple (%s, %s)",
						valuesFile.RelativePath(), file.RelativePath())
				}
				file.MarkForOutput(false)
				values = vals
				valuesFile = file
			}
		}
	}
	if recurse {
		for _, subLib := range lib.children {
			if !subLib.private {
				vals, file, err := l.loadValues(subLib, recurse)
				if err != nil {
					return nil, nil, err
				} else if file != nil {
					if valuesFile != nil {
						// TODO until overlays are here
						return nil, nil, fmt.Errorf(
							"Template values could only be specified once, but found multiple (%s, %s)",
							valuesFile.RelativePath(), file.RelativePath())
					}
					values = vals
					valuesFile = file
				}
			}
		}
	}
	if values == nil {
		values = map[interface{}]interface{}{}
	}
	return

}

func (l *LoadedLibrary) evalLibrary(lib *Library, res *eval.Result, overlayProcess *OverlayPostProcessing) error {
	loader := NewTemplateLoader(l.Values, l.UI, l.Opts)

	for _, file := range lib.files {
		fileInLib := FileInLibrary{file, lib}
		if !file.IsForOutput() {
			continue
		}

		// TODO find more generic way
		switch fileInLib.File.Type() {
		case files.TypeYAML:
			_, resultVal, err := loader.EvalYAML(fileInLib.Library, fileInLib.File)
			if err != nil {
				return err
			}

			resultDocSet := resultVal.(*yamlmeta.DocumentSet)
			overlayProcess.docSets[fileInLib.RelativePath(l.Library)] = resultDocSet

		case files.TypeText:
			_, resultVal, err := loader.EvalText(fileInLib.Library, fileInLib.File)
			if err != nil {
				return err
			}

			resultStr := resultVal.(*texttemplate.NodeRoot).AsString()

			l.UI.Debugf("### %s result\n%s", fileInLib.RelativePath(l.Library), resultStr)
			res.OutputFiles = append(res.OutputFiles, files.NewOutputFile(fileInLib.RelativePath(l.Library), []byte(resultStr)))

		default:
			return fmt.Errorf("Unknown file type %v for %s", file.Type(), fileInLib.RelativePath(l.Library))
		}
	}
	for _, subLib := range lib.children {
		if !subLib.private {
			err := l.evalLibrary(subLib, res, overlayProcess)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *LoadedLibrary) Eval() (*eval.Result, error) {
	vals, _, err := l.loadValues(l.Library, true)
	if err != nil {
		return nil, err
	}

	err = l.filterValues(vals)
	if err != nil {
		return nil, err
	}

	res := &eval.Result{
		DocSet: &yamlmeta.DocumentSet{},
	}
	overlayProcess := OverlayPostProcessing{
		docSets: map[string]*yamlmeta.DocumentSet{},
	}

	err = l.evalLibrary(l.Library, res, &overlayProcess)
	if err != nil {
		return nil, err
	}

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
