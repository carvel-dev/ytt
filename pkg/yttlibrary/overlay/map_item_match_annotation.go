// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"
	"reflect"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/template"
	tplcore "carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yamltemplate"
	"github.com/k14s/starlark-go/starlark"
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

func (a MapItemMatchAnnotation) Indexes(leftMap *yamlmeta.Map) ([]int, error) {
	idxs, matches, err := a.MatchNodes(leftMap)
	if err != nil {
		return []int{}, err
	}

	return idxs, a.expects.Check(matches)
}

func (a MapItemMatchAnnotation) MatchNodes(leftMap *yamlmeta.Map) ([]int, []*filepos.Position, error) {
	matcher := a.matcher

	if matcher == nil {
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

	if _, ok := (*matcher).(starlark.String); ok {
		matcherFunc, err := starlark.Call(a.thread, overlayModule{}.MapKey(),
			starlark.Tuple{*matcher}, []starlark.Tuple{})
		if err != nil {
			return nil, nil, err
		}

		matcher = &matcherFunc
	}

	switch typedVal := (*matcher).(type) {
	case starlark.Callable:
		var leftIdxs []int
		var matches []*filepos.Position

		for i, item := range leftMap.Items {
			matcherArgs := starlark.Tuple{
				yamltemplate.NewGoValueWithYAML(item.Key).AsStarlarkValue(),
				yamltemplate.NewGoValueWithYAML(item.Value).AsStarlarkValue(),
				yamltemplate.NewGoValueWithYAML(a.newItem.Value).AsStarlarkValue(),
			}

			// TODO check thread correctness
			result, err := starlark.Call(a.thread, *matcher, matcherArgs, []starlark.Tuple{})
			if err != nil {
				return nil, nil, err
			}

			resultBool, err := tplcore.NewStarlarkValue(result).AsBool()
			if err != nil {
				return nil, nil, err
			}
			if resultBool {
				leftIdxs = append(leftIdxs, i)
				matches = append(matches, item.Position)
			}
		}
		return leftIdxs, matches, nil

	default:
		return nil, nil, fmt.Errorf("Expected '%s' annotation keyword argument 'by' "+
			"to be either string (for map key) or function, but was %T", AnnotationMatch, typedVal)
	}
}
