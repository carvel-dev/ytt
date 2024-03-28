// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/cmd/ui"
	"carvel.dev/ytt/pkg/files"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/validations"
	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/starlark-go/starlark"
)

// LibraryExecution is the total set of configuration and dependencies that are used to accomplish the execution of a
// Library.
type LibraryExecution struct {
	libraryCtx               LibraryExecutionContext
	ui                       ui.UI
	templateLoaderOpts       TemplateLoaderOpts
	libraryExecFactory       *LibraryExecutionFactory
	skipDataValuesValidation bool // when true, any validation rules present on data values are skipped
}

type EvalResult struct {
	Files   []files.OutputFile
	DocSet  *yamlmeta.DocumentSet
	Exports []EvalExport
}

type EvalExport struct {
	Path    string
	Symbols starlark.StringDict
}

// NewLibraryExecution configures a new instance of a LibraryExecution.
func NewLibraryExecution(libraryCtx LibraryExecutionContext, ui ui.UI, templateLoaderOpts TemplateLoaderOpts, libraryExecFactory *LibraryExecutionFactory, skipDataValuesValidation bool) *LibraryExecution {

	return &LibraryExecution{
		libraryCtx:               libraryCtx,
		ui:                       ui,
		templateLoaderOpts:       templateLoaderOpts,
		libraryExecFactory:       libraryExecFactory,
		skipDataValuesValidation: skipDataValuesValidation,
	}
}

// Schemas calculates the final schema for the Data Values in this library by combining/overlaying the schema file(s)
// in the library and the passed-in overlays.
//
// Returns this library's Schema and a slice of Schema intended for child libraries.
func (ll *LibraryExecution) Schemas(overlays []*datavalues.SchemaEnvelope) (*datavalues.Schema, []*datavalues.SchemaEnvelope, error) {
	loader := NewTemplateLoader(datavalues.NewEmptyEnvelope(), nil, nil, ll.templateLoaderOpts, ll.libraryExecFactory, ll.ui)

	files, err := ll.schemaFiles(loader)
	if err != nil {
		return nil, nil, err
	}

	spp := DataValuesSchemaPreProcessing{
		schemaFiles:    files,
		schemaOverlays: overlays,
		loader:         loader,
		rootLibrary:    ll.libraryCtx.Root,
	}

	return spp.Apply()
}

// Values calculates the final Data Values for this library by combining/overlaying defaults from the schema, the Data
// Values file(s) in the library, and the passed-in Data Values overlays. The final Data Values are validated using
// the validations annotated on a Data Value.
//
// Returns this library's Data Values and a collection of Data Values addressed to child libraries.
// Returns an error if the overlay operation fails or the result over an overlay fails a schema check.
func (ll *LibraryExecution) Values(valuesOverlays []*datavalues.Envelope, schema *datavalues.Schema) (*datavalues.Envelope, []*datavalues.Envelope, error) {
	loader := NewTemplateLoader(datavalues.NewEmptyEnvelope(), nil, nil, ll.templateLoaderOpts, ll.libraryExecFactory, ll.ui)

	valuesFiles, err := ll.valuesFiles(loader)
	if err != nil {
		return nil, nil, err
	}

	dvpp := DataValuesPreProcessing{
		valuesFiles:    valuesFiles,
		valuesOverlays: valuesOverlays,
		schema:         schema,
		loader:         loader,
		rootLibrary:    ll.libraryCtx.Root,
	}

	values, libValues, err := dvpp.Apply()
	if err != nil {
		return nil, nil, err
	}

	if !ll.skipDataValuesValidation {
		err = ll.validateValues(values)
		if err != nil {
			return nil, nil, fmt.Errorf("Validating final data values:\n%s", err)
		}
	}
	return values, libValues, err
}

