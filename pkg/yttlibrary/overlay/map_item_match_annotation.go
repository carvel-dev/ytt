package overlay

import (
	"fmt"
	"github.com/k14s/ytt/pkg/filepos"
	"reflect"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"go.starlark.net/starlark"
)

type MapItemMatchAnnotation struct {
	newItem *yamlmeta.MapItem
	expects MatchAnnotationExpectsKwarg
}

func NewMapItemMatchAnnotation(newItem *yamlmeta.MapItem,
	defaults MatchChildDefaultsAnnotation,
	thread *starlark.Thread) (MapItemMatchAnnotation, error) {

	annotation := MapItemMatchAnnotation{
		newItem: newItem,
		expects: MatchAnnotationExpectsKwarg{thread: thread},
	}
	kwargs := template.NewAnnotations(newItem).Kwargs(AnnotationMatch)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case "expects":
			annotation.expects.expects = &kwarg[1]
		case "missing_ok":
			annotation.expects.missingOK = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationMatch, kwargName)
		}
	}

	annotation.expects.FillInDefaults(defaults)

	return annotation, nil
}

func (a MapItemMatchAnnotation) Index(leftMap *yamlmeta.Map) (int, bool, error) {
	idx, match, found := a.MatchNode(leftMap)

	matches := []*filepos.Position{}
	if found {
		matches = append(matches, match)
	}

	return idx, found, a.expects.Check(matches)
}

func (a MapItemMatchAnnotation) MatchNode(leftMap *yamlmeta.Map) (int, *filepos.Position, bool) {
	for i, item := range leftMap.Items {
		if reflect.DeepEqual(item.Key, a.newItem.Key) {
			return i, item.Position, true
		}
	}
	return 0, filepos.NewUnknownPosition(), false
}
