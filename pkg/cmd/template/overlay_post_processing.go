package template

import (
	"fmt"

	"github.com/get-ytt/ytt/pkg/template"
	"github.com/get-ytt/ytt/pkg/yamlmeta"
	yttoverlay "github.com/get-ytt/ytt/pkg/yttlibrary/overlay"
	"go.starlark.net/starlark"
)

type OverlayPostProcessing struct {
	docSets map[string]*yamlmeta.DocumentSet
}

func (o OverlayPostProcessing) Apply() (map[string]*yamlmeta.DocumentSet, error) {
	overlays := []*yamlmeta.Document{}
	newDocSets := []*yamlmeta.DocumentSet{}
	newDocSetPaths := map[*yamlmeta.DocumentSet]string{}

	for relPath, docSet := range o.docSets {
		var newItems []*yamlmeta.Document
		for _, doc := range docSet.Items {
			if template.NewAnnotations(doc).Has(yttoverlay.AnnotationMatch) {
				overlays = append(overlays, doc)
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
			newDocSets = append(newDocSets, docSet)
			newDocSetPaths[docSet] = relPath
		}
	}

	for _, overlay := range overlays {
		op := yttoverlay.OverlayOp{
			// special case: array of docsets so that file association can be preserved
			Left: newDocSets,
			Right: &yamlmeta.DocumentSet{
				Items: []*yamlmeta.Document{overlay},
			},
			Thread: &starlark.Thread{Name: "overlay-post-processing"},
		}
		newLeft, err := op.Apply()
		if err != nil {
			return nil, err
		}
		newDocSets = newLeft.([]*yamlmeta.DocumentSet)
	}

	result := map[string]*yamlmeta.DocumentSet{}

	for _, docSet := range newDocSets {
		if path, ok := newDocSetPaths[docSet]; ok {
			result[path] = docSet
		} else {
			return nil, fmt.Errorf("Expected to find path for docset")
		}
	}

	return result, nil
}
