// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/template/core"
	"carvel.dev/ytt/pkg/yamlmeta"
	"carvel.dev/ytt/pkg/yamltemplate"
	"carvel.dev/ytt/pkg/yttlibrary"
	"github.com/k14s/starlark-go/starlark"
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
	err := yamlmeta.WalkWithParent(node, nil, "", validation)
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
	return &validationRun{thread: &starlark.Thread{Name: threadName}, root: root}
}

// VisitWithParent if `node` has validations in its meta.
// Runs those validations, collecting any violations
//
// This visitor stores error(violations) in the validationRun and returns nil.
func (a *validationRun) VisitWithParent(value yamlmeta.Node, parent yamlmeta.Node, path string) error {
	// get rules in node's meta
	validations := Get(value)

	if validations == nil {
		return nil
	}
	for _, v := range validations {
		invalid, err := v.Validate(value, parent, a.root, path, a.thread)
		if err != nil {
			return err
		}
		if len(invalid.Violations) > 0 {
			a.chk.Invalidations = append(a.chk.Invalidations, invalid)
		}
	}

	return nil
}

// Validate runs the assertions in the rules with the node's value as arguments IF
// the ValidationKwargs conditional options pass.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (v NodeValidation) Validate(node yamlmeta.Node, parent yamlmeta.Node, root yamlmeta.Node, path string, thread *starlark.Thread) (Invalidation, error) {
	nodeValue := v.newStarlarkValue(node)
	parentValue := v.newStarlarkValue(parent)
	rootValue := v.newStarlarkValue(root)

	executeRules, err := v.kwargs.shouldValidate(nodeValue, parentValue, thread, rootValue)
	if err != nil {
		return Invalidation{}, fmt.Errorf("Validating %s: %s", path, err)
	}
	if !executeRules {
		return Invalidation{}, nil
	}

	displayedPath := path
	if displayedPath == "" {
		displayedPath = fmt.Sprintf("(%s)", yamlmeta.TypeName(node))
	}
	invalid := Invalidation{
		Path:        displayedPath,
		ValueSource: node.GetPosition(),
	}

	for _, rul := range byPriority(v.rules) {
		result, err := starlark.Call(thread, rul.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
		if err != nil {
			violation := Violation{
				RuleSource:  v.position,
				Description: rul.msg,
				Results:     strings.TrimPrefix(strings.TrimPrefix(err.Error(), "fail: "), "check: "),
			}
			invalid.Violations = append(invalid.Violations, violation)
			if rul.isCritical {
				break
			}
		} else {
			if !(result == starlark.True) {
				violation := Violation{
					RuleSource:  v.position,
					Description: rul.msg,
					Results:     "",
				}
				invalid.Violations = append(invalid.Violations, violation)
				if rul.isCritical {
					break
				}
			}
		}
	}
	return invalid, nil
}

