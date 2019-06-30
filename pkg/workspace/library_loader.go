package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/eval"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

type LibraryLoader struct {
	library            *Library
	ui                 files.UI
	templateLoaderOpts TemplateLoaderOpts
	libraryFiles       []*FileInLibrary
	evalFiles          []*FileInLibrary
}

func NewLibraryLoader(lib *Library, evalFiles []*FileInLibrary, ui files.UI, templateLoaderOpts TemplateLoaderOpts) *LibraryLoader {
	libraryFiles := lib.ListAccessibleFiles()
	if evalFiles == nil {
		evalFiles = libraryFiles
	}
	return &LibraryLoader{
		library:            lib,
		ui:                 ui,
		templateLoaderOpts: templateLoaderOpts,
		libraryFiles:       libraryFiles,
		evalFiles:          evalFiles,
	}
}

func (ll *LibraryLoader) Values(valuesFlagsAst ...eval.ValuesAst) (eval.ValuesAst, error) {
	loader := NewTemplateLoader(nil, ll.ui, ll.templateLoaderOpts)

	valuesFiles, err := ll.valuesFiles(loader)
	if err != nil {
		return nil, err
	}

	dvpp := DataValuesPreProcessing{
		valuesFiles:           valuesFiles,
		valuesFlagsAst:        valuesFlagsAst,
		loader:                loader,
		IgnoreUnknownComments: ll.templateLoaderOpts.IgnoreUnknownComments,
	}

	return dvpp.Apply()
}

func (ll *LibraryLoader) valuesFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	var valuesFiles []*FileInLibrary

	for _, fileInLib := range ll.libraryFiles {
		if fileInLib.File.Type() == files.TypeYAML && fileInLib.File.IsTemplate() {
			docSet, err := loader.ParseYAML(fileInLib.File)
			if err != nil {
				return nil, err
			}

			tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: ll.templateLoaderOpts.IgnoreUnknownComments}

			valuesDocs, _, err := yttlibrary.DataValues{docSet, tplOpts}.Extract()
			if err != nil {
				return nil, err
			}

			if len(valuesDocs) > 0 {
				valuesFiles = append(valuesFiles, fileInLib)
				fileInLib.File.MarkForOutput(false)
			}
		}
	}

	return valuesFiles, nil
}

func (ll *LibraryLoader) Eval(values eval.ValuesAst) (*eval.Result, error) {
	docSets, outputFiles, exports, err := ll.eval(values)
	if err != nil {
		return nil, err
	}

	docSets, err = (&OverlayPostProcessing{docSets: docSets}).Apply()
	if err != nil {
		return nil, err
	}

	result := &eval.Result{
		Files:   outputFiles,
		DocSet:  &yamlmeta.DocumentSet{},
		DocSets: map[string]*yamlmeta.DocumentSet{},
		Exports: exports,
	}

	for _, fileInLib := range ll.sortedOutputDocSets(docSets) {
		docSet := docSets[fileInLib]
		result.DocSets[fileInLib.RelativePath()] = docSet
		result.DocSet.Items = append(result.DocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return nil, fmt.Errorf("Marshaling template result: %s", err)
		}

		ll.ui.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultDocBytes)
		result.Files = append(result.Files, files.NewOutputFile(fileInLib.RelativePath(), resultDocBytes))
	}

	return result, nil
}

func (ll *LibraryLoader) eval(values eval.ValuesAst) (map[*FileInLibrary]*yamlmeta.DocumentSet, []files.OutputFile, []eval.Export, error) {
	loader := NewTemplateLoader(values, ll.ui, ll.templateLoaderOpts)

	docSets := map[*FileInLibrary]*yamlmeta.DocumentSet{}
	outputFiles := []files.OutputFile{}
	var exports []eval.Export

	for _, fileInLib := range ll.evalFiles {
		switch fileInLib.File.Type() {
		case files.TypeYAML:

			globals, resultDocSet, err := loader.EvalYAML(fileInLib)
			if err != nil {
				return nil, nil, nil, err
			}

			exports = append(exports, eval.Export{fileInLib.RelativePath(), globals})

			if fileInLib.File.IsForOutput() {
				docSets[fileInLib] = resultDocSet
			}

		case files.TypeText:
			globals, resultVal, err := loader.EvalText(fileInLib)
			if err != nil {
				return nil, nil, nil, err
			}

			exports = append(exports, eval.Export{fileInLib.RelativePath(), globals})

			if fileInLib.File.IsForOutput() {
				resultStr := resultVal.AsString()

				ll.ui.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultStr)
				outputFiles = append(outputFiles, files.NewOutputFile(fileInLib.RelativePath(), []byte(resultStr)))
			}

		case files.TypeStarlark:
			globals, err := loader.EvalStarlark(fileInLib)
			if err != nil {
				return nil, nil, nil, err
			}

			exports = append(exports, eval.Export{fileInLib.RelativePath(), globals})

		default:
			if fileInLib.File.IsForOutput() {
				return nil, nil, nil, fmt.Errorf("Unknown file type")
			}
		}
	}

	return docSets, outputFiles, exports, nil
}

func (*LibraryLoader) sortedOutputDocSets(outputDocSets map[*FileInLibrary]*yamlmeta.DocumentSet) []*FileInLibrary {
	var files []*FileInLibrary
	for file, _ := range outputDocSets {
		files = append(files, file)
	}
	SortFilesInLibrary(files)
	return files
}
