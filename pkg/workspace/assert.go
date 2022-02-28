package workspace

import (
	"fmt"
	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamltemplate"
)

const AnnotationAssertValidate template.AnnotationName = "assert/validate"

func RunAssertions(n yamlmeta.Node) error {
	if n == nil {
		return nil
	}
	assertChecker := newAssertChecker()

	yamlmeta.Walk(n, assertChecker)
	if assertChecker.hasViolations() {
		return assertChecker.errs[0]
	}
	return nil
}

func newAssertChecker() *assertChecker {
	return &assertChecker{errs: nil}
}

type assertChecker struct {
	errs []error
}

func (a *assertChecker) hasViolations() bool {
	if len(a.errs) > 0 {
		return true
	}
	return false
}

func (a *assertChecker) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if nodeAnnotations.Has(AnnotationAssertValidate) {
		validations := newValidationsFromAssertValidateAnnotation(nodeAnnotations[AnnotationAssertValidate])
		for _, v := range validations {
			err := v.Validate(node)
			if err != nil {
				a.errs = append(a.errs, err)
			}
		}
	}

	return nil
}

func newValidationsFromAssertValidateAnnotation(annotation template.NodeAnnotation) []Validation {
	var validations []Validation
	thread := &starlark.Thread{Name: fmt.Sprintf("## %v ##", AnnotationAssertValidate)}
	for _, tup := range annotation.Args {
		validationTuple, ok := tup.(starlark.Tuple)
		if !ok || len(validationTuple) != 2 {
			panic("arg is not a tuple")
		}
		message, ok := validationTuple[0].(starlark.String)
		if !ok {
			panic("first arg of tuple is not string")
		}
		lambda, ok := validationTuple[1].(starlark.Callable)
		if !ok {
			panic("second arg of tuple is not a function")
		}
		validations = append(validations, Validation{
			thread:   thread,
			msg:      message.String(),
			f:        lambda,
			position: annotation.Position,
		})
	}

	return validations
}

type Validation struct {
	thread   *starlark.Thread
	msg      string
	f        starlark.Callable
	position *filepos.Position
}

// Validate compares two values for equality
func (v Validation) Validate(node yamlmeta.Node) error {
	var nodeValue starlark.Value
	values := node.GetValues()
	switch len(values) {
	case 0:
	case 1:
		nodeValue = yamltemplate.NewGoValueWithYAML(values[0]).AsStarlarkValue()

	default:
		nodeValue = yamltemplate.NewGoValueWithYAML(values).AsStarlarkValue()
	}

	result, err := starlark.Call(v.thread, v.f, starlark.Tuple{nodeValue}, []starlark.Tuple{})
	if err != nil {
		return err
	}

	if !result.Truth() {
		node.GetAnnotations()
		fmt.Errorf("error on %v: %v", v.position.AsCompactString(), v.msg)
	}
	return nil
	//if args.Len() != 2 {
	//	return starlark.None, fmt.Errorf("expected two arguments")
	//}

	//assertionArg := args.Index(1)
	//if _, ok := assertionArg.(starlark.Callable); !ok {
	//	return starlark.None, fmt.Errorf("expected argument to be a function, but was %T", assertionArg)
	//}

	//expectedString, err := core.NewStarlarkValue(args.Index(0)).AsString()
	//if err != nil {
	//	return starlark.None, err
	//}
	//
	//lambda := args.Index(1)
	//if _, ok := lambda.(starlark.Callable); !ok {
	//	return starlark.None, fmt.Errorf("expected argument to be a function, but was %T", lambda)
	//}
	//
	//retVal, err := starlark.Call(thread, lambda, nil, nil)
	//if err != nil {
	//	return starlark.Tuple{starlark.None, starlark.String(expectedString + " " + err.Error())}, nil
	//}
	//
	//return starlark.Tuple{retVal, starlark.None}, nil
	return nil

}