// shouldValidate uses validationKwargs and the node's value to run checks on the value. If the value satisfies the checks,
// then the NodeValidation's rules should execute, otherwise the rules will be skipped.
func (v validationKwargs) shouldValidate(value starlark.Value, parent starlark.Value, thread *starlark.Thread, root starlark.Value) (bool, error) {
	_, valueIsNull := value.(starlark.NoneType)
	if valueIsNull && !v.notNull {
		return false, nil
	}

	if v.when != nil && !reflect.ValueOf(v.when).IsNil() {
		args, err := v.populateArgs(value, parent, root)
		if err != nil {
			return false, fmt.Errorf("Failed to evaluate when=: %s", err)
		}

		result, err := starlark.Call(thread, v.when, args, []starlark.Tuple{})
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

func (v validationKwargs) populateArgs(value starlark.Value, parent starlark.Value, root starlark.Value) ([]starlark.Value, error) {
	args := []starlark.Value{}
	args = append(args, value)

	whenFunc := v.when.(*starlark.Function)
	switch whenFunc.NumParams() {
	case 1:
	case 2:
		ctx := orderedmap.NewMap()
		ctx.Set("parent", parent)
		ctx.Set("root", root)
		args = append(args, core.NewStarlarkStruct(ctx))
	default:
		return nil, fmt.Errorf("function must accept 1 or 2 arguments (%d given)", whenFunc.NumParams())
	}
	return args, nil
}

func (v validationKwargs) asRules() []rule {
	var rules []rule

	if v.minLength != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length >= %v", *v.minLength),
			assertion: yttlibrary.NewAssertMinLen(*v.minLength).CheckFunc(),
		})
	}
	if v.maxLength != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("length <= %v", *v.maxLength),
			assertion: yttlibrary.NewAssertMaxLen(*v.maxLength).CheckFunc(),
		})
	}
	if v.min != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value >= %v", v.min),
			assertion: yttlibrary.NewAssertMin(v.min).CheckFunc(),
		})
	}
	if v.max != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("a value <= %v", v.max),
			assertion: yttlibrary.NewAssertMax(v.max).CheckFunc(),
		})
	}
	if v.notNull {
		rules = append(rules, rule{
			msg:        "not null",
			assertion:  yttlibrary.NewAssertNotNull().CheckFunc(),
			isCritical: true,
			priority:   100,
		})
	}
	if v.oneNotNull != nil {
		var assertion *yttlibrary.Assertion
		var childKeys = ""

		switch oneNotNull := v.oneNotNull.(type) {
		case starlark.Bool:
			if oneNotNull {
				assertion = yttlibrary.NewAssertOneNotNull(nil)
				childKeys = "all children"
			} else {
				// should have been caught when args were parsed
				panic("one_not_null= cannot be False")
			}
		case starlark.Sequence:
			assertion = yttlibrary.NewAssertOneNotNull(oneNotNull)
			childKeys = oneNotNull.String()
		default:
			// should have been caught when args were parsed
			panic(fmt.Sprintf("Unexpected type \"%s\" for one_not_null=", v.oneNotNull.Type()))
		}

		rules = append(rules, rule{
			msg:       fmt.Sprintf("exactly one of %s to be not null", childKeys),
			assertion: assertion.CheckFunc(),
		})
	}
	if v.oneOf != nil {
		rules = append(rules, rule{
			msg:       fmt.Sprintf("one of %s", v.oneOf.String()),
			assertion: yttlibrary.NewAssertOneOf(v.oneOf).CheckFunc(),
		})
	}

	return rules
}

// Invalidation describes a value that was invalidated, and how.
type Invalidation struct {
	Path        string
	ValueSource *filepos.Position
	Violations  []Violation
}

// Violation describes how a value failed to satisfy a rule.
type Violation struct {
	RuleSource  *filepos.Position
	Description string
	Results     string
}

// Check holds the complete set of Invalidations (if any) resulting from checking all validation rules.
type Check struct {
	Invalidations []Invalidation
}

// ResultsAsString generates the error message composed of the total set of Check.Invalidations.
func (c Check) ResultsAsString() string {
	if !c.HasInvalidations() {
		return ""
	}

	msg := ""
	for _, inval := range c.Invalidations {
		msg += fmt.Sprintf("  %s\n    from: %s\n", inval.Path, inval.ValueSource.AsCompactString())
		for _, viol := range inval.Violations {
			msg += fmt.Sprintf("    - must be: %s (by: %s)\n", viol.Description, viol.RuleSource.AsCompactString())
			if viol.Results != "" {
				msg += fmt.Sprintf("      found: %s\n", viol.Results)
			}
		}
		msg += "\n"
	}
	return msg
}

// HasInvalidations indicates whether this Check contains any violations.
func (c Check) HasInvalidations() bool {
	return len(c.Invalidations) > 0
}

func (v NodeValidation) newStarlarkValue(node yamlmeta.Node) starlark.Value {
	if node == nil || reflect.ValueOf(node).IsNil() {
		return starlark.None
	}

	switch node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		return yamltemplate.NewGoValueWithYAML(node).AsStarlarkValue()
	case *yamlmeta.Document, *yamlmeta.MapItem, *yamlmeta.ArrayItem:
		return yamltemplate.NewGoValueWithYAML(node.GetValues()[0]).AsStarlarkValue()
	default:
		panic(fmt.Sprintf("Unexpected node type %T (at or near %s)", node, node.GetPosition().AsCompactString()))
	}
}
