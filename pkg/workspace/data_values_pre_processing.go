package workspace

import (
	"fmt"

	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type DataValuesPreProcessing struct {
	valuesFiles           []*FileInLibrary
	valuesFlagsAst        interface{}
	loader                *TemplateLoader
	IgnoreUnknownComments bool // TODO remove?
}

func (o DataValuesPreProcessing) Apply() (interface{}, error) {
	var values *yamlmeta.Document

	// Respect assigned file order for data values overlaying to succeed
	SortFilesInLibrary(o.valuesFiles)

	for _, fileInLib := range o.valuesFiles {
		valuesDocs, err := o.templateFile(fileInLib)
		if err != nil {
			return nil, err
		}

		for _, valuesDoc := range valuesDocs {
			if values == nil {
				values = valuesDoc
				continue
			}

			var err error
			values, err = o.overlay(values, valuesDoc)
			if err != nil {
				return nil, err
			}
		}
	}

	valuesWithFlags, err := o.overlayFlags(values)
	if err != nil {
		return nil, err
	}

	return valuesWithFlags.AsInterface(yamlmeta.InterfaceConvertOpts{}), nil
}

func (p DataValuesPreProcessing) templateFile(fileInLib *FileInLibrary) ([]*yamlmeta.Document, error) {
	_, resultDocSet, err := p.loader.EvalYAML(fileInLib.Library, fileInLib.File)
	if err != nil {
		return nil, err
	}

	tplOpts := yamltemplate.MetasOpts{IgnoreUnknown: p.IgnoreUnknownComments}

	// Extract _all_ data values docs from the templated result
	valuesDocs, nonValuesDocs, err := yttlibrary.DataValues{resultDocSet, tplOpts}.Extract()
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

func (p DataValuesPreProcessing) overlay(valuesDoc, newValuesDoc *yamlmeta.Document) (*yamlmeta.Document, error) {
	op := yttoverlay.OverlayOp{
		Left:   valuesDoc.Value,
		Right:  newValuesDoc.Value,
		Thread: &starlark.Thread{Name: "data-values-pre-processing"},
	}

	newLeft, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return &yamlmeta.Document{Value: newLeft}, nil
}

func (p DataValuesPreProcessing) overlayFlags(valuesDoc *yamlmeta.Document) (*yamlmeta.Document, error) {
	if valuesDoc == nil {
		// TODO get rid of assumption that data values is a map?
		valuesDoc = &yamlmeta.Document{Value: &yamlmeta.Map{}}
	}

	astFlagValues := &yamlmeta.Document{Value: p.valuesFlagsAst}

	result, err := p.overlay(valuesDoc, astFlagValues)
	if err != nil {
		return nil, fmt.Errorf("Overlaying data values from flags (provided via --data-value-* or library loading) on top of data values from files (marked as @data/values): %s", err)
	}

	return result, nil
}
