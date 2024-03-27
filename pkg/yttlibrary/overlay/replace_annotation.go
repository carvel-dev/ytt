// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"carvel.dev/ytt/pkg/template"
	tplcore "carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamltemplate"
	"github.com/k14s/starlark-go/starlark"
)

const (
	ReplaceAnnotationKwargVia   string = "via"
	ReplaceAnnotationKwargOrAdd string = "or_add"
)

type ReplaceAnnotation struct {
	newNode template.EvaluationNode
	thread  *starlark.Thread
	via     *starlark.Callable
	orAdd   bool
}

func NewReplaceAnnotation(newNode template.EvaluationNode, thread *starlark.Thread) (ReplaceAnnotation, error) {
	annotation := ReplaceAnnotation{
		newNode: newNode,
		thread:  thread,
	}
	kwargs := template.NewAnnotations(newNode).Kwargs(AnnotationReplace)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))

		switch kwargName {
		case ReplaceAnnotationKwargVia:
			via, ok := kwarg[1].(starlark.Callable)
			if !ok {
				return ReplaceAnnotation{}, fmt.Errorf(
					"Expected '%s' annotation keyword argument '%s' to be function, "+
						"but was %T", AnnotationReplace, kwargName, kwarg[1])
			}
			annotation.via = &via

		case ReplaceAnnotationKwargOrAdd:
			resultBool, err := tplcore.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return ReplaceAnnotation{}, err
			}
			annotation.orAdd = resultBool

		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationReplace, kwargName)
		}
	}

	return annotation, nil
}

func (a ReplaceAnnotation) Value(existingNode template.EvaluationNode) (interface{}, error) {
	// Make sure original nodes are not affected in any way
	newNode := a.newNode.DeepCopyAsInterface().(template.EvaluationNode)

	// TODO currently assumes that we can always get at least one value
	if a.via == nil {
		return newNode.GetValues()[0], nil
	}

	var existingVal interface{}
	if existingNode != nil {
		// Make sure original nodes are not affected in any way
		existingVal = existingNode.DeepCopyAsInterface().(template.EvaluationNode).GetValues()[0]
	} else {
		existingVal = nil
	}

	viaArgs := starlark.Tuple{
		yamltemplate.NewGoValueWithYAML(existingVal).AsStarlarkValue(),
		yamltemplate.NewGoValueWithYAML(newNode.GetValues()[0]).AsStarlarkValue(),
	}

	// TODO check thread correctness
	result, err := starlark.Call(a.thread, *a.via, viaArgs, []starlark.Tuple{})
	if err != nil {
		return nil, err
	}

	return tplcore.NewStarlarkValue(result).AsGoValue()
}

func (a ReplaceAnnotation) OrAdd() bool { return a.orAdd }
