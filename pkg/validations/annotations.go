// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"

	"carvel.dev/ytt/pkg/filepos"
	"carvel.dev/ytt/pkg/template"
	"carvel.dev/ytt/pkg/yamlmeta"
	"github.com/k14s/starlark-go/starlark"
)

// Declare @assert/... annotation and keyword argument names
const (
	AnnotationAssertValidate template.AnnotationName = "assert/validate"

	KwargWhen       string = "when"
	KwargMinLength  string = "min_len"
	KwargMaxLength  string = "max_len"
	KwargMin        string = "min"
	KwargMax        string = "max"
	KwargNotNull    string = "not_null"
	KwargOneNotNull string = "one_not_null"
	KwargOneOf      string = "one_of"
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
// Checks annotation, and stores the validationRun on Node's validations meta.
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
		validation, err := NewValidationFromAnn(nodeAnnotations[AnnotationAssertValidate])
		if err != nil {
			return fmt.Errorf("Invalid @%s annotation - %s", AnnotationAssertValidate, err.Error())
		}
		// store rules in node's validations meta without overriding any existing rules
		Add(node, []NodeValidation{*validation})
	}

	return nil
}

// NewValidationFromAnn creates a NodeValidation from the values provided in a validationRun-style annotation.
func NewValidationFromAnn(annotation template.NodeAnnotation) (*NodeValidation, error) {
	var rules []rule
	if len(annotation.Args) == 0 && len(annotation.Kwargs) == 0 {
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

		assertion, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			var err error
			assertion, err = assertionFromCheckAttr(ruleTuple[1])
			if err != nil {
				return nil, fmt.Errorf("%s (at %s)", err, annotation.Position.AsCompactString())
			}
		}
		rules = append(rules, rule{
			msg:       message.GoString(),
			assertion: assertion,
		})
	}
	kwargs, err := newValidationKwargs(annotation.Kwargs, annotation.Position)
	if err != nil {
		return nil, err
	}

	rules = append(rules, kwargs.asRules()...)

	return &NodeValidation{rules, kwargs, annotation.Position}, nil
}

func assertionFromCheckAttr(value starlark.Value) (starlark.Callable, error) {
	val, hasAttrs := value.(starlark.HasAttrs)
	if !hasAttrs {
		return nil, fmt.Errorf("expected second item in the 2-tuple to be an assertion function, but was %s", value.Type())
	}

	checkAttr, err := val.Attr("check")
	if err != nil || (checkAttr == nil && err == nil) {
		return nil, fmt.Errorf("expected second item in tuple to be an assertion function or assertion object, but was a %s", value.Type())
	}

	assertionFunc, ok := checkAttr.(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("expected struct with attribute check(), but was %s", checkAttr.Type())
	}

	return assertionFunc, nil
}

// newValidationKwargs takes the keyword arguments from a Validation annotation,
// and makes sure they are well-formed.
func newValidationKwargs(kwargs []starlark.Tuple, annPos *filepos.Position) (validationKwargs, error) {
	var processedKwargs validationKwargs
	for _, value := range kwargs {
		kwargName := string(value[0].(starlark.String))
		switch kwargName {
		case KwargWhen:
			v, ok := value[1].(starlark.Callable)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a function, but was %s (at %s)", KwargWhen, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.when = v
		case KwargMinLength:
			v, err := starlark.NumberToInt(value[1])
			if err != nil {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a number, but was %s (at %s)", KwargMinLength, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.minLength = &v
		case KwargMaxLength:
			v, err := starlark.NumberToInt(value[1])
			if err != nil {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a number, but was %s (at %s)", KwargMaxLength, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.maxLength = &v
		case KwargMin:
			processedKwargs.min = value[1]
		case KwargMax:
			processedKwargs.max = value[1]
		case KwargNotNull:
			v, ok := value[1].(starlark.Bool)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a boolean, but was %s (at %s)", KwargNotNull, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.notNull = bool(v)
		case KwargOneNotNull:
			switch v := value[1].(type) {
			case starlark.Bool:
				if v {
					processedKwargs.oneNotNull = v
				} else {
					return validationKwargs{}, fmt.Errorf("one_not_null= cannot be False")
				}
			case starlark.Sequence:
				processedKwargs.oneNotNull = v
			default:
				return validationKwargs{}, fmt.Errorf("expected True or a sequence of keys, but was a '%s'", value[1].Type())
			}
		case KwargOneOf:
			v, ok := value[1].(starlark.Sequence)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %s to be a sequence, but was %s", KwargOneOf, value[1].Type())
			}
			processedKwargs.oneOf = v
		default:
			return validationKwargs{}, fmt.Errorf("unknown keyword argument %q (at %s)", kwargName, annPos.AsCompactString())
		}
	}
	return processedKwargs, nil
}
