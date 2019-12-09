package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

type LibraryLoader struct {
	library            *Library
	ui                 files.UI
	templateLoaderOpts TemplateLoaderOpts
	libraryExecFactory *LibraryExecutionFactory
}

type EvalResult struct {
	Files  []files.OutputFile
	DocSet *yamlmeta.DocumentSet
}

type EvalValuesAst interface{}

func NewLibraryLoader(library *Library, ui files.UI, templateLoaderOpts TemplateLoaderOpts,
	libraryExecFactory *LibraryExecutionFactory) *LibraryLoader {

	return &LibraryLoader{
		library:            library,
		ui:                 ui,
		templateLoaderOpts: templateLoaderOpts,
		libraryExecFactory: libraryExecFactory,
	}
}

func (ll *LibraryLoader) Values(valuesFlagsAst EvalValuesAst) (EvalValuesAst, error) {
	loader := NewTemplateLoader(nil, ll.ui, ll.templateLoaderOpts, ll.libraryExecFactory)

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

	vals, err := dvpp.Apply()
	if err != nil {
		return nil, fmt.Errorf("Processing data values: %s", err)
	}

	return vals, nil
}

func (ll *LibraryLoader) valuesFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	var valuesFiles []*FileInLibrary

	for _, fileInLib := range ll.library.ListAccessibleFiles() {
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

func (ll *LibraryLoader) Eval(values EvalValuesAst) (*EvalResult, error) {
	docSets, outputFiles, err := ll.eval(values)
	if err != nil {
		return nil, err
	}

	docSets, err = (&OverlayPostProcessing{docSets: docSets}).Apply()
	if err != nil {
		return nil, err
	}

	result := &EvalResult{
		Files:  outputFiles,
		DocSet: &yamlmeta.DocumentSet{},
	}

	for _, fileInLib := range ll.sortedOutputDocSets(docSets) {
		docSet := docSets[fileInLib]
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

func (ll *LibraryLoader) eval(values EvalValuesAst) (map[*FileInLibrary]*yamlmeta.DocumentSet, []files.OutputFile, error) {
	loader := NewTemplateLoader(values, ll.ui, ll.templateLoaderOpts, ll.libraryExecFactory)

	docSets := map[*FileInLibrary]*yamlmeta.DocumentSet{}
	outputFiles := []files.OutputFile{}

	for _, fileInLib := range ll.library.ListAccessibleFiles() {
		if !fileInLib.File.IsForOutput() {
			continue
		}

		switch fileInLib.File.Type() {
		case files.TypeYAML:
			_, resultDocSet, err := loader.EvalYAML(fileInLib.Library, fileInLib.File)
			if err != nil {
				return nil, nil, err
			}

			docSets[fileInLib] = resultDocSet

		case files.TypeText:
			_, resultVal, err := loader.EvalText(fileInLib.Library, fileInLib.File)
			if err != nil {
				return nil, nil, err
			}

			resultStr := resultVal.AsString()

			ll.ui.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultStr)
			outputFiles = append(outputFiles, files.NewOutputFile(fileInLib.RelativePath(), []byte(resultStr)))

		default:
			return nil, nil, fmt.Errorf("Unknown file type")
		}
	}

	return docSets, outputFiles, nil
}

func (*LibraryLoader) sortedOutputDocSets(outputDocSets map[*FileInLibrary]*yamlmeta.DocumentSet) []*FileInLibrary {
	var files []*FileInLibrary
	for file, _ := range outputDocSets {
		files = append(files, file)
	}
	SortFilesInLibrary(files)
	return files
}