// validateValues runs validations on Data Values for the current library.
// Validations are attached to data value and come from two sources:
//  1. @schema/validation annotations in a data values schema file.
//  2. @assert/validate annotations in a data values file.
//
// Returns an error if the arguments to an @assert/validate are invalid,
// otherwise, checks the Check for violations, and returns nil if there are no violations.
func (ll *LibraryExecution) validateValues(values *datavalues.Envelope) error {
	err := validations.ProcessAssertValidateAnns(values.Doc)
	if err != nil {
		return err
	}

	chk, err := validations.Run(values.Doc, "run-data-values-validations")
	if err != nil {
		return err
	}

	if chk.HasInvalidations() {
		return errors.New(chk.ResultsAsString())
	}

	return nil
}

func (ll *LibraryExecution) schemaFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	return ll.filesByAnnotation(datavalues.AnnotationDataValuesSchema, loader)
}

func (ll *LibraryExecution) valuesFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	return ll.filesByAnnotation(datavalues.AnnotationDataValues, loader)

}

func (ll *LibraryExecution) filesByAnnotation(annName template.AnnotationName, loader *TemplateLoader) ([]*FileInLibrary, error) {
	var valuesFiles []*FileInLibrary

	for _, fileInLib := range ll.libraryCtx.Current.ListAccessibleFiles() {
		if fileInLib.File.Type() == files.TypeYAML && fileInLib.File.IsTemplate() {
			docSet, err := loader.EvalPlainYAML(fileInLib.File)
			if err != nil {
				return nil, err
			}

			values, _, err := DocExtractor{docSet}.Extract(annName)
			if err != nil {
				return nil, err
			}

			if len(values) > 0 {
				valuesFiles = append(valuesFiles, fileInLib)
				fileInLib.File.MarkForOutput(false)
			}
		}
	}

	return valuesFiles, nil
}

// Eval given the final data values (as the parameter values) runs this LibraryExecution, evaluating all templates in
// this library and then applying overlays over that result.
//
// Returns the final set of Documents and output files.
// Returns an error if any template fails to evaluate, any overlay fails to apply, or if one or more "Envelopes" were
// not delivered/used.
func (ll *LibraryExecution) Eval(values *datavalues.Envelope, libraryValues []*datavalues.Envelope, librarySchemas []*datavalues.SchemaEnvelope) (*EvalResult, error) {
	exports, docSets, outputFiles, err := ll.eval(values, libraryValues, librarySchemas)
	if err != nil {
		return nil, err
	}

	docSets, err = (&OverlayPostProcessing{docSets: docSets}).Apply()
	if err != nil {
		return nil, err
	}

	result := &EvalResult{
		Files:   outputFiles,
		DocSet:  &yamlmeta.DocumentSet{},
		Exports: exports,
	}

	for _, fileInLib := range ll.sortedOutputDocSets(docSets) {
		docSet := docSets[fileInLib]
		result.DocSet.Items = append(result.DocSet.Items, docSet.Items...)

		resultDocBytes, err := docSet.AsBytes()
		if err != nil {
			return nil, fmt.Errorf("Marshaling template result: %s", err)
		}

		ll.ui.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultDocBytes)
		result.Files = append(result.Files, files.NewOutputFile(fileInLib.RelativePath(), resultDocBytes, fileInLib.File.Type()))
	}

	return result, nil
}

