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
	msg        string
	assertion  starlark.Callable
	priority   int  // how early to run this rule. 0 = order it appears; more positive: earlier, more negative: later.
	isCritical bool // whether not satisfying this rule prevents others rules from running.
}

// byPriority sorts (a copy) of "rules" by priority in descending order (i.e. the order in which the rules should run)
func byPriority(rules []rule) []rule {
	sorted := make([]rule, len(rules))
	copy(sorted, rules)
	sort.SliceStable(sorted, func(idx, jdx int) bool {
		return sorted[idx].priority > sorted[jdx].priority
	})
	return sorted
}

// validationKwargs represent the optional keyword arguments and their values in a validationRun annotation.
type validationKwargs struct {
	when       starlark.Callable
	minLength  *starlark.Int // 0 len("") == 0, this always passes
	maxLength  *starlark.Int
	min        starlark.Value
	max        starlark.Value
	notNull    bool
	oneNotNull starlark.Value // valid values are either starlark.Sequence or starlark.Bool
	oneOf      starlark.Sequence
}

// Run takes a root Node, and threadName, and validates each Node in the tree.
//
// When a Node's value is invalid, the errors are collected and returned in a Check.
// Otherwise, returns empty Check and nil error.
func Run(node yamlmeta.Node, threadName string) (Check, error) {
	if node == nil {
		return Check{}, nil
	}

	validation := newValidationRun(threadName, node)
	err := yamlmeta.WalkWithParent(node, nil, validation)
	if err != nil {
		return Check{}, err
	}

	return validation.chk, nil
}

type validationRun struct {
	thread *starlark.Thread
	chk    Check
	root   yamlmeta.Node
}

func newValidationRun(threadName string, root yamlmeta.Node) *validationRun {
	return &validationRun{thread: &starlark.Thread{Name: threadName}, chk: Check{[]error{}}, root: root}
}

// VisitWithParent if `node` has validations in its meta.
// Runs those validations, collecting any violations
//
// This visitor stores error(violations) in the validationRun and returns nil.
func (a *validationRun) VisitWithParent(node yamlmeta.Node, parent yamlmeta.Node) error {
	// get rules in node's meta
	validations := Get(node)

	if validations == nil {
		return nil
	}
	for _, v := range validations {
		// possible refactor to check validationRun kwargs prior to validating rules
		violations, err := v.Validate(node, parent, a.root, a.thread)
		if err != nil {
			return err
		}
		if violations != nil {
			a.chk.Violations = append(a.chk.Violations, violations...)
		}
	}

	return nil
}

// Validate runs the assertions in the rules with the node's value as arguments IF
// the ValidationKwargs conditional options pass.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (v NodeValidation) Validate(node yamlmeta.Node, parent yamlmeta.Node, root yamlmeta.Node, thread *starlark.Thread) ([]error, error) {
	nodeKey, nodeValue := v.newKeyAndStarlarkValue(node)
	_, parentValue := v.newKeyAndStarlarkValue(parent)
	_, rootValue := v.newKeyAndStarlarkValue(root)

	executeRules, err := v.kwargs.shouldValidate(nodeValue, parentValue, thread, rootValue)
	if err != nil {
		return nil, fmt.Errorf("Validating %s: %s", nodeKey, err)
	}
	if !executeRules {
		return nil, nil
	}

	var violations []error
	for _, rul := range byPriority(v.rules) {
		result, err := starlark.Call(thread, rul.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
		if err != nil {
			violations = append(violations, fmt.Errorf("%s (%s) requires %q; %s (by %s)", nodeKey, node.GetPosition().AsCompactString(), rul.msg, err.Error(), v.position.AsCompactString()))
			if rul.isCritical {
				break
			}
		} else {
			if !(result == starlark.True) {
				violations = append(violations, fmt.Errorf("%s (%s) requires %q (by %s)", nodeKey, node.GetPosition().AsCompactString(), rul.msg, v.position.AsCompactString()))
				if rul.isCritical {
					break
				}
			}
		}
	}
	return violations, nil
}

// shouldValidate uses validationKwargs and the node's value to run checks on the value. If the value satisfies the checks,
// then the NodeValidation's rules should execute, otherwise the rules will be skipped.
func (v validationKwargs) shouldValidate(value starlark.Value, parent starlark.Value, thread *starlark.Thread, root starlark.Value) (bool, error) {
	_, valueIsNull := value.(starlark.NoneType)
	if valueIsNull && !v.notNull {
		return false, nil
	}

	if v.when != nil && !reflect.ValueOf(v.when).IsNil() {
		kwargs, err := v.populateArgs(value, parent, root)
		if err != nil {
			return false, fmt.Errorf("Failed to evaluate when=: %s", err)
		}

		result, err := starlark.Call(thread, v.when, starlark.Tuple{}, kwargs)
		if err != nil {
			return false, fmt.Errorf("Failure evaluating when=: %s", err)
		}

		resultBool, isBool := result.(starlark.Bool)
		if !isBool {
			return false, fmt.Errorf("want when= to be bool, got %s", result.Type())
		}

		return bool(resultBool), nil
	}

	return true, nil
}

func (v validationKwargs) populateArgs(value starlark.Value, parent starlark.Value, root starlark.Value) ([]starlark.Tuple, error) {
	kwargs := []starlark.Tuple{}
	if whenFunc, isFunc := v.when.(*starlark.Function); isFunc {
		for idx := 0; idx < whenFunc.NumParams(); idx++ {
			name, _ := whenFunc.Param(idx)
			switch name {
			case "v", "val", "value":
				kwargs = append(kwargs, starlark.Tuple{starlark.String(name), value})
			case "p", "par", "parent":
				kwargs = append(kwargs, starlark.Tuple{starlark.String(name), parent})
			case "r", "root", "d", "doc", "document":
				// When validation support is extended beyond Data Values and into documents in general, differentiate
				//   between "root" and "document":
				//   - since data values validation occurs on a document, "root" and "document" are synonyms, right now.
				//   - once validations can be defined in templates, then "root" will be a DocSet.
				kwargs = append(kwargs, starlark.Tuple{starlark.String(name), root})
			default:
				return nil, fmt.Errorf("Unknown argument (%s) expected one of [value, parent, document]", name)
			}
		}
	}
	return kwargs, nil
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
			msg:        fmt.Sprintf("not null"),
			assertion:  yttlibrary.NewAssertNotNull().CheckFunc(),
			isCritical: true,
			priority:   100,
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
	case *yamlmeta.DocumentSet:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode).AsStarlarkValue()
	case *yamlmeta.Array:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode).AsStarlarkValue()
	case *yamlmeta.Map:
		key = yamlmeta.TypeName(typedNode)
		nodeValue = yamltemplate.NewGoValueWithYAML(typedNode).AsStarlarkValue()
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
