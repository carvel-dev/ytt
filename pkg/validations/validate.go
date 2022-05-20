// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yttlibrary"
)

// NodeValidation represents a validation attached to a Node via an annotation.
type NodeValidation struct {
	rules    []rule
	kwargs   validationKwargs
	position *filepos.Position
}

// A rule contains a string description of what constitutes a valid value,
// and a function that asserts the rule against an actual value.
type rule struct {
	msg       string
	assertion starlark.Callable
}

// validationKwargs represent the optional keyword arguments and their values in a validation annotation.
type validationKwargs struct {
	when         *starlark.Callable
	whenNullSkip *bool // default: nil if kwarg is not set, True if value is Nullable
	minLength    int   // 0 len("") == 0, this always passes
	maxLength    *int
	min          starlark.Value
	max          starlark.Value
	notNull      bool
	// *int , default=nil possible_values={&1, &2, &-3, ..}
}

// Run takes a root Node, and threadName, and validates each Node in the tree.
//
// When a Node's value is invalid, the errors are collected and returned in an AssertCheck.
// Otherwise, returns empty AssertCheck and nil error.
func Run(n yamlmeta.Node, threadName string) AssertCheck {
	if n == nil {
		return AssertCheck{}
	}

	validationRunner := newValidationRunner(threadName)
	err := yamlmeta.Walk(n, validationRunner)
	if err != nil {
		return AssertCheck{}
	}

	return validationRunner.chk
}

type validationRunner struct {
	thread *starlark.Thread
	chk    AssertCheck
}

func newValidationRunner(threadName string) *validationRunner {
	return &validationRunner{thread: &starlark.Thread{Name: threadName}, chk: AssertCheck{[]error{}}}
}

// Visit if `node` has validations in its meta.
// Runs the validations, any violations from executing the assertions are collected.
//
// This visitor stores error(violations) in the validationRunner and returns nil.
func (a *validationRunner) Visit(node yamlmeta.Node) error {
	// get rules in node's meta
	validations := Get(node)

	if validations == nil {
		return nil
	}
	for _, v := range validations {
		// possible refactor to check validation kwargs prior to validating rules
		errs := v.Validate(node, a.thread)
		if errs != nil {
			a.chk.Violations = append(a.chk.Violations, errs...)
		}
	}

	return nil
}

// Validate runs the assertions in the rules with the node's value as arguments IF
// the ValidationKwargs conditional options pass.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (v NodeValidation) Validate(node yamlmeta.Node, thread *starlark.Thread) []error {
	key, nodeValue := v.newKeyAndStarlarkValue(node)

	executeRules, err := v.kwargs.shouldValidate(nodeValue, thread)
	if err != nil {
		return []error{err}
	}
	if !executeRules {
		return nil
	}

	var failures []error
	for _, r := range v.rules {
		result, err := starlark.Call(thread, r.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
		if err != nil {
			failures = append(failures, fmt.Errorf("%s (%s) requires %q; %s (by %s)", key, node.GetPosition().AsCompactString(), r.msg, err.Error(), v.position.AsCompactString()))
		} else {
			_, isNone := result.(starlark.NoneType)
			isTrue := bool(result.Truth())
			// in order to pass, the assertion must return Truthy value or None
			if !(isNone || isTrue) {
				failures = append(failures, fmt.Errorf("%s (%s) requires %q (by %s)", key, node.GetPosition().AsCompactString(), r.msg, v.position.AsCompactString()))
			}
		}
	}
	return failures
}

// DefaultNullSkipTrue sets the kwarg when_null_skip to true if not set explicitly.
func (v *NodeValidation) DefaultNullSkipTrue() {
	if v.kwargs.whenNullSkip == nil {
		t := true
		v.kwargs.whenNullSkip = &t
	}
}

// shouldValidate uses validationKwargs and the node's value to run checks on the value. If the value satisfies the checks,
// then the NodeValidation's rules should execute, otherwise the rules will be skipped.
func (v validationKwargs) shouldValidate(value starlark.Value, thread *starlark.Thread) (bool, error) {
	_, isNull := value.(starlark.NoneType)
	nullAllowed := !v.notNull
	whenNullSkip := v.whenNullSkip != nil && *v.whenNullSkip
	if isNull && nullAllowed && whenNullSkip {
		return false, nil
	}

	if v.when != nil {
		result, err := starlark.Call(thread, *v.when, starlark.Tuple{value}, []starlark.Tuple{})
		if err != nil {
			return false, err
		}

		isTrue := bool(result.Truth())
		return isNull || isTrue, nil
	}
	// if no kwargs then execute rules
	return true, nil
}

func (v validationKwargs) convertToRules() []rule {
	var rules []rule
	if minLen := v.minLength; minLen > 0 {
		a := yttlibrary.NewAssertMinLength(minLen)
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length greater or equal to %v", minLen),
			assertion: a,
		})
	}
	if v.maxLength != nil {
		a := yttlibrary.NewAssertMaxLength(*v.maxLength)
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length less than or equal to %v", *v.maxLength),
			assertion: a,
		})
	}
	if v.min != nil {
		a := yttlibrary.NewAssertMin(v.min)
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value greater or equal to %v", v.min),
			assertion: a,
		})
	}
	if v.max != nil {
		a := yttlibrary.NewAssertMax(v.max)
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value less than or equal to %v", v.max),
			assertion: a,
		})
	}
	if v.notNull {
		a := yttlibrary.NewAssertNotNull()
		rules = append(rules, rule{
			msg:       fmt.Sprintf("not null"),
			assertion: a,
		})
	}

	return rules
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

func (v NodeValidation) newKeyAndStarlarkValue(node yamlmeta.Node) (string, starlark.Value) {
	var key string
	var nodeValue starlark.Value
	switch typedNode := node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		panic(fmt.Sprintf("validation at %s - not supported on %s at %s", v.position.AsCompactString(), yamlmeta.TypeName(node), node.GetPosition().AsCompactString()))
	case *yamlmeta.Document:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.MapItem:
		key = fmt.Sprintf("%q", typedNode.Key)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	case *yamlmeta.ArrayItem:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode.Value).AsStarlarkValue()
	}

	return key, nodeValue
}
