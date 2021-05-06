// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
)

type AssertAnnotation struct {
	newNode template.EvaluationNode
	thread  *starlark.Thread
	via     *starlark.Value
}

func NewAssertAnnotation(newNode template.EvaluationNode, thread *starlark.Thread) (AssertAnnotation, error) {
	annotation := AssertAnnotation{
		newNode: newNode,
		thread:  thread,
	}
	kwargs := template.NewAnnotations(newNode).Kwargs(AnnotationAssert)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case "via":
			annotation.via = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationAssert, kwargName)
		}
	}

	return annotation, nil
}

func (a AssertAnnotation) Check(existingNode template.EvaluationNode) error {
	// Make sure original nodes are not affected in any way
	existingNode = existingNode.DeepCopyAsInterface().(template.EvaluationNode)
	newNode := a.newNode.DeepCopyAsInterface().(template.EvaluationNode)

	// TODO currently assumes that we can always get at least one value
	existingVal := existingNode.GetValues()[0]
	newVal := newNode.GetValues()[0]

	if a.via == nil {
		actualObj := yamlmeta.NewASTFromInterface(existingVal)
		expectedObj := yamlmeta.NewASTFromInterface(newVal)

		// TODO use generic equal function from our library?
		equal, desc := Comparison{}.Compare(actualObj, expectedObj)
		if !equal {
			return fmt.Errorf("Expected objects to equal, but did not: %s", desc)
		}
		return nil
	}

	switch typedVal := (*a.via).(type) {
	case starlark.Callable:
		viaArgs := starlark.Tuple{
			yamltemplate.NewGoValueWithYAML(existingVal).AsStarlarkValue(),
			yamltemplate.NewGoValueWithYAML(newVal).AsStarlarkValue(),
		}

		result, err := starlark.Call(a.thread, *a.via, viaArgs, []starlark.Tuple{})
		if err != nil {
			return err
		}

		switch typedResult := result.(type) {
		case nil, starlark.NoneType:
			// Assume if via didnt error then it's successful
			return nil

		case starlark.Bool:
			if !bool(typedResult) {
				return fmt.Errorf("Expected via invocation to return true, but was false")
			}
			return nil

		default:
			result, err := tplcore.NewStarlarkValue(result).AsGoValue()
			if err != nil {
				return err
			}

			// Extract result tuple(bool, string) to determine success
			if typedResult, ok := result.([]interface{}); ok {
				if len(typedResult) == 2 {
					resultSuccess, ok1 := typedResult[0].(bool)
					resultMsg, ok2 := typedResult[1].(string)
					if ok1 && ok2 {
						if !resultSuccess {
							return fmt.Errorf("Expected via invocation to return true, "+
								"but was false with message: %s", resultMsg)
						}
						return nil
					}
				}
			}

			return fmt.Errorf("Expected via invocation to return NoneType, " +
				"Bool or Tuple(Bool,String), but returned neither of those")
		}

	default:
		return fmt.Errorf("Expected '%s' annotation keyword argument 'via'"+
			" to be function, but was %T", AnnotationAssert, typedVal)
	}
}
