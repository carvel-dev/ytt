// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yttlibrary"
)

// NodeValidation represents a validationRun attached to a Node via an annotation.
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
	priority  int  // how early to run this rule. 0 = order it appears; more positive: earlier, more negative: later.
	isFatal   bool // whether not satisfying this rule prevents others rules from running.
}

// byPriority is a sort.Interface that orders rules based on priority
type byPriority []rule

func (r byPriority) Len() int {
	return len(r)
}
func (r byPriority) Swap(idx, jdx int) {
	r[idx], r[jdx] = r[jdx], r[idx]
}

// Less reports whether the rule at "idx" should run _later_ than the rule at "jdx".
func (r byPriority) Less(idx, jdx int) bool {
	return r[idx].priority > r[jdx].priority
}

// validationKwargs represent the optional keyword arguments and their values in a validationRun annotation.
type validationKwargs struct {
	when         starlark.Callable
	whenNullSkip *bool         // default: nil if kwarg is not set, True if value is Nullable
	minLength    *starlark.Int // 0 len("") == 0, this always passes
	maxLength    *starlark.Int
	min          starlark.Value
	max          starlark.Value
	notNull      bool
	oneNotNull   starlark.Value // valid values are either starlark.Sequence or starlark.Bool
	oneOf        starlark.Sequence
}

// Run takes a root Node, and threadName, and validates each Node in the tree.
//
// When a Node's value is invalid, the errors are collected and returned in a Check.
// Otherwise, returns empty Check and nil error.
func Run(node yamlmeta.Node, threadName string) Check {
	if node == nil {
		return Check{}
	}

	validation := newValidationRun(threadName)
	err := yamlmeta.Walk(node, validation)
	if err != nil {
		return Check{}
	}

	return validation.chk
}

type validationRun struct {
	thread *starlark.Thread
	chk    Check
}

func newValidationRun(threadName string) *validationRun {
	return &validationRun{thread: &starlark.Thread{Name: threadName}, chk: Check{[]error{}}}
}

// Visit if `node` has validations in its meta.
// Runs the validations, any violations from executing the assertions are collected.
//
// This visitor stores error(violations) in the validationRun and returns nil.
func (a *validationRun) Visit(node yamlmeta.Node) error {
	// get rules in node's meta
	validations := Get(node)

	if validations == nil {
		return nil
	}
	for _, v := range validations {
		// possible refactor to check validationRun kwargs prior to validating rules
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

	sort.Sort(byPriority(v.rules))
	var failures []error
	for _, r := range v.rules {
		result, err := starlark.Call(thread, r.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
		if err != nil {
			failures = append(failures, fmt.Errorf("%s (%s) requires %q; %s (by %s)", key, node.GetPosition().AsCompactString(), r.msg, err.Error(), v.position.AsCompactString()))
			if r.isFatal {
				break
			}
		} else {
			if !(result == starlark.True) {
				failures = append(failures, fmt.Errorf("%s (%s) requires %q (by %s)", key, node.GetPosition().AsCompactString(), r.msg, v.position.AsCompactString()))
				if r.isFatal {
					break
				}
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
	// avoid nil pointer errors: handle `when_null_skip=`, first
	_, valueIsNull := value.(starlark.NoneType)
	nullIsAllowed := !v.notNull
	whenNullSkip := v.whenNullSkip != nil && *v.whenNullSkip
	if valueIsNull && nullIsAllowed && whenNullSkip {
		return false, nil
	}

	if v.when != nil && !reflect.ValueOf(v.when).IsNil() {
		result, err := starlark.Call(thread, v.when, starlark.Tuple{value}, []starlark.Tuple{})
		if err != nil {
			return false, err
		}

		resultBool, isBool := result.(starlark.Bool)
		if !isBool {
			return false, fmt.Errorf("want when= to be bool, got %s", result.Type())
		}

		return bool(resultBool), nil
	}

	return true, nil
}

func (v validationKwargs) asRules() []rule {
	var rules []rule

	if v.minLength != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length greater or equal to %v", *v.minLength),
			assertion: yttlibrary.NewAssertMinLen(*v.minLength).CheckFunc(),
		})
	}
	if v.maxLength != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length less than or equal to %v", *v.maxLength),
			assertion: yttlibrary.NewAssertMaxLen(*v.maxLength).CheckFunc(),
		})
	}
	if v.min != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value greater or equal to %v", v.min),
			assertion: yttlibrary.NewAssertMin(v.min).CheckFunc(),
		})
	}
	if v.max != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value less than or equal to %v", v.max),
			assertion: yttlibrary.NewAssertMax(v.max).CheckFunc(),
		})
	}
	if v.notNull {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("not null"),
			assertion: yttlibrary.NewAssertNotNull().CheckFunc(),
			isFatal:   true,
			priority:  100,
		})
	}
	if v.oneNotNull != nil {
		var assertion *yttlibrary.Assertion

		switch oneNotNull := v.oneNotNull.(type) {
		case starlark.Bool:
			if oneNotNull {
				assertion = yttlibrary.NewAssertOneNotNull(nil)
			} else {
				// should have been caught when args were parsed
				panic("one_not_null= cannot be False")
			}
		case starlark.Sequence:
			assertion = yttlibrary.NewAssertOneNotNull(oneNotNull)
		default:
			// should have been caught when args were parsed
			panic(fmt.Sprintf("Unexpected type \"%s\" for one_not_null=", v.oneNotNull.Type()))
		}

		rules = append(rules, rule{
			msg:       fmt.Sprintf("exactly one child not null"),
			assertion: assertion.CheckFunc(),
		})
	}
	if v.oneOf != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("one of"),
			assertion: yttlibrary.NewAssertOneOf(v.oneOf).CheckFunc(),
		})
	}

	return rules
}

// Check holds the resulting violations from executing Validations on a node.
type Check struct {
	Violations []error
}

// Error generates the error message composed of the total set of Check.Violations.
func (c Check) Error() string {
	if !c.HasViolations() {
		return ""
	}

	msg := ""
	for _, err := range c.Violations {
		msg = msg + "- " + err.Error() + "\n"
	}
	return msg
}

// HasViolations indicates whether this Check contains any violations.
func (c *Check) HasViolations() bool {
	return len(c.Violations) > 0
}

func (v NodeValidation) newKeyAndStarlarkValue(node yamlmeta.Node) (string, starlark.Value) {
	var key string
	var nodeValue starlark.Value
	switch typedNode := node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		panic(fmt.Sprintf("validationRun at %s - not supported on %s at %s", v.position.AsCompactString(), yamlmeta.TypeName(node), node.GetPosition().AsCompactString()))
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
