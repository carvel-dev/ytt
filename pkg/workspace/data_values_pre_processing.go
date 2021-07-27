// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

type DataValuesPreProcessing struct {
	valuesFiles           []*FileInLibrary
	valuesOverlays        []*DataValues
	schema                Schema
	loader                *TemplateLoader
	IgnoreUnknownComments bool // TODO remove?
}

func (o DataValuesPreProcessing) Apply() (*DataValues, []*DataValues, error) {
	files := append([]*FileInLibrary{}, o.valuesFiles...)

	// Respect assigned file order for data values overlaying to succeed
	SortFilesInLibrary(files)

	dataValues, libraryDataValues, err := o.apply(files)
	if err != nil {
		errMsg := "Overlaying data values (in following order: %s): %s"
		return nil, nil, fmt.Errorf(errMsg, o.allFileDescs(files), err)
	}

	return dataValues, libraryDataValues, nil
}

func (o DataValuesPreProcessing) apply(files []*FileInLibrary) (*DataValues, []*DataValues, error) {
	allDvs, err := o.collectDataValuesDocs(files)
	if err != nil {
		return nil, nil, err
	}

	// merge all Data Values YAML documents into one
	var otherLibraryDVs []*DataValues
	var resultDVsDoc *yamlmeta.Document
	for _, dv := range allDvs {
		if dv.IntendedForAnotherLibrary() {
			otherLibraryDVs = append(otherLibraryDVs, dv)
			continue
		}

		if resultDVsDoc == nil {
			resultDVsDoc = dv.Doc
		} else {
			resultDVsDoc, err = o.overlay(resultDVsDoc, dv.Doc)
			if err != nil {
				return nil, nil, err
			}
		}
		typeCheck := o.typeAndCheck(resultDVsDoc)
		if len(typeCheck.Violations) > 0 {
			return nil, nil, schema.NewSchemaError("One or more data values were invalid", typeCheck.Violations...)
		}
	}

	if resultDVsDoc == nil {
		resultDVsDoc = newEmptyDataValuesDocument()
	}
	dataValues, err := NewDataValues(resultDVsDoc)
	if err != nil {
		return nil, nil, err
	}
	return dataValues, otherLibraryDVs, nil
}

func (o DataValuesPreProcessing) collectDataValuesDocs(files []*FileInLibrary) ([]*DataValues, error) {
	var allDvs []*DataValues
	if defaults := o.schema.DefaultDataValues(); defaults != nil {
		dv, err := NewDataValues(defaults)
		if err != nil {
			return nil, err
		}
		// o.schema has already been determined to be the schema for the current library.
		// set the default data value libref to nil, signaling that it is for the current library.
		dv.libRef = nil
		allDvs = append(allDvs, dv)
	}
	for _, fileInLib := range files {
		docs, err := o.extractDataValueDocs(fileInLib)
		if err != nil {
			return nil, fmt.Errorf("Templating file '%s': %s", fileInLib.File.RelativePath(), err)
		}
		for _, doc := range docs {
			dv, err := NewDataValues(doc)
			if err != nil {
				return nil, err
			}
			allDvs = append(allDvs, dv)
		}
	}
	allDvs = append(allDvs, o.valuesOverlays...)
	return allDvs, nil
}

func (o DataValuesPreProcessing) typeAndCheck(dataValuesDoc *yamlmeta.Document) yamlmeta.TypeCheck {
	chk := o.schema.AssignType(dataValuesDoc)
	if _, checkable := o.schema.(*schema.DocumentSchema); checkable {
		if len(chk.Violations) > 0 {
			return chk
		}
		chk = dataValuesDoc.Check()
	}
	return chk
}

func (o DataValuesPreProcessing) allFileDescs(files []*FileInLibrary) string {
	var result []string
	for _, fileInLib := range files {
		result = append(result, fileInLib.File.RelativePath())
	}
	if len(o.valuesOverlays) > 0 {
		result = append(result, "additional data values")
	}
	return strings.Join(result, ", ")
}

func (o DataValuesPreProcessing) extractDataValueDocs(fileInLib *FileInLibrary) ([]*yamlmeta.Document, error) {
	libraryCtx := LibraryExecutionContext{Current: fileInLib.Library, Root: NewRootLibrary(nil)}

	_, resultDocSet, err := o.loader.EvalYAML(libraryCtx, fileInLib.File)
	if err != nil {
		return nil, err
	}

	// Extract _all_ data values docs from the templated result
	valuesDocs, nonValuesDocs, err := DocExtractor{resultDocSet}.Extract(AnnotationDataValues)
	if err != nil {
		return nil, err
	}

	// Fail if there any non-empty docs that are not data values
	if len(nonValuesDocs) > 0 {
		for _, doc := range nonValuesDocs {
			if !doc.IsEmpty() {
				errStr := "Expected data values file '%s' to only have data values documents"
				return nil, fmt.Errorf(errStr, fileInLib.File.RelativePath())
			}
		}
	}

	return valuesDocs, nil
}

func (o DataValuesPreProcessing) overlay(dataValues, overlay *yamlmeta.Document) (*yamlmeta.Document, error) {
	op := yttoverlay.Op{
		Left:   &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{dataValues}},
		Right:  &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{overlay}},
		Thread: &starlark.Thread{Name: "data-values-pre-processing"},

		ExactMatch: true,
	}

	newLeft, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return newLeft.(*yamlmeta.DocumentSet).Items[0], nil
}
