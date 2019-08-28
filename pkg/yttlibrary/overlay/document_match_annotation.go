package overlay

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
)

type DocumentMatchAnnotation struct {
	newDoc *yamlmeta.Document
	thread *starlark.Thread

	matcher *starlark.Value
	expects MatchAnnotationExpectsKwarg
}

func NewDocumentMatchAnnotation(newDoc *yamlmeta.Document,
	defaults MatchChildDefaultsAnnotation,
	thread *starlark.Thread) (DocumentMatchAnnotation, error) {

	annotation := DocumentMatchAnnotation{
		newDoc:  newDoc,
		thread:  thread,
		expects: MatchAnnotationExpectsKwarg{thread: thread},
	}
	anns := template.NewAnnotations(newDoc)

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

func (a DocumentMatchAnnotation) IndexTuples(leftDocSets []*yamlmeta.DocumentSet) ([][]int, error) {
	idxs, err := a.MatchNodes(leftDocSets)
	if err != nil {
		return nil, err
	}

	return idxs, a.expects.Check(len(idxs))
}

func (a DocumentMatchAnnotation) MatchNodes(leftDocSets []*yamlmeta.DocumentSet) ([][]int, error) {
	if a.matcher == nil {
		return nil, fmt.Errorf("Expected '%s' annotation "+
			"keyword argument 'by'  to be specified", AnnotationMatch)
	}

	switch typedVal := (*a.matcher).(type) {
	case starlark.Callable:
		var leftIdxs [][]int
		var combinedIdx int

		for i, leftDocSet := range leftDocSets {
			for j, item := range leftDocSet.Items {
				matcherArgs := starlark.Tuple{
					starlark.MakeInt(combinedIdx),
					yamltemplate.NewStarlarkFragment(item),
					yamltemplate.NewStarlarkFragment(a.newDoc),
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
					leftIdxs = append(leftIdxs, []int{i, j})
				}

				combinedIdx += 1
			}
		}

		return leftIdxs, nil

	default:
		return nil, fmt.Errorf("Expected '%s' annotation keyword argument 'by'"+
			" to be function, but was %T", AnnotationMatch, typedVal)
	}
}
