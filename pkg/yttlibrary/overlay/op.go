package overlay

import (
	"fmt"
	// "os" // yamlmeta.NewPrinter(os.Stdout).Print(typedLeft)

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"go.starlark.net/starlark"
)

type OverlayOp struct {
	Left  interface{}
	Right interface{}

	Thread *starlark.Thread
}

func (o OverlayOp) Apply() (interface{}, error) {
	leftObj := yamlmeta.NewASTFromInterface(o.Left)
	rightObj := yamlmeta.NewASTFromInterface(o.Right)

	_, err := o.apply(leftObj, rightObj)
	if err != nil {
		return nil, err
	}

	o.removeOverlayAnns(leftObj)
	return leftObj, nil
}

func (o OverlayOp) apply(left, right interface{}) (bool, error) {
	switch typedRight := right.(type) {
	case *yamlmeta.DocumentSet:
		var docSetArray []*yamlmeta.DocumentSet

		typedLeft, isDocSet := left.(*yamlmeta.DocumentSet)
		if !isDocSet {
			// support array of docsets to allow consumers to
			// keep proper association of document to docsets
			// (see matching for overlay post processing)
			typedLeft, isDocSetArray := left.([]*yamlmeta.DocumentSet)
			if !isDocSetArray {
				return false, fmt.Errorf("Expected docset, but was %T", left)
			}
			docSetArray = typedLeft
		} else {
			docSetArray = []*yamlmeta.DocumentSet{typedLeft}
		}

		return o.applyDocSet(docSetArray, typedRight)

	case *yamlmeta.Document:
		panic("Unexpected doc")

	case *yamlmeta.Map:
		typedLeft, isMap := left.(*yamlmeta.Map)
		if !isMap {
			return false, fmt.Errorf("Expected map, but was %T", left)
		}

		for _, item := range typedRight.Items {
			item := item.DeepCopy()

			var errPrefix string = "Overlaying"
			op, err := whichOp(item)
			if err == nil {
				switch op {
				case AnnotationMerge:
					err = o.mergeMapItem(typedLeft, item)
					errPrefix = "Merging"
				case AnnotationRemove:
					err = o.removeMapItem(typedLeft, item)
					errPrefix = "Removing"
				case AnnotationReplace:
					err = o.replaceMapItem(typedLeft, item)
					errPrefix = "Replacing"
				default:
					err = fmt.Errorf("Overlay op %s is not supported on map item", op)
				}
			}
			if err != nil {
				return false, fmt.Errorf("%s map item (key '%s') on %s: %s",
					errPrefix, item.Key, item.Position.AsString(), err)
			}
		}

	case *yamlmeta.MapItem:
		panic("Unexpected mapitem")

	case *yamlmeta.Array:
		typedLeft, isArray := left.(*yamlmeta.Array)
		if !isArray {
			return false, fmt.Errorf("Expected array, but was %T", left)
		}

		for _, item := range typedRight.Items {
			item := item.DeepCopy()

			op, err := whichOp(item)
			if err == nil {
				switch op {
				case AnnotationMerge:
					err = o.mergeArrayItem(typedLeft, item)
				case AnnotationRemove:
					err = o.removeArrayItem(typedLeft, item)
				case AnnotationReplace:
					err = o.replaceArrayItem(typedLeft, item)
				case AnnotationInsert:
					err = o.insertArrayItem(typedLeft, item)
				case AnnotationAppend:
					err = o.appendArrayItem(typedLeft, item)
				default:
					err = fmt.Errorf("Overlay op %s is not supported on array item", op)
				}
			}
			if err != nil {
				return false, fmt.Errorf("Array item on %s: %s", item.Position.AsString(), err)
			}
		}

	case *yamlmeta.ArrayItem:
		panic("Unexpected arrayitem")

	default:
		return true, nil
	}

	return false, nil
}

func (o OverlayOp) applyDocSet(typedLeft []*yamlmeta.DocumentSet, typedRight *yamlmeta.DocumentSet) (bool, error) {
	for _, doc := range typedRight.Items {
		doc := doc.DeepCopy()

		op, err := whichOp(doc)
		if err == nil {
			switch op {
			case AnnotationMerge:
				err = o.mergeDocument(typedLeft, doc)
			case AnnotationRemove:
				err = o.removeDocument(typedLeft, doc)
			case AnnotationReplace:
				err = o.replaceDocument(typedLeft, doc)
			case AnnotationInsert:
				err = o.insertDocument(typedLeft, doc)
			case AnnotationAppend:
				err = o.appendDocument(typedLeft, doc)
			default:
				err = fmt.Errorf("Overlay op %s is not supported on document", op)
			}
		}
		if err != nil {
			return false, fmt.Errorf("Document on %s: %s", doc.Position.AsString(), err)
		}
	}

	return false, nil
}

func (o OverlayOp) removeOverlayAnns(val interface{}) {
	node, ok := val.(yamlmeta.Node)
	if !ok {
		return
	}

	template.NewAnnotations(node).DeleteNs(AnnotationNs)

	for _, childVal := range node.GetValues() {
		o.removeOverlayAnns(childVal)
	}
}
