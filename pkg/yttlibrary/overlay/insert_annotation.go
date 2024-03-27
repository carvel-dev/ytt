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

// Kwargs of overlay/insert
const (
	InsertAnnotationKwargBefore string = "before"
	InsertAnnotationKwargAfter  string = "after"
	InsertAnnotationKwargVia    string = "via"
)

type InsertAnnotation struct {
	newItem template.EvaluationNode
	before  bool
	after   bool
	via     *starlark.Callable
	thread  *starlark.Thread
}

// NewInsertAnnotation returns a new InsertAnnotation for the given node and with the given Starlark thread
func NewInsertAnnotation(newItem template.EvaluationNode, thread *starlark.Thread) (InsertAnnotation, error) {
	annotation := InsertAnnotation{
		newItem: newItem,
		thread:  thread,
	}
	anns := template.NewAnnotations(newItem)

	if !anns.Has(AnnotationInsert) {
		return annotation, fmt.Errorf(
			"Expected item to have '%s' annotation", AnnotationInsert)
	}

	kwargs := anns.Kwargs(AnnotationInsert)
	if len(kwargs) == 0 {
		return annotation, fmt.Errorf("Expected '%s' annotation to have "+
			"at least one keyword argument (before=..., after=...)", AnnotationInsert)
	}

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))

		switch kwargName {
		case InsertAnnotationKwargBefore:
			resultBool, err := tplcore.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return InsertAnnotation{}, err
			}
			annotation.before = resultBool

		case InsertAnnotationKwargAfter:
			resultBool, err := tplcore.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return InsertAnnotation{}, err
			}
			annotation.after = resultBool

		case InsertAnnotationKwargVia:
			via, ok := kwarg[1].(starlark.Callable)
			if !ok {
				return InsertAnnotation{}, fmt.Errorf(
					"Expected '%s' annotation keyword argument '%s' to be function, "+
						"but was %T", AnnotationInsert, kwargName, kwarg[1])
			}
			annotation.via = &via

		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationInsert, kwargName)
		}
	}

	return annotation, nil
}

func (a InsertAnnotation) IsBefore() bool { return a.before }
func (a InsertAnnotation) IsAfter() bool  { return a.after }

// Value returns the new value for the given, existing node. If `via` is not provided, the value of the existing
// node is returned, otherwise the result of `via`.
func (a InsertAnnotation) Value(existingNode template.EvaluationNode) (interface{}, error) {
	newNode := a.newItem.DeepCopyAsInterface().(template.EvaluationNode)
	if a.via == nil {
		return newNode.GetValues()[0], nil
	}

	var existingVal interface{}
	if existingNode != nil {
		existingVal = existingNode.DeepCopyAsInterface().(template.EvaluationNode).GetValues()[0]
	} else {
		existingVal = nil
	}

	viaArgs := starlark.Tuple{
		yamltemplate.NewGoValueWithYAML(existingVal).AsStarlarkValue(),
		yamltemplate.NewGoValueWithYAML(newNode.GetValues()[0]).AsStarlarkValue(),
	}

	result, err := starlark.Call(a.thread, *a.via, viaArgs, []starlark.Tuple{})
	if err != nil {
		return nil, err
	}

	return tplcore.NewStarlarkValue(result).AsGoValue()
}
