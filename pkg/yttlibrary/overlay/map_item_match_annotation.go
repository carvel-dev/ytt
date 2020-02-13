package overlay

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"go.starlark.net/starlark"
)

type MapItemMatchAnnotation struct {
	newItem *yamlmeta.MapItem
	thread  *starlark.Thread

	matcher *starlark.Value
	expects MatchAnnotationExpectsKwarg
}

func NewMapItemMatchAnnotation(newItem *yamlmeta.MapItem,
	defaults MatchChildDefaultsAnnotation,
	thread *starlark.Thread) (MapItemMatchAnnotation, error) {

	annotation := MapItemMatchAnnotation{
		newItem: newItem,
		thread:  thread,
		expects: MatchAnnotationExpectsKwarg{thread: thread},
	}
	kwargs := template.NewAnnotations(newItem).Kwargs(AnnotationMatch)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case MatchAnnotationKwargBy:
			annotation.matcher = &kwarg[1]
		case MatchAnnotationKwargExpects:
			annotation.expects.expects = &kwarg[1]
		case MatchAnnotationKwargMissingOK:
			annotation.expects.missingOK = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationMatch, kwargName)
		}
	}

	annotation.expects.FillInDefaults(defaults)

	return annotation, nil
}

func (a MapItemMatchAnnotation) Indexes(leftMap *yamlmeta.Map) ([]int, error) {
	idxs, matches, err := a.MatchNodes(leftMap)
	if err != nil {
		return []int{}, err
	}

	return idxs, a.expects.Check(matches)
}

func (a MapItemMatchAnnotation) MatchNodes(leftMap *yamlmeta.Map) ([]int, []*filepos.Position, error) {
	var leftIdxs []int
	var matches []*filepos.Position

	for i, item := range leftMap.Items {
		if reflect.DeepEqual(item.Key, a.newItem.Key) {
			leftIdxs = append(leftIdxs, i)
			matches = append(matches, item.Position)
		}
	}

	return leftIdxs, matches, nil
}
