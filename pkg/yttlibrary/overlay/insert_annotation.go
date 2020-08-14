// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
)

type InsertAnnotation struct {
	newItem template.EvaluationNode
	before  bool
	after   bool
}

func NewInsertAnnotation(newItem template.EvaluationNode) (InsertAnnotation, error) {
	annotation := InsertAnnotation{newItem: newItem}
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
		case "before":
			resultBool, err := tplcore.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return InsertAnnotation{}, err
			}
			annotation.before = resultBool

		case "after":
			resultBool, err := tplcore.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return InsertAnnotation{}, err
			}
			annotation.after = resultBool

		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationInsert, kwargName)
		}
	}

	return annotation, nil
}

func (a InsertAnnotation) IsBefore() bool { return a.before }
func (a InsertAnnotation) IsAfter() bool  { return a.after }
