package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type DataValuesPreProcessing struct {
	valuesFiles           []*FileInLibrary
	valuesAsts            []EvalValuesAst
	loader                *TemplateLoader
	IgnoreUnknownComments bool // TODO remove?
}

func (o DataValuesPreProcessing) Apply() (interface{}, error) {
	// Respect assigned file order for data values overlaying to succeed
	SortFilesInLibrary(o.valuesFiles)

	result, err := o.apply()
	if err != nil {
		return nil, fmt.Errorf("Overlaying data values (in following order: %s): %s", o.allFileDescs(), err)
	}

	return result, nil
}

func (o DataValuesPreProcessing) apply() (interface{}, error) {
	var values *yamlmeta.Document

	for _, fileInLib := range o.valuesFiles {
		valuesDocs, err := o.templateFile(fileInLib)
		if err != nil {
			return nil, fmt.Errorf("Templating file '%s': %s", fileInLib.File.RelativePath(), err)
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

	finalValues, err := o.overlayAsts(values)
	if err != nil {
		return nil, err
	}

	return finalValues.AsInterface(), nil
}

func (p DataValuesPreProcessing) allFileDescs() string {
	var result []string

	for _, fileInLib := range p.valuesFiles {
		result = append(result, fileInLib.File.RelativePath())
	}

	if len(p.valuesAsts) > 0 {
		result = append(result, "additional data values")
	}

	return strings.Join(result, ", ")
}

func (p DataValuesPreProcessing) templateFile(fileInLib *FileInLibrary) ([]*yamlmeta.Document, error) {
	libraryCtx := LibraryExecutionContext{Current: fileInLib.Library, Root: NewRootLibrary(nil)}

	_, resultDocSet, err := p.loader.EvalYAML(libraryCtx, fileInLib.File)
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
		Left:   &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{valuesDoc}},
		Right:  &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{newValuesDoc}},
		Thread: &starlark.Thread{Name: "data-values-pre-processing"},

		ExactMatch: true,
	}

	newLeft, err := op.Apply()
	if err != nil {
		return nil, err
	}

	return newLeft.(*yamlmeta.DocumentSet).Items[0], nil
}

func (p DataValuesPreProcessing) overlayAsts(valuesDoc *yamlmeta.Document) (*yamlmeta.Document, error) {
	if valuesDoc == nil {
		// TODO get rid of assumption that data values is a map?
		valuesDoc = &yamlmeta.Document{
			Value:    &yamlmeta.Map{},
			Position: filepos.NewUnknownPosition(),
		}
	}

	var result *yamlmeta.Document

	// by default return itself
	result = valuesDoc

	for _, valuesAst := range p.valuesAsts {
		var err error

		astFlagValues := &yamlmeta.Document{
			Value:    valuesAst,
			Position: filepos.NewUnknownPosition(),
		}

		result, err = p.overlay(result, astFlagValues)
		if err != nil {
			// TODO improve error message?
			return nil, fmt.Errorf("Overlaying additional data values on top of "+
				"data values from files (marked as @data/values): %s", err)
		}
	}

	return result, nil
}
