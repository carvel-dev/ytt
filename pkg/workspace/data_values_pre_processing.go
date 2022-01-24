// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/workspace/datavalues"
	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

// DataValuesPreProcessing combines all data values documents (and any overlays) into a result set.
type DataValuesPreProcessing struct {
	valuesFiles    []*FileInLibrary
	valuesOverlays []*datavalues.Envelope
	schema         *datavalues.Schema
	loader         *TemplateLoader
}

// Apply executes the pre-processing of data values for all libraries.
//
// Returns the data values for the root library and enveloped data values for children libraries.
func (pp DataValuesPreProcessing) Apply() (*datavalues.Envelope, []*datavalues.Envelope, error) {
	files := append([]*FileInLibrary{}, pp.valuesFiles...)

	// Respect assigned file order for data values overlaying to succeed
	SortFilesInLibrary(files)

	dataValues, libraryDataValues, err := pp.apply(files)
	if err != nil {
		errMsg := "Overlaying data values (in following order: %s): %s"
		return nil, nil, fmt.Errorf(errMsg, pp.allFileDescs(files), err)
	}

	return dataValues, libraryDataValues, nil
}

func (pp DataValuesPreProcessing) apply(files []*FileInLibrary) (*datavalues.Envelope, []*datavalues.Envelope, error) {
	allDvs, err := pp.collectDataValuesDocs(files)
	if err != nil {
		return nil, nil, err
	}

	// merge all Data Values YAML documents into one
	var childrenLibDVs []*datavalues.Envelope
	var dvsDoc *yamlmeta.Document
	for _, dv := range allDvs {
		if dv.IntendedForAnotherLibrary() {
			childrenLibDVs = append(childrenLibDVs, dv)
			continue
		}

		if dvsDoc == nil {
			dvsDoc = dv.Doc
		} else {
			dvsDoc, err = pp.overlay(dvsDoc, dv.Doc)
			if err != nil {
				return nil, nil, err
			}
		}
		typeCheck := pp.typeAndCheck(dvsDoc)
		if len(typeCheck.Violations) > 0 {
			return nil, nil, schema.NewSchemaError("One or more data values were invalid", typeCheck.Violations...)
		}
	}

	if dvsDoc == nil {
		dvsDoc = datavalues.NewEmptyDataValuesDocument()
	}
	dataValues, err := datavalues.NewEnvelope(dvsDoc)
	if err != nil {
		return nil, nil, err
	}

	return dataValues, childrenLibDVs, nil
}

func (pp DataValuesPreProcessing) collectDataValuesDocs(dvFiles []*FileInLibrary) ([]*datavalues.Envelope, error) {
	var allDvs []*datavalues.Envelope
	if defaults := pp.schema.DefaultDataValues(); defaults != nil {
		dv, err := datavalues.NewEnvelope(defaults)
		if err != nil {
			return nil, err
		}
		allDvs = append(allDvs, dv)
	}
	for _, file := range dvFiles {
		docs, err := pp.extractDataValueDocs(file)
		if err != nil {
			return nil, fmt.Errorf("Templating file '%s': %s", file.File.RelativePath(), err)
		}
		for _, d := range docs {
			dv, err := datavalues.NewEnvelope(d)
			if err != nil {
				return nil, err
			}
			allDvs = append(allDvs, dv)
		}
	}
	allDvs = append(allDvs, pp.valuesOverlays...)
	return allDvs, nil
}

func (pp DataValuesPreProcessing) typeAndCheck(dataValuesDoc *yamlmeta.Document) schema.TypeCheck {
	chk := pp.schema.AssignType(dataValuesDoc)
	if len(chk.Violations) > 0 {
		return chk
	}
	chk = schema.CheckNode(dataValuesDoc)
	return chk
}

func (pp DataValuesPreProcessing) extractDataValueDocs(dvFile *FileInLibrary) ([]*yamlmeta.Document, error) {
	libraryCtx := LibraryExecutionContext{Current: dvFile.Library, Root: NewRootLibrary(nil)}

	_, resultDocSet, err := pp.loader.EvalYAML(libraryCtx, dvFile.File)
	if err != nil {
		return nil, err
	}

	valuesDocs, nonValuesDocs, err := DocExtractor{resultDocSet}.Extract(datavalues.AnnotationDataValues)
	if err != nil {
		return nil, err
	}

	// For simplicity's sake, prohibit mixing data value documents with other kinds.
	if len(nonValuesDocs) > 0 {
		for _, doc := range nonValuesDocs {
			if !doc.IsEmpty() {
				errStr := "Expected data values file '%s' to only have data values documents"
				return nil, fmt.Errorf(errStr, dvFile.File.RelativePath())
			}
		}
	}

	return valuesDocs, nil
}

func (pp DataValuesPreProcessing) allFileDescs(files []*FileInLibrary) string {
	var result []string
	for _, f := range files {
		result = append(result, f.File.RelativePath())
	}
	if len(pp.valuesOverlays) > 0 {
		result = append(result, "additional data values")
	}
	return strings.Join(result, ", ")
}

func (pp DataValuesPreProcessing) overlay(doc, overlay *yamlmeta.Document) (*yamlmeta.Document, error) {
	op := yttoverlay.Op{
		Left:   &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{doc}},
		Right:  &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{overlay}},
		Thread: &starlark.Thread{Name: "data-values-pre-processing"},

		ExactMatch: true,
	}

	result, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return result.(*yamlmeta.DocumentSet).Items[0], nil
}
