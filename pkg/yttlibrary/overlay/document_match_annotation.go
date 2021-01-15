// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

type DocumentMatchAnnotation struct {
	newDoc *yamlmeta.Document
	exact  bool
	thread *starlark.Thread

	matcher *starlark.Value
	expects MatchAnnotationExpectsKwarg
}

func NewDocumentMatchAnnotation(newDoc *yamlmeta.Document,
	defaults MatchChildDefaultsAnnotation,
	exact bool, thread *starlark.Thread) (DocumentMatchAnnotation, error) {

	annotation := DocumentMatchAnnotation{
		newDoc:  newDoc,
		exact:   exact,
		thread:  thread,
		expects: MatchAnnotationExpectsKwarg{thread: thread},
	}
	anns := template.NewAnnotations(newDoc)

	kwargs := anns.Kwargs(AnnotationMatch)
	if !exact && len(kwargs) == 0 {
		return annotation, fmt.Errorf("Expected '%s' annotation to have "+
			"at least one keyword argument (by=..., expects=...)", AnnotationMatch)
	}

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case "by":
			annotation.matcher = &kwarg[1]
		case MatchAnnotationKwargExpects:
			annotation.expects.expects = &kwarg[1]
		case MatchAnnotationKwargMissingOK:
			annotation.expects.missingOK = &kwarg[1]
		case MatchAnnotationKwargWhen:
			annotation.expects.when = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationMatch, kwargName)
		}
	}

	annotation.expects.FillInDefaults(defaults)

	return annotation, nil
}

func (a DocumentMatchAnnotation) IndexTuples(leftDocSets []*yamlmeta.DocumentSet) ([][]int, error) {
	idxs, matches, err := a.MatchNodes(leftDocSets)
	if err != nil {
		return nil, err
	}

	return idxs, a.expects.Check(matches)
}

func (a DocumentMatchAnnotation) MatchNodes(leftDocSets []*yamlmeta.DocumentSet) ([][]int, []*filepos.Position, error) {
	if a.exact {
		if len(leftDocSets) != 1 && len(leftDocSets[0].Items) != 1 {
			return nil, nil, fmt.Errorf("Expected to find exactly one left doc when merging exactly two documents")
		}
		return [][]int{{0, 0}}, []*filepos.Position{leftDocSets[0].Items[0].Position}, nil
	}

	if a.matcher == nil {
		return nil, nil, fmt.Errorf("Expected '%s' annotation "+
			"keyword argument 'by'  to be specified", AnnotationMatch)
	}

	switch typedVal := (*a.matcher).(type) {
	case starlark.Callable:
		var leftIdxs [][]int
		var combinedIdx int
		var matches []*filepos.Position

		for i, leftDocSet := range leftDocSets {
			for j, item := range leftDocSet.Items {
				matcherArgs := starlark.Tuple{
					starlark.MakeInt(combinedIdx),
					yamltemplate.NewGoValueWithYAML(item.Value).AsStarlarkValue(),
					yamltemplate.NewGoValueWithYAML(a.newDoc.Value).AsStarlarkValue(),
				}

				// TODO check thread correctness
				result, err := starlark.Call(a.thread, *a.matcher, matcherArgs, []starlark.Tuple{})
				if err != nil {
					return nil, nil, err
				}

				resultBool, err := tplcore.NewStarlarkValue(result).AsBool()
				if err != nil {
					return nil, nil, err
				}
				if resultBool {
					leftIdxs = append(leftIdxs, []int{i, j})
					matches = append(matches, item.Position)
				}

				combinedIdx++
			}
		}

		return leftIdxs, matches, nil

	default:
		return nil, nil, fmt.Errorf("Expected '%s' annotation keyword argument 'by'"+
			" to be function, but was %T", AnnotationMatch, typedVal)
	}
}
