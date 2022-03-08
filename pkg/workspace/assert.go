// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
)

// Declare @assert/... annotation names
const (
	AnnotationAssertValidate template.AnnotationName = "assert/validate"
)

// ProcessAndRunValidations takes a root Node and traverses the tree checking for assert annotations.
// Validations are processed and executed using the value of the annotated node as the parameter to the assertions.
//
// When the assertions have violations, the errors are collected and returned.
// Otherwise, returns nil.
func ProcessAndRunValidations(n yamlmeta.Node, checker *assertChecker) error {
	if n == nil {
		return nil
	}

	err := yamlmeta.Walk(n, checker)
	if err != nil {
		return err
	}

	if checker.hasViolations() {
		var compiledViolations string
		for _, err := range checker.violations {
			compiledViolations = compiledViolations + err.Error() + "\n"
		}
		return fmt.Errorf("%s", compiledViolations)
	}

	return nil
}

type assertChecker struct {
	thread     *starlark.Thread
	violations []error
}

func newAssertChecker(threadName string, _ *TemplateLoader) *assertChecker {
	return &assertChecker{thread: &starlark.Thread{Name: threadName}, violations: []error{}}
}

func (a *assertChecker) hasViolations() bool {
	return len(a.violations) > 0
}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks, validates, and runs the validation Rules, any violations from running the assertions are collected.
//
// This visitor returns and error if any assertion is not well-formed,
// otherwise, returns nil.
func (a *assertChecker) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if nodeAnnotations.Has(AnnotationAssertValidate) {
		switch node.(type) {
		case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
			return fmt.Errorf("Invalid @%s annotation - not supported on %s at %s", AnnotationAssertValidate, yamlmeta.TypeName(node), node.GetPosition().AsCompactString())
		}
		rules, syntaxErr := newRulesFromAssertValidateAnnotation(nodeAnnotations[AnnotationAssertValidate], node)
		if syntaxErr != nil {
			return syntaxErr
		}
		for _, rule := range rules {
			err := rule.Validate(node, a.thread)
			if err != nil {
				a.violations = append(a.violations, err)
			}
		}
	}

	return nil
}

func newRulesFromAssertValidateAnnotation(annotation template.NodeAnnotation, n yamlmeta.Node) ([]Rule, error) {
	var validations []Rule
	validationPosition := createAssertAnnotationPosition(annotation.Position, n)

	if len(annotation.Kwargs) != 0 {
		return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found keyword argument (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, validationPosition.AsCompactString())
	}
	if len(annotation.Args) == 0 {
		return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found no arguments (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, validationPosition.AsCompactString())
	}
	for _, arg := range annotation.Args {
		ruleTuple, ok := arg.(starlark.Tuple)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found: %s (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, arg.String(), validationPosition.AsCompactString())
		}
		if len(ruleTuple) != 2 {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s 2-tuple, but found tuple with length %v (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, len(ruleTuple), validationPosition.AsCompactString())
		}
		message, ok := ruleTuple[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have string describing a valid value as the first item in the 2-tuple, but found type: %s (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, ruleTuple[0].Type(), validationPosition.AsCompactString())
		}
		lambda, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have an assertion function as the second item in the 2-tuple, but found type: %s (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, ruleTuple[1].Type(), validationPosition.AsCompactString())
		}
		validations = append(validations, Rule{
			msg:       message.String(),
			assertion: lambda,
			position:  validationPosition,
		})
	}

	return validations, nil
}

// A Rule represents an argument to an @assert/validate annotation;
// it contains a string description of what constitutes a valid value,
// and a function that asserts the rule against an actual value.
// One @assert/validate annotation can have multiple Rules.
type Rule struct {
	msg       string
	assertion starlark.Callable
	position  *filepos.Position
}

// Validate runs the assertion from the Rule with the node's value as arguments.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (r Rule) Validate(node yamlmeta.Node, thread *starlark.Thread) error {
	var nodeValue starlark.Value
	if values := node.GetValues(); len(values) == 1 {
		nodeValue = yamltemplate.NewGoValueWithYAML(values[0]).AsStarlarkValue()
	} else {
		panic(fmt.Sprintf("%v number of values to validate with @%s annotation at %s", len(values), AnnotationAssertValidate, r.position.AsCompactString()))
	}

	result, err := starlark.Call(thread, r.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
	if err != nil {
		return fmt.Errorf("%s requires a valid value: %s; %s (by %s)", node.GetPosition().AsCompactString(), r.msg, err.Error(), r.position.AsCompactString())
	}

	// in order to pass, the assertion must return True or None
	if _, ok := result.(starlark.NoneType); !ok {
		if !result.Truth() {
			node.GetAnnotations()
			return fmt.Errorf("%s requires a valid value: %s (by %s)", node.GetPosition().AsCompactString(), r.msg, r.position.AsCompactString())
		}
	}

	return nil
}

func createAssertAnnotationPosition(annPosition *filepos.Position, node yamlmeta.Node) *filepos.Position {
	filePosComments := make([]filepos.Meta, 0, len(node.GetComments()))
	for _, c := range node.GetComments() {
		filePosComments = append(filePosComments, filepos.Meta{Data: c.Data, Position: c.Position.DeepCopy()})
	}

	return filepos.PopulateAnnotationPositionFromNode(annPosition, node.GetPosition(), filePosComments)

}
