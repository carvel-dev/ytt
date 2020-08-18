// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/ytt/pkg/template"
)

type MatchChildDefaultsAnnotation struct {
	expects MatchAnnotationExpectsKwarg
}

func NewEmptyMatchChildDefaultsAnnotation() MatchChildDefaultsAnnotation {
	return MatchChildDefaultsAnnotation{
		expects: MatchAnnotationExpectsKwarg{},
	}
}

func NewMatchChildDefaultsAnnotation(node template.EvaluationNode,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) (MatchChildDefaultsAnnotation, error) {

	annotation := MatchChildDefaultsAnnotation{
		// TODO do we need to propagate thread?
		expects: MatchAnnotationExpectsKwarg{},
	}
	kwargs := template.NewAnnotations(node).Kwargs(AnnotationMatchChildDefaults)

	for _, kwarg := range kwargs {
		kwargName := string(kwarg[0].(starlark.String))
		switch kwargName {
		case MatchAnnotationKwargExpects:
			annotation.expects.expects = &kwarg[1]
		case MatchAnnotationKwargMissingOK:
			annotation.expects.missingOK = &kwarg[1]
		case MatchAnnotationKwargWhen:
			annotation.expects.when = &kwarg[1]
		default:
			return annotation, fmt.Errorf(
				"Unknown '%s' annotation keyword argument '%s'", AnnotationMatchChildDefaults, kwargName)
		}
	}

	annotation.expects.FillInDefaults(parentMatchChildDefaults)

	return annotation, nil
}
