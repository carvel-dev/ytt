// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"carvel.dev/ytt/pkg/workspace/datavalues"
	"carvel.dev/ytt/pkg/yamlmeta"
	yttoverlay "carvel.dev/ytt/pkg/yttlibrary/overlay"
	"github.com/k14s/starlark-go/starlark"
)

// DataValuesSchemaPreProcessing combines all data values schema documents (and any overlays) into a result set.
type DataValuesSchemaPreProcessing struct {
	schemaFiles    []*FileInLibrary
	schemaOverlays []*datavalues.SchemaEnvelope
	loader         *TemplateLoader
	rootLibrary    *Library
}

// Apply executes the pre-processing of schema for data values for all libraries.
//
// Returns the schema for the root library and enveloped schema for children libraries.
func (pp DataValuesSchemaPreProcessing) Apply() (*datavalues.Schema, []*datavalues.SchemaEnvelope, error) {
	files := append([]*FileInLibrary{}, pp.schemaFiles...)

	// Ensure files are in assigned order so that overlays will be applied correctly.
	SortFilesInLibrary(files)

	schema, libSchemas, err := pp.apply(files)
	if err != nil {
		errMsg := "Overlaying data values schema (in following order: %s): %s"
		return nil, nil, fmt.Errorf(errMsg, pp.allFileDescs(files), err)
	}

	return schema, libSchemas, nil
}

func (pp DataValuesSchemaPreProcessing) apply(files []*FileInLibrary) (*datavalues.Schema, []*datavalues.SchemaEnvelope, error) {
	allSchemas, err := pp.collectSchemaDocs(files)
	if err != nil {
		return nil, nil, err
	}

	// merge all Schema documents into one
	var schemaDoc *yamlmeta.Document
	var childLibSchemas []*datavalues.SchemaEnvelope
	for _, schema := range allSchemas {
		if schema.IntendedForAnotherLibrary() {
			childLibSchemas = append(childLibSchemas, schema)
			continue
		}

		if schemaDoc == nil {
			schemaDoc = schema.Source()
		} else {
			schemaDoc, err = pp.overlay(schemaDoc, schema.Source())
			if err != nil {
				return nil, nil, err
			}
		}
	}

	var schema *datavalues.Schema
	if schemaDoc == nil {
		schema = datavalues.NewNullSchema()
	} else {
		schema, err = datavalues.NewSchema(schemaDoc)
		if err != nil {
			return nil, nil, err
		}
	}

	return schema, childLibSchemas, nil
}

func (pp DataValuesSchemaPreProcessing) collectSchemaDocs(schemaFiles []*FileInLibrary) ([]*datavalues.SchemaEnvelope, error) {
	var allSchema []*datavalues.SchemaEnvelope
	for _, file := range schemaFiles {
		docs, err := pp.extractSchemaDocs(file)
		if err != nil {
			return nil, fmt.Errorf("Templating file '%s': %s", file.File.RelativePath(), err)
		}

		for _, d := range docs {
			s, err := datavalues.NewSchemaEnvelope(d)
			if err != nil {
				return nil, err
			}
			allSchema = append(allSchema, s)
		}
	}
	allSchema = append(allSchema, pp.schemaOverlays...)
	return allSchema, nil
}

func (pp DataValuesSchemaPreProcessing) extractSchemaDocs(schemaFile *FileInLibrary) ([]*yamlmeta.Document, error) {
	libraryCtx := LibraryExecutionContext{Current: schemaFile.Library, Root: pp.rootLibrary}

	_, resultDocSet, err := pp.loader.EvalYAML(libraryCtx, schemaFile.File)
	if err != nil {
		return nil, err
	}

	schemaDocs, nonSchemaDocs, err := DocExtractor{resultDocSet}.Extract(datavalues.AnnotationDataValuesSchema)
	if err != nil {
		return nil, err
	}

	// For simplicity's sake, prohibit mixing data value schema documents with other kinds.
	if len(nonSchemaDocs) > 0 {
		for _, doc := range nonSchemaDocs {
			if !doc.IsEmpty() {
				errStr := "Expected schema file '%s' to only have schema documents"
				return nil, fmt.Errorf(errStr, schemaFile.File.RelativePath())
			}
		}
	}

	return schemaDocs, nil
}

func (pp DataValuesSchemaPreProcessing) allFileDescs(files []*FileInLibrary) string {
	var result []string
	for _, f := range files {
		result = append(result, f.File.RelativePath())
	}
	if len(pp.schemaOverlays) > 0 {
		result = append(result, "additional data value schema")
	}
	return strings.Join(result, ", ")
}

func (pp DataValuesSchemaPreProcessing) overlay(doc, overlay *yamlmeta.Document) (*yamlmeta.Document, error) {
	op := yttoverlay.Op{
		Left:   &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{doc}},
		Right:  &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{overlay}},
		Thread: &starlark.Thread{Name: "data-values-schema-pre-processing"},

		ExactMatch: true,
	}

	result, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return result.(*yamlmeta.DocumentSet).Items[0], nil
}
