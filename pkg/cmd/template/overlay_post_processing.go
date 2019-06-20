package template

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	yttoverlay "github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type OverlayPostProcessing struct {
	docSets map[*workspace.FileInLibrary]*yamlmeta.DocumentSet
}

func (o OverlayPostProcessing) Apply() (map[*workspace.FileInLibrary]*yamlmeta.DocumentSet, error) {
	overlayDocSets := map[*workspace.FileInLibrary][]*yamlmeta.Document{}
	docSetsWithoutOverlays := []*yamlmeta.DocumentSet{}
	docSetToFilesMapping := map[*yamlmeta.DocumentSet]*workspace.FileInLibrary{}

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
	var sortedOverlayFiles []*workspace.FileInLibrary
	for file, _ := range overlayDocSets {
		sortedOverlayFiles = append(sortedOverlayFiles, file)
	}
	workspace.SortFilesInLibrary(sortedOverlayFiles)

	for _, file := range sortedOverlayFiles {
		for _, overlay := range overlayDocSets[file] {
			op := yttoverlay.OverlayOp{
				// special case: array of docsets so that file association can be preserved
				Left: docSetsWithoutOverlays,
				Right: &yamlmeta.DocumentSet{
					Items: []*yamlmeta.Document{overlay},
				},
				Thread: &starlark.Thread{Name: "overlay-post-processing"},
			}
			newLeft, err := op.Apply()
			if err != nil {
				return nil, err
			}
			docSetsWithoutOverlays = newLeft.([]*yamlmeta.DocumentSet)
		}
	}

	result := map[*workspace.FileInLibrary]*yamlmeta.DocumentSet{}

	for _, docSet := range docSetsWithoutOverlays {
		if file, ok := docSetToFilesMapping[docSet]; ok {
			result[file] = docSet
		} else {
			return nil, fmt.Errorf("Expected to find file for docset")
		}
	}

	return result, nil
}
