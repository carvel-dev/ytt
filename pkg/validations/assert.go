// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

// Declare @assert/... annotation names
const (
	AnnotationAssertValidate    template.AnnotationName = "assert/validate"
	ValidationKwargWhen         string                  = "when"
	ValidationKwargWhenNullSkip string                  = "when_null_skip"
)

// ProcessAssertValidateAnns checks Assert annotations on data values and stores them on a Node as Validations.
// Returns an error if any Assert annotations are malformed.
func ProcessAssertValidateAnns(rootNode yamlmeta.Node) error {
	if rootNode == nil {
		return nil
	}
	return yamlmeta.Walk(rootNode, &convertAssertAnnsToValidations{})
}

type convertAssertAnnsToValidations struct{}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks annotation, and stores the validation on Node's validations meta.
//
// This visitor returns and error if any assert annotation is not well-formed,
// otherwise, returns nil.
func (a *convertAssertAnnsToValidations) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if !nodeAnnotations.Has(AnnotationAssertValidate) {
		return nil
	}
	switch node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		return fmt.Errorf("Invalid @%s annotation - not supported on %s at %s", AnnotationAssertValidate, yamlmeta.TypeName(node), node.GetPosition().AsCompactString())
	default:
		validation, err := NewValidationFromValidationAnnotation(nodeAnnotations[AnnotationAssertValidate])
		if err != nil {
			return fmt.Errorf("Invalid @%s annotation - %s", AnnotationAssertValidate, err.Error())
		}
		// store rules in node's validations meta without overriding any existing rules
		Add(node, []NodeValidation{*validation})
	}

	return nil
}

// NewValidationFromValidationAnnotation creates a NodeValidation from the values provided in a validation annotation.
// If any value in the annotation is not well-formed, it returns an error.
func NewValidationFromValidationAnnotation(annotation template.NodeAnnotation) (*NodeValidation, error) {
	var rules []rule

	if len(annotation.Args) == 0 {
		return nil, fmt.Errorf("expected annotation to have 2-tuple as argument(s), but found no arguments (by %s)", annotation.Position.AsCompactString())
	}
	for _, arg := range annotation.Args {
		ruleTuple, ok := arg.(starlark.Tuple)
		if !ok {
			return nil, fmt.Errorf("expected annotation to have 2-tuple as argument(s), but found: %s (by %s)", arg.String(), annotation.Position.AsCompactString())
		}
		if len(ruleTuple) != 2 {
			return nil, fmt.Errorf("expected 2-tuple, but found tuple with length %v (by %s)", len(ruleTuple), annotation.Position.AsCompactString())
		}
		message, ok := ruleTuple[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("expected first item in the 2-tuple to be a string describing a valid value, but was %s (at %s)", ruleTuple[0].Type(), annotation.Position.AsCompactString())
		}
		lambda, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			return nil, fmt.Errorf("expected second item in the 2-tuple to be an assertion function, but was %s (at %s)", ruleTuple[1].Type(), annotation.Position.AsCompactString())
		}
		rules = append(rules, rule{
			msg:       message.GoString(),
			assertion: lambda,
		})
	}
	kwargs, err := newValidationKwargs(annotation.Kwargs, annotation.Position)
	if err != nil {
		return nil, err
	}

	return &NodeValidation{rules, kwargs, annotation.Position}, nil
}

func newValidationKwargs(kwargs []starlark.Tuple, annPos *filepos.Position) (validationKwargs, error) {
	var processedKwargs validationKwargs
	for _, value := range kwargs {
		kwargName := string(value[0].(starlark.String))
		switch kwargName {
		case ValidationKwargWhen:
			lambda, ok := value[1].(starlark.Callable)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a function, but was %s (at %s)", ValidationKwargWhen, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.when = &lambda
		case ValidationKwargWhenNullSkip:
			b, ok := value[1].(starlark.Bool)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a boolean, but was %s (at %s)", ValidationKwargWhenNullSkip, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.whenNullSkip = bool(b)
		default:
			return validationKwargs{}, fmt.Errorf("unknown keyword argument %q (at %s)", kwargName, annPos.AsCompactString())
		}
	}
	return processedKwargs, nil
}