func (ll *LibraryExecution) eval(values *datavalues.Envelope, libraryValues []*datavalues.Envelope, librarySchemas []*datavalues.SchemaEnvelope) ([]EvalExport, map[*FileInLibrary]*yamlmeta.DocumentSet, []files.OutputFile, error) {

	loader := NewTemplateLoader(values, libraryValues, librarySchemas, ll.templateLoaderOpts, ll.libraryExecFactory, ll.ui)

	exports := []EvalExport{}
	docSets := map[*FileInLibrary]*yamlmeta.DocumentSet{}
	outputFiles := []files.OutputFile{}

	for _, fileInLib := range ll.libraryCtx.Current.ListAccessibleFiles() {
		libraryCtx := LibraryExecutionContext{Current: fileInLib.Library, Root: ll.libraryCtx.Root}

		switch {
		case fileInLib.File.IsForOutput():
			// Do not collect globals produced by templates
			switch fileInLib.File.Type() {
			case files.TypeYAML:
				_, resultDocSet, err := loader.EvalYAML(libraryCtx, fileInLib.File)
				if err != nil {
					return nil, nil, nil, err
				}

				docSets[fileInLib] = resultDocSet

			case files.TypeText:
				_, resultVal, err := loader.EvalText(libraryCtx, fileInLib.File)
				if err != nil {
					return nil, nil, nil, err
				}

				resultStr := resultVal.AsString()

				ll.ui.Debugf("### %s result\n%s", fileInLib.RelativePath(), resultStr)
				outputFiles = append(outputFiles, files.NewOutputFile(fileInLib.RelativePath(), []byte(resultStr), fileInLib.File.Type()))

			default:
				return nil, nil, nil, fmt.Errorf("Unknown file type")
			}

		case fileInLib.File.IsLibrary():
			// Collect globals produced by library files
			var evalFunc func(LibraryExecutionContext, *files.File) (starlark.StringDict, error)

			switch fileInLib.File.Type() {
			case files.TypeYAML:
				evalFunc = func(libraryCtx LibraryExecutionContext, file *files.File) (starlark.StringDict, error) {
					globals, _, err := loader.EvalYAML(libraryCtx, fileInLib.File)
					return globals, err
				}

			case files.TypeText:
				evalFunc = func(libraryCtx LibraryExecutionContext, file *files.File) (starlark.StringDict, error) {
					globals, _, err := loader.EvalText(libraryCtx, fileInLib.File)
					return globals, err
				}

			case files.TypeStarlark:
				evalFunc = loader.EvalStarlark

			default:
				// TODO should we allow skipping over unknown library files?
				// do nothing
			}

			if evalFunc != nil {
				globals, err := evalFunc(libraryCtx, fileInLib.File)
				if err != nil {
					return nil, nil, nil, err
				}

				exports = append(exports, EvalExport{Path: fileInLib.RelativePath(), Symbols: globals})
			}

		default:
			// do nothing
		}
	}

	return exports, docSets, outputFiles, ll.checkUnusedDVsOrSchemas(libraryValues, librarySchemas)
}

func (*LibraryExecution) sortedOutputDocSets(outputDocSets map[*FileInLibrary]*yamlmeta.DocumentSet) []*FileInLibrary {
	var files []*FileInLibrary
	for file := range outputDocSets {
		files = append(files, file)
	}
	SortFilesInLibrary(files)
	return files
}

func (LibraryExecution) checkUnusedDVsOrSchemas(libraryValues []*datavalues.Envelope, librarySchemas []*datavalues.SchemaEnvelope) error {
	var unusedValuesDescs []string
	var unusedDocTypes []string
	numDVNotUsed := 0

	for _, dv := range libraryValues {
		if !dv.IsUsed() {
			unusedValuesDescs = append(unusedValuesDescs, dv.Desc())
		}
	}

	if numDVNotUsed = len(unusedValuesDescs); numDVNotUsed > 0 {
		unusedDocTypes = append(unusedDocTypes, "data values")
	}

	for _, s := range librarySchemas {
		if !s.IsUsed() {
			unusedValuesDescs = append(unusedValuesDescs, s.Desc())
		}
	}
	if len(unusedValuesDescs) > numDVNotUsed {
		unusedDocTypes = append(unusedDocTypes, "schema")
	}

	if len(unusedValuesDescs) == 0 {
		return nil
	}

	return fmt.Errorf("Expected all provided library %s documents "+
		"to be used but found unused: %s", strings.Join(unusedDocTypes, ", and "), strings.Join(unusedValuesDescs, ", "))
}
