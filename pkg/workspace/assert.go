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

const (
	AnnotationAssertValidate template.AnnotationName = "assert/validate"
)

// ProcessAndRunAssertions takes a root Node and traverses the tree checking for assert annotations.
// Assertions are processed and executed using the value of the annotated node as the parameter to the assertions.
//
// When the assertions have violations, the errors are collected and returned, prefixed by the errMsg context string.
// Otherwise, returns nil.
func ProcessAndRunAssertions(n yamlmeta.Node, errMsg string) error {
	if n == nil {
		return nil
	}

	assertionChecker := newAssertChecker("assertions-process-and-run")
	err := yamlmeta.Walk(n, assertionChecker)
	if err != nil {
		return err
	}

	if assertionChecker.hasViolations() {
		var compiledViolations string
		for _, err := range assertionChecker.violations {
			compiledViolations = compiledViolations + err.Error() + "\n"
		}
		return fmt.Errorf("%v\n%v", errMsg, compiledViolations)
	}

	return nil
}

type assertChecker struct {
	thread *starlark.Thread
	violations []error
}

func newAssertChecker(threadName string) *assertChecker {
	return &assertChecker{thread: &starlark.Thread{Name: threadName}, violations: []error{}}
}

func (a *assertChecker) hasViolations() bool {
	if len(a.violations) > 0 {
		return true
	}
	return false
}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks, validates, and runs the assertions, any violations from running the assertions are collected and stored.
//
// This visitor returns and error if any assertion is not well-formed,
// otherwise, returns nil.
func (a *assertChecker) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if nodeAnnotations.Has(AnnotationAssertValidate) {
		validations, syntaxErr := newValidationsFromAssertValidateAnnotation(nodeAnnotations[AnnotationAssertValidate], node, a.thread)
		if syntaxErr != nil {
			return syntaxErr
		}
		for _, v := range validations {
			err := v.Validate(node)
			if err != nil {
				a.violations = append(a.violations, err)
			}
		}
	}

	return nil
}

func newValidationsFromAssertValidateAnnotation(annotation template.NodeAnnotation, n yamlmeta.Node, thread *starlark.Thread) ([]Validation, error) {
	var validations []Validation
	validationPosition := createAssertAnnotationPosition(annotation.Position, n)

	if len(annotation.Kwargs) != 0 {
		return nil, fmt.Errorf("Expected @%v to have validation 2-tuple as argument(s), but found keyword argument (by %v)", AnnotationAssertValidate, validationPosition.AsCompactString())
	}
	if len(annotation.Args) == 0 {
		return nil, fmt.Errorf("Expected @%v to have validation 2-tuple as argument(s), but found no arguments (by %v)", AnnotationAssertValidate, validationPosition.AsCompactString())
	}
	for _, arg := range annotation.Args {
		validationTuple, ok := arg.(starlark.Tuple)
		if !ok {
			return nil, fmt.Errorf("Expected @%v to have validation 2-tuple as argument(s), but found: %v (by %v)", AnnotationAssertValidate, arg.String(), validationPosition.AsCompactString())
		}
		if len(validationTuple) != 2 {
			return nil, fmt.Errorf("Expected @%v 2-tuple, but found tuple with length %v (by %v)", AnnotationAssertValidate, len(validationTuple), validationPosition.AsCompactString())
		}
		message, ok := validationTuple[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("Expected @%v to have descriptive string as the first item in the 2-tuple, but found type: %v (by %v)", AnnotationAssertValidate, validationTuple[0].Type(), validationPosition.AsCompactString())
		}
		lambda, ok := validationTuple[1].(starlark.Callable)
		if !ok {
			return nil, fmt.Errorf("Expected @%v to have a validation function as the second item in the 2-tuple, but found type: %v (by %v)", AnnotationAssertValidate, validationTuple[1].Type(), validationPosition.AsCompactString())
		}
		validations = append(validations, Validation{
			thread:   thread,
			msg:      message.String(),
			f:        lambda,
			position: validationPosition,
		})
	}

	return validations, nil
}

// Validation is an assertion from an @assert/validate annotation.
// One @assert/validate annotation can have multiple Validations.
type Validation struct {
	thread   *starlark.Thread
	msg      string
	f        starlark.Callable
	position *filepos.Position
}

// Validate runs the function from the Validation with the node's value as arguments.
//
// Returns an error if the function returns an error, or if the resulting value from the function is falsy.
// Otherwise, returns nil.
func (v Validation) Validate(node yamlmeta.Node) error {
	var nodeValue starlark.Value
	switch values := node.GetValues(); len(values) {
	case 0: // run validation with no arguments
	case 1:
		nodeValue = yamltemplate.NewGoValueWithYAML(values[0]).AsStarlarkValue()
	default:
		nodeValue = yamltemplate.NewGoValueWithYAML(values).AsStarlarkValue()
	}

	result, err := starlark.Call(v.thread, v.f, starlark.Tuple{nodeValue}, []starlark.Tuple{})
	if err != nil {
		return fmt.Errorf("%v requires a valid value: %v; %v (by %v)", node.GetPosition().AsCompactString(), v.msg, err.Error(), v.position.AsCompactString())
	}
	// From proposal: "if all rules pass (either returns True or None), then the value is valid."
	// I believe a None result would return an error
	if !result.Truth() {
		node.GetAnnotations()
		return fmt.Errorf("%v requires a valid value: %v (by %v)", node.GetPosition().AsCompactString(), v.msg, v.position.AsCompactString())
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
