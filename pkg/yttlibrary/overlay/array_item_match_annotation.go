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

type ArrayItemMatchAnnotation struct {
	newItem *yamlmeta.ArrayItem
	thread  *starlark.Thread

	matcher     *starlark.Value
	expects     MatchAnnotationExpectsKwarg
	unannotated bool
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
		var expectsNone starlark.Value = starlark.MakeInt(0)
		annotation.expects = MatchAnnotationExpectsKwarg{
			expects: &expectsNone,
			thread:  thread,
		}
		annotation.unannotated = true
		return annotation, nil
	}

	kwargs := anns.Kwargs(AnnotationMatch)
	if len(kwargs) == 0 {
		return annotation, fmt.Errorf("Expected '%s' annotation to have "+
			"at least one keyword argument (by=..., expects=...)", AnnotationMatch)
	}

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

func (a ArrayItemMatchAnnotation) Indexes(leftArray *yamlmeta.Array) ([]int, error) {
	idxs, matches, err := a.MatchNodes(leftArray)
	if err != nil {
		return nil, err
	}

	return idxs, a.expects.Check(matches)
}

func (a ArrayItemMatchAnnotation) MatchNodes(leftArray *yamlmeta.Array) ([]int, []*filepos.Position, error) {
	if a.unannotated {
		return nil, nil, nil
	}

	matcher := a.matcher

	if matcher == nil {
		return nil, nil, fmt.Errorf("Expected '%s' annotation "+
			"keyword argument 'by' to be specified", AnnotationMatch)
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

		for i, item := range leftArray.Items {
			matcherArgs := starlark.Tuple{
				starlark.MakeInt(i),
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
