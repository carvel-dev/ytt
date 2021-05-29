// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

type DataValuesFile struct {
	doc *yamlmeta.Document
}

func NewDataValuesFile(doc *yamlmeta.Document) DataValuesFile {
	return DataValuesFile{doc.DeepCopy()}
}

func (f DataValuesFile) AsOverlay() (*yamlmeta.Document, error) {
	doc := f.doc.DeepCopy()

	if yamltemplate.HasTemplating(doc) {
		return nil, fmt.Errorf("Expected to not find annotations inside data values file " +
			"(hint: remove comments starting with '#@')")
	}

	f.addOverlayReplace(doc)

	return doc, nil
}

func (f DataValuesFile) addOverlayReplace(node yamlmeta.Node) {
	anns := template.NodeAnnotations{
		yttoverlay.AnnotationMatch: template.NodeAnnotation{
			Kwargs: []starlark.Tuple{{
				starlark.String(yttoverlay.MatchAnnotationKwargMissingOK),
				starlark.Bool(true),
			}},
		},
	}

	replaceAnn := template.NodeAnnotation{
		Kwargs: []starlark.Tuple{{
			starlark.String(yttoverlay.ReplaceAnnotationKwargOrAdd),
			starlark.Bool(true),
		}},
	}

	for _, val := range node.GetValues() {
		switch typedVal := val.(type) {
		case *yamlmeta.Array:
			anns[yttoverlay.AnnotationReplace] = replaceAnn
		case yamlmeta.Node:
			f.addOverlayReplace(typedVal)
		default:
			anns[yttoverlay.AnnotationReplace] = replaceAnn
		}
	}

	node.SetAnnotations(anns)
}
