// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
)

type OverlayPostProcessing struct {
	docSets map[*FileInLibrary]*yamlmeta.DocumentSet
}

func (o OverlayPostProcessing) Apply() (map[*FileInLibrary]*yamlmeta.DocumentSet, error) {
	overlayDocSets := map[*FileInLibrary][]*yamlmeta.Document{}
	docSetsWithoutOverlays := []*yamlmeta.DocumentSet{}
	docSetToFilesMapping := map[*yamlmeta.DocumentSet]*FileInLibrary{}

	for file, docSet := range o.docSets {
		var newItems []*yamlmeta.Document
		for _, doc := range docSet.Items {
			if template.NewAnnotations(doc).Has(yttoverlay.AnnotationMatch) {
				overlayDocSets[file] = append(overlayDocSets[file], doc)
			} else {
				// TODO avoid filtering out docs?
				if doc.IsEmpty() {
					continue
				}
				newItems = append(newItems, doc)
			}
		}

		if len(newItems) > 0 {
			docSet.Items = newItems
			docSetsWithoutOverlays = append(docSetsWithoutOverlays, docSet)
			docSetToFilesMapping[docSet] = file
		}
	}

	// Respect assigned file order for data values overlaying to succeed
	var sortedOverlayFiles []*FileInLibrary
	for file := range overlayDocSets {
		sortedOverlayFiles = append(sortedOverlayFiles, file)
	}
	SortFilesInLibrary(sortedOverlayFiles)

	for _, file := range sortedOverlayFiles {
		for _, overlay := range overlayDocSets[file] {
			op := yttoverlay.Op{
				// special case: array of docsets so that file association can be preserved
				Left: docSetsWithoutOverlays,
				Right: &yamlmeta.DocumentSet{
					Items: []*yamlmeta.Document{overlay},
				},
				Thread: &starlark.Thread{Name: "overlay-post-processing"},
			}
			newLeft, err := op.Apply()
			if err != nil {
				return nil, fmt.Errorf("Overlaying (in following order: %s): %s",
					o.allFileDescs(sortedOverlayFiles), err)
			}
			docSetsWithoutOverlays = newLeft.([]*yamlmeta.DocumentSet)
		}
	}

	result := map[*FileInLibrary]*yamlmeta.DocumentSet{}

	for _, docSet := range docSetsWithoutOverlays {
		if file, ok := docSetToFilesMapping[docSet]; ok {
			result[file] = docSet
		} else {
			return nil, fmt.Errorf("Expected to find file for docset")
		}
	}

	return result, nil
}

func (o OverlayPostProcessing) allFileDescs(files []*FileInLibrary) string {
	var result []string
	for _, fileInLib := range files {
		result = append(result, fileInLib.File.RelativePath())
	}
	return strings.Join(result, ", ")
}
