// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

type LibraryExecution struct {
	libraryCtx         LibraryExecutionContext
	ui                 ui.UI
	templateLoaderOpts TemplateLoaderOpts
	libraryExecFactory *LibraryExecutionFactory
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

func NewLibraryExecution(libraryCtx LibraryExecutionContext,
	ui ui.UI, templateLoaderOpts TemplateLoaderOpts,
	libraryExecFactory *LibraryExecutionFactory) *LibraryExecution {

	return &LibraryExecution{
		libraryCtx:         libraryCtx,
		ui:                 ui,
		templateLoaderOpts: templateLoaderOpts,
		libraryExecFactory: libraryExecFactory,
	}
}

func (ll *LibraryExecution) Schemas(schemaOverlays []*schema.DocumentSchemaEnvelope) (Schema, []*schema.DocumentSchemaEnvelope, error) {
	loader := NewTemplateLoader(NewEmptyDataValues(), nil, nil, ll.templateLoaderOpts, ll.libraryExecFactory, ll.ui)

	schemaFiles, err := ll.schemaFiles(loader)
	if err != nil {
		return nil, nil, err
	}

	documentSchemas, err := collectSchemaDocs(schemaFiles, loader)
	if err != nil {
		return nil, nil, err
	}

	documentSchemas = append(documentSchemas, schemaOverlays...)

	var resultSchemasDoc *yamlmeta.Document
	var childLibrarySchemas []*schema.DocumentSchemaEnvelope
	for _, docSchema := range documentSchemas {
		if docSchema.IntendedForAnotherLibrary() {
			childLibrarySchemas = append(childLibrarySchemas, docSchema)
			continue
		}
		if resultSchemasDoc == nil {
			resultSchemasDoc = docSchema.Source()
		} else {
			resultSchemasDoc, err = ll.overlay(resultSchemasDoc, docSchema.Source())
			if err != nil {
				return nil, nil, err
			}
		}
	}
	if resultSchemasDoc != nil {
		currentLibrarySchema, err := schema.NewDocumentSchema(resultSchemasDoc)
		if err != nil {
			return nil, nil, err
		}
		return currentLibrarySchema, childLibrarySchemas, nil
	}
	return schema.NewNullSchema(), childLibrarySchemas, nil
}

func collectSchemaDocs(schemaFiles []*FileInLibrary, loader *TemplateLoader) ([]*schema.DocumentSchemaEnvelope, error) {
	var documentSchemas []*schema.DocumentSchemaEnvelope
	for _, file := range schemaFiles {
		libraryCtx := LibraryExecutionContext{Current: file.Library, Root: NewRootLibrary(nil)}

		_, resultDocSet, err := loader.EvalYAML(libraryCtx, file.File)
		if err != nil {
			return nil, err
		}

		docs, _, err := DocExtractor{resultDocSet}.Extract(AnnotationDataValuesSchema)
		if err != nil {
			return nil, err
		}
		for _, doc := range docs {
			newSchema, err := schema.NewDocumentSchemaEnvelope(doc)
			if err != nil {
				return nil, err
			}
			documentSchemas = append(documentSchemas, newSchema)
		}
	}
	return documentSchemas, nil
}

func (ll *LibraryExecution) overlay(schema, overlay *yamlmeta.Document) (*yamlmeta.Document, error) {
	op := yttoverlay.Op{
		Left:   &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{schema}},
		Right:  &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{overlay}},
		Thread: &starlark.Thread{Name: "schema-pre-processing"},

		ExactMatch: true,
	}

	newLeft, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return newLeft.(*yamlmeta.DocumentSet).Items[0], nil
}

func (ll *LibraryExecution) Values(valuesOverlays []*DataValues, schema Schema) (*DataValues, []*DataValues, error) {
	loader := NewTemplateLoader(NewEmptyDataValues(), nil, nil, ll.templateLoaderOpts, ll.libraryExecFactory, ll.ui)

	valuesFiles, err := ll.valuesFiles(loader)
	if err != nil {
		return nil, nil, err
	}

	dvpp := DataValuesPreProcessing{
		valuesFiles:           valuesFiles,
		valuesOverlays:        valuesOverlays,
		schema:                schema,
		loader:                loader,
		IgnoreUnknownComments: ll.templateLoaderOpts.IgnoreUnknownComments,
	}

	return dvpp.Apply()
}

func (ll *LibraryExecution) schemaFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	return ll.filesByAnnotation(AnnotationDataValuesSchema, loader)
}

func (ll *LibraryExecution) valuesFiles(loader *TemplateLoader) ([]*FileInLibrary, error) {
	return ll.filesByAnnotation(AnnotationDataValues, loader)

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

func (ll *LibraryExecution) Eval(values *DataValues, libraryValues []*DataValues, librarySchemas []*schema.DocumentSchemaEnvelope) (*EvalResult, error) {
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

func (ll *LibraryExecution) eval(values *DataValues, libraryValues []*DataValues, librarySchemas []*schema.DocumentSchemaEnvelope) ([]EvalExport, map[*FileInLibrary]*yamlmeta.DocumentSet, []files.OutputFile, error) {

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

func (LibraryExecution) checkUnusedDVsOrSchemas(libraryValues []*DataValues, librarySchemas []*schema.DocumentSchemaEnvelope) error {
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
