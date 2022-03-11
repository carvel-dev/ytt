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

// ProcessAndRunValidations takes a root Node, and threadName, and traverses the tree checking for assert annotations.
// Validations are processed and executed using the value of the annotated node as the parameter to the assertions.
//
// When the assertions have violations, the errors are collected and returned in an AssertCheck.
// Otherwise, returns empty AssertCheck and nil.
func ProcessAndRunValidations(n yamlmeta.Node, threadName string) (AssertCheck, error) {
	if n == nil {
		return AssertCheck{}, nil
	}

	assertionChecker := newAssertChecker(threadName)
	err := yamlmeta.Walk(n, assertionChecker)
	if err != nil {
		return AssertCheck{}, err
	}

	return assertionChecker.AssertCheck, nil
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

type assertChecker struct {
	thread *starlark.Thread
	AssertCheck
}

func newAssertChecker(threadName string) *assertChecker {
	return &assertChecker{thread: &starlark.Thread{Name: threadName}, AssertCheck: AssertCheck{[]error{}}}
}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks, validates, and runs the validation Rules, any violations from running the assertions are collected.
//
// This visitor returns and error if any assertion is not well-formed,
// otherwise, returns nil.
func (a *assertChecker) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if !nodeAnnotations.Has(AnnotationAssertValidate) {
		return nil
	}
	switch node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		return fmt.Errorf("Invalid @%s annotation - not supported on %s at %s", AnnotationAssertValidate, yamlmeta.TypeName(node), node.GetPosition().AsCompactString())
	default:
		rules, syntaxErr := newRulesFromAssertValidateAnnotation(nodeAnnotations[AnnotationAssertValidate], node)
		if syntaxErr != nil {
			return syntaxErr
		}
		for _, rule := range rules {
			err := rule.Validate(node, a.thread)
			if err != nil {
				a.AssertCheck.Violations = append(a.AssertCheck.Violations, err)
			}
		}
	}

	return nil
}

func newRulesFromAssertValidateAnnotation(annotation template.NodeAnnotation, n yamlmeta.Node) ([]Rule, error) {
	var rules []Rule
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
			return nil, fmt.Errorf("Invalid @%s annotation - expected first item in the 2-tuple to be a string describing a valid value, but was %s (at %s)", AnnotationAssertValidate, ruleTuple[0].Type(), validationPosition.AsCompactString())
		}
		lambda, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			return nil, fmt.Errorf("Invalid @%s annotation - expected second item in the 2-tuple to be an assertion function, but was %s (at %s)", AnnotationAssertValidate, ruleTuple[1].Type(), validationPosition.AsCompactString())
		}
		rules = append(rules, Rule{
			msg:       message.String(),
			assertion: lambda,
			position:  validationPosition,
		})
	}

	return rules, nil
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
	var key string
	var nodeValue starlark.Value
	switch typedNode := node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		panic(fmt.Sprintf("@%s annotation at %s - not supported on %s at %s", AnnotationAssertValidate, r.position.AsCompactString(), yamlmeta.TypeName(node), node.GetPosition().AsCompactString()))
	case *yamlmeta.MapItem:
		key = fmt.Sprintf("%s", typedNode.Key)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.ArrayItem:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.Document:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	}

	result, err := starlark.Call(thread, r.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
	if err != nil {
		return fmt.Errorf("%s (%s) requires %s; %s (by %s)", key, node.GetPosition().AsCompactString(), r.msg, err.Error(), r.position.AsCompactString())
	}

	// in order to pass, the assertion must return True or None
	if _, ok := result.(starlark.NoneType); !ok {
		if !result.Truth() {
			return fmt.Errorf("%s (%s) requires %s (by %s)", key, node.GetPosition().AsCompactString(), r.msg, r.position.AsCompactString())
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
