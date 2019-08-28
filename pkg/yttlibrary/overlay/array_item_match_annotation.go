package overlay

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
)

type ArrayItemMatchAnnotation struct {
	newItem *yamlmeta.ArrayItem
	thread  *starlark.Thread

	matcher *starlark.Value
	expects MatchAnnotationExpectsKwarg
}

func NewArrayItemMatchAnnotation(newItem *yamlmeta.ArrayItem,
	defaults MatchChildDefaultsAnnotation,
	thread *starlark.Thread) (ArrayItemMatchAnnotation, error) {

	annotation := ArrayItemMatchAnnotation{
		newItem: newItem,
		thread:  thread,
		expects: MatchAnnotationExpectsKwarg{thread: thread},
	}
	anns := template.NewAnnotations(newItem)

	if !anns.Has(AnnotationMatch) {
		return annotation, fmt.Errorf(
			"Expected array item to have '%s' annotation", AnnotationMatch)
	}

	kwargs := anns.Kwargs(AnnotationMatch)
	if len(kwargs) == 0 {
		return annotation, fmt.Errorf("Expected '%s' annotation to have "+
			"at least one keyword argument (by=..., expects=...)", AnnotationMatch)
	}

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case "by":
			annotation.matcher = &kwarg[1]
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

func (a ArrayItemMatchAnnotation) Indexes(leftArray *yamlmeta.Array) ([]int, error) {
	idxs, err := a.MatchNodes(leftArray)
	if err != nil {
		return nil, err
	}

	return idxs, a.expects.Check(len(idxs))
}

func (a ArrayItemMatchAnnotation) MatchNodes(leftArray *yamlmeta.Array) ([]int, error) {
	if a.matcher == nil {
		return nil, fmt.Errorf("Expected '%s' annotation "+
			"keyword argument 'by'  to be specified", AnnotationMatch)
	}

	switch typedVal := (*a.matcher).(type) {
	case starlark.String:
		var leftIdxs []int

		for i, item := range leftArray.Items {
			result, err := overlayModule{}.compareByMapKey(string(typedVal), item, a.newItem)
			if err != nil {
				return nil, err
			}
			if result {
				leftIdxs = append(leftIdxs, i)
			}
		}

		return leftIdxs, nil

	case starlark.Callable:
		var leftIdxs []int

		for i, item := range leftArray.Items {
			matcherArgs := starlark.Tuple{
				starlark.MakeInt(i),
				yamltemplate.NewStarlarkFragment(item),
				yamltemplate.NewStarlarkFragment(a.newItem),
			}

			// TODO check thread correctness
			result, err := starlark.Call(a.thread, *a.matcher, matcherArgs, []starlark.Tuple{})
			if err != nil {
				return nil, err
			}

			resultBool, err := tplcore.NewStarlarkValue(result).AsBool()
			if err != nil {
				return nil, err
			}
			if resultBool {
				leftIdxs = append(leftIdxs, i)
			}
		}

		return leftIdxs, nil

	default:
		return nil, fmt.Errorf("Expected '%s' annotation keyword argument 'by'"+
			" to be either string (for map key) or function, but was %T", AnnotationMatch, typedVal)
	}
}
