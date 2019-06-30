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
}

func NewLibraryLoader(lib *Library, ui files.UI, templateLoaderOpts TemplateLoaderOpts) *LibraryLoader {
	return &LibraryLoader{
		library:            lib,
		ui:                 ui,
		templateLoaderOpts: templateLoaderOpts,
	}
}

func LoadRootLibrary(absolutePaths []string, ui files.UI, astValues interface{}, opts TemplateLoaderOpts) (*eval.Result, error) {
	filesToProcess, err := files.NewSortedFilesFromPaths(absolutePaths)
	if err != nil {
		return nil, err
	}

	rootLibrary := NewRootLibrary(filesToProcess)

	ll := NewLibraryLoader(rootLibrary, ui, opts)

	astValues, err = ll.Values(astValues)
	if err != nil {
		return nil, err
	}

	return ll.Eval(astValues)
}

func (ll *LibraryLoader) Values(valuesFlagsAst eval.ValuesAst) (eval.ValuesAst, error) {
	valuesFiles, err := ll.valuesFiles()
	if err != nil {
		return nil, err
	}

	dvpp := DataValuesPreProcessing{
		valuesFiles:           valuesFiles,
		valuesFlagsAst:        valuesFlagsAst,
		loader:                NewTemplateLoader(nil, ll.ui, ll.templateLoaderOpts),
		IgnoreUnknownComments: ll.templateLoaderOpts.IgnoreUnknownComments,
	}

	return dvpp.Apply()
}

func (ll *LibraryLoader) valuesFiles() ([]*FileInLibrary, error) {
	var valuesFiles []*FileInLibrary

	for _, fileInLib := range ll.library.ListAccessibleFiles() {
		if fileInLib.File.Type() == files.TypeYAML && fileInLib.File.IsTemplate() {
			fileBs, err := fileInLib.File.Bytes()
			if err != nil {
				return nil, err
			}

			docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, fileInLib.File.RelativePath())
			if err != nil {
				return nil, fmt.Errorf("Unmarshaling YAML template '%s': %s", fileInLib.File.RelativePath(), err)
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
	docSets, outputFiles, err := ll.eval(values)
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

func (ll *LibraryLoader) eval(values eval.ValuesAst) (map[*FileInLibrary]*yamlmeta.DocumentSet, []files.OutputFile, error) {
	loader := NewTemplateLoader(values, ll.ui, ll.templateLoaderOpts)

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
