// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
)

type NodeValidation struct {
	rules []Rule
	ValidationKwargs
	position *filepos.Position
}

// A Rule represents a validation attached to a Node via an annotation;
// it contains a string description of what constitutes a valid value,
// and a function that asserts the rule against an actual value.
type Rule struct {
	msg       string
	assertion starlark.Callable
}

// Validate runs the assertion from the Rule with the node's value as arguments.
//
// Returns an error if the assertion returns False (not-None), or assert.fail()s.
// Otherwise, returns nil.
func (v NodeValidation) Validate(node yamlmeta.Node, thread *starlark.Thread) []error {
	key, nodeValue := v.convertToStarlarkValue(node)

	// if value does not satisfy the 'when' clauses
	shouldCheck, err := v.ValidationKwargs.Check(nodeValue, thread)
	if err != nil {
		return []error{err}
	}
	if shouldCheck {
		var failures []error
		for _, r := range v.rules {
			result, fail := starlark.Call(thread, r.assertion, starlark.Tuple{nodeValue}, []starlark.Tuple{})
			if fail != nil {
				failures = append(failures, fail)
			} else if _, ok := result.(starlark.NoneType); !ok {
				// in order to pass, the assertion must return True or None
				if !result.Truth() {
					failures = append(failures, fmt.Errorf("%s (%s) requires %q (by %s)", key, node.GetPosition().AsCompactString(), r.msg, v.position.AsCompactString()))
				}
			}
		}
		return failures
	}

	return nil
}

func (v NodeValidation) convertToStarlarkValue(node yamlmeta.Node) (string, starlark.Value) {
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

type ValidationKwargs struct {
	when         *starlark.Callable
	whenNullSkip bool
}

func NewValidationKwargs(kwargs []starlark.Tuple, annPos *filepos.Position) (ValidationKwargs, error) {
	var processedKwargs ValidationKwargs
	for _, value := range kwargs {
		kwargName := string(value[0].(starlark.String))
		switch kwargName {
		case ValidationKwargWhen:
			lambda, ok := value[1].(starlark.Callable)
			if !ok {
				return ValidationKwargs{}, fmt.Errorf("expected keyword argument %q to a function, but was %s (at %s)", ValidationKwargWhen, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.when = &lambda
		case ValidationKwargWhenNull:
			b, ok := value[1].(starlark.Bool)
			if !ok {
				return ValidationKwargs{}, fmt.Errorf("expected keyword argument %q to a boolean, but was %s (at %s)", ValidationKwargWhenNull, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.whenNullSkip = bool(b)
		default:
			return ValidationKwargs{}, fmt.Errorf("unknown keyword argument %q (at %s)", kwargName, annPos.AsCompactString())
		}
	}
	return processedKwargs, nil
}

func (v ValidationKwargs) Check(value starlark.Value, thread *starlark.Thread) (bool, error) {
	if v.whenNullSkip {
		if _, ok := value.(starlark.NoneType); ok {
			return false, nil
		}
	}

	if v.when != nil {
		result, err := starlark.Call(thread, *v.when, starlark.Tuple{value}, []starlark.Tuple{})
		if err != nil {
			return false, err
		}

		switch typedResult := result.(type) {
		case starlark.NoneType:
			return true, nil
		case starlark.Bool:
			return bool(typedResult), nil
		default:
			return false, fmt.Errorf("when kwarg function must reutrn a bool or None type")
		}
	}
	// if no kwags then we should check
	return true, nil
}
