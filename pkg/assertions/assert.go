// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assertions

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

// Declare @assert/... annotation names
const (
	AnnotationAssertValidate template.AnnotationName = "assert/validate"
)

// ProcessAndRunValidations takes a root Node, and threadName, and validates each Node in the tree.
// Assert annotations are stored on the Node as Validations, which are then executed using the
// value of the annotated node as the parameter to the assertions.
//
// When a Node's value is invalid, the errors are collected and returned in an AssertCheck.
// Otherwise, returns empty AssertCheck and nil error.
func ProcessAndRunValidations(n yamlmeta.Node, threadName string) (AssertCheck, error) {
	if n == nil {
		return AssertCheck{}, nil
	}
	err := yamlmeta.Walk(n, &convertAssertAnnsToValidations{})
	if err != nil {
		return AssertCheck{}, err
	}

	validationRunner := newValidationRunner(threadName)
	err = yamlmeta.Walk(n, validationRunner)
	if err != nil {
		return AssertCheck{}, err
	}

	return validationRunner.chk, nil
}

// AssertCheck holds the resulting violations from executing Validations on a node.
type AssertCheck struct {
	Violations []error
}

// Error generates the error message composed of the total set of AssertCheck.Violations.
func (ac AssertCheck) Error() string {
	if !ac.HasViolations() {
		return ""
	}

	msg := ""
	for _, err := range ac.Violations {
		msg = msg + "- " + err.Error() + "\n"
	}
	return msg
}

// HasViolations indicates whether this AssertCheck contains any violations.
func (ac *AssertCheck) HasViolations() bool {
	return len(ac.Violations) > 0
}

type convertAssertAnnsToValidations struct{}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks annotation, and stores validation Rules on Node's validations meta.
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
		rules, syntaxErr := newRulesFromAssertValidateAnnotation(nodeAnnotations[AnnotationAssertValidate])
		if syntaxErr != nil {
			return syntaxErr
		}
		// store rules in node's validations meta without overriding any existing rules
		AddValidations(node, rules)
	}

	return nil
}

func newRulesFromAssertValidateAnnotation(annotation template.NodeAnnotation) ([]Rule, error) {
	var rules []Rule

	if len(annotation.Kwargs) != 0 {
		return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found keyword argument (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, annotation.Position.AsCompactString())
	}
	if len(annotation.Args) == 0 {
		return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found no arguments (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, annotation.Position.AsCompactString())
	}
	for _, arg := range annotation.Args {
		ruleTuple, ok := arg.(starlark.Tuple)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s to have 2-tuple as argument(s), but found: %s (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, arg.String(), annotation.Position.AsCompactString())
		}
		if len(ruleTuple) != 2 {
			return nil, fmt.Errorf("Invalid @%s annotation - expected @%s 2-tuple, but found tuple with length %v (by %s)", AnnotationAssertValidate, AnnotationAssertValidate, len(ruleTuple), annotation.Position.AsCompactString())
		}
		message, ok := ruleTuple[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected first item in the 2-tuple to be a string describing a valid value, but was %s (at %s)", AnnotationAssertValidate, ruleTuple[0].Type(), annotation.Position.AsCompactString())
		}
		lambda, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected second item in the 2-tuple to be an assertion function, but was %s (at %s)", AnnotationAssertValidate, ruleTuple[1].Type(), annotation.Position.AsCompactString())
		}
		rules = append(rules, Rule{
			Msg:       message.GoString(),
			Assertion: lambda,
			Position:  annotation.Position,
		})
	}

	return rules, nil
}

type validationRunner struct {
	thread *starlark.Thread
	chk    AssertCheck
}

func newValidationRunner(threadName string) *validationRunner {
	return &validationRunner{thread: &starlark.Thread{Name: threadName}, chk: AssertCheck{[]error{}}}
}

// Visit if `node` is has validations in its meta.
// Runs the validation Rules, any violations from running the assertions are collected.
//
// This visitor stores error(violations) in the validationRunner and returns nil.
func (a *validationRunner) Visit(node yamlmeta.Node) error {
	// get rules in node's meta
	rules := GetValidations(node)
	if rules == nil {
		return nil
	}
	for _, rule := range rules {
		err := rule.Validate(node, a.thread)
		if err != nil {
			a.chk.Violations = append(a.chk.Violations, err)
		}
	}

	return nil
}
