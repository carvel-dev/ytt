package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
)

// LibraryLoader is an object that is able to evaluate a Library given some
// options and return an evaluation result.
type LibraryLoader struct {
	*Library
	UI   files.UI
	Opts TemplateLoaderOpts
}

// EvalResult represents the output of an evaluation.
type EvalResult struct {
	DocSet      *yamlmeta.DocumentSet
	DocSets     map[*FileInLibrary]*yamlmeta.DocumentSet
	OutputFiles []files.OutputFile
}

// EvalValuesAst represents in AST form the data.values to be fed to the
// evaluation process.
type EvalValuesAst interface{}

func NewLibraryLoader(lib *Library, ui files.UI, loaderOpts TemplateLoaderOpts) *LibraryLoader {
	return &LibraryLoader{
		UI:      ui,
		Opts:    loaderOpts,
		Library: lib,
	}
}

func (ll *LibraryLoader) LoadValues(parentValues EvalValuesAst) (EvalValuesAst, error) {
	valuesFiles, err := ll.loadValueFiles(ll.Library)
	if err != nil {
		return nil, err
	}

	noValuesLoader := NewTemplateLoader(nil, ll.UI, ll.Opts)

	vals, err := DataValuesPreProcessing{valuesFiles, parentValues, noValuesLoader, ll.Opts.IgnoreUnknownComments}.Apply()
	if err != nil {
		return nil, err
	}

	return vals, nil
}

func (ll *LibraryLoader) loadValueFiles(lib *Library) ([]*FileInLibrary, error) {
	var valuesFiles []*FileInLibrary

	for _, fileInLib := range lib.ListAccessibleFiles() {
		if fileInLib.File.Type() == files.TypeYAML && fileInLib.File.IsTemplate() {
			fileBs, err := fileInLib.File.Bytes()
			if err != nil {
				return nil, err
			}

			docSet, err := yamlmeta.NewDocumentSetFromBytes(fileBs, fileInLib.File.RelativePath())
			if err != nil {
				return nil, fmt.Errorf("Unmarshaling YAML template '%s': %s", fileInLib.File.RelativePath(), err)
			}

			tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: ll.Opts.IgnoreUnknownComments}

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
	res := &EvalResult{
		DocSet:      &yamlmeta.DocumentSet{},
		OutputFiles: []files.OutputFile{},
	}
	overlayProcess := &OverlayPostProcessing{
		docSets: map[*FileInLibrary]*yamlmeta.DocumentSet{},
	}

	err := ll.evalLibrary(ll.Library, values, res, overlayProcess)
	if err != nil {
		return nil, err
	}

	res.DocSets, err = overlayProcess.Apply()
	if err != nil {
		return nil, err
	}

	for _, fileInLib := range sortedOutputDocSets(res.DocSets) {
		docSet := res.DocSets[fileInLib]
		res.DocSet.Items = append(res.DocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return nil, fmt.Errorf("Marshaling template result: %s", err)
		}

		ll.UI.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultDocBytes)
		outputFile := files.NewOutputFile(fileInLib.RelativePath(), resultDocBytes)
		res.OutputFiles = append(res.OutputFiles, outputFile)
	}

	return res, nil
}

func (ll *LibraryLoader) evalLibrary(lib *Library, values EvalValuesAst, res *EvalResult, overlayProcess *OverlayPostProcessing) error {
	loader := NewTemplateLoader(values, ll.UI, ll.Opts)

	for _, fileInLib := range lib.ListAccessibleFiles() {
		if !fileInLib.File.IsForOutput() {
			continue
		}

		switch fileInLib.File.Type() {
		case files.TypeYAML:
			_, resultDocSet, err := loader.EvalYAML(fileInLib.Library, fileInLib.File)
			if err != nil {
				return err
			}

			overlayProcess.docSets[fileInLib] = resultDocSet

		case files.TypeText:
			_, resultVal, err := loader.EvalText(fileInLib.Library, fileInLib.File)
			if err != nil {
				return err
			}

			resultStr := resultVal.AsString()

			ll.UI.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultStr)
			res.OutputFiles = append(res.OutputFiles, files.NewOutputFile(fileInLib.RelativePath(), []byte(resultStr)))

		default:
			return fmt.Errorf("Unknown file type")
		}
	}
	return nil
}

func sortedOutputDocSets(outputDocSets map[*FileInLibrary]*yamlmeta.DocumentSet) []*FileInLibrary {
	var files []*FileInLibrary
	for file, _ := range outputDocSets {
		files = append(files, file)
	}
	SortFilesInLibrary(files)
	return files
}
