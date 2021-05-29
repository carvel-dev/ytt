// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

const (
	ReplaceAnnotationKwargVia   string = "via"
	ReplaceAnnotationKwargOrAdd string = "or_add"
)

type ReplaceAnnotation struct {
	newNode template.EvaluationNode
	thread  *starlark.Thread
	via     *starlark.Value
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
			annotation.via = &kwarg[1]

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

	switch typedVal := (*a.via).(type) {
	case starlark.Callable:
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

	default:
		return nil, fmt.Errorf("Expected '%s' annotation keyword argument 'via'"+
			" to be function, but was %T", AnnotationReplace, typedVal)
	}
}

func (a ReplaceAnnotation) OrAdd() bool { return a.orAdd }
