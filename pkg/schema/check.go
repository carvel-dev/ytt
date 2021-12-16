// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func checkNode(n yamlmeta.Node) TypeCheck {
	switch typed := n.(type) {
	case *yamlmeta.Document:
		return CheckDocument(typed)
	case *yamlmeta.Map:
		return checkMap(typed)
	case *yamlmeta.MapItem:
		return checkMapItem(typed)
	case *yamlmeta.Array:
		return checkArray(typed)
	case *yamlmeta.ArrayItem:
		return checkArrayItem(typed)
	default:
		panic(fmt.Sprintf("unknown Node type: %T", n))
	}
}

// CheckDocument attempts type check of `d`.
// If `d` has "schema/type" metadata (typically attached using SetType()), `d` is checked against that schema, recursively.
// `chk` contains all the type violations found in the check.
func CheckDocument(d *yamlmeta.Document) (chk TypeCheck) {
	switch typedContents := d.Value.(type) {
	case yamlmeta.Node:
		chk = checkNode(typedContents)
	}

	return chk
}

func checkMap(m *yamlmeta.Map) (chk TypeCheck) {
	if GetType(m) == nil {
		return
	}
	check := GetType(m).CheckType(m)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
		return
	}

	for _, item := range m.Items {
		check = checkMapItem(item)
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}

func checkMapItem(mi *yamlmeta.MapItem) (chk TypeCheck) {
	check := GetType(mi).CheckType(mi)
	if check.HasViolations() {
		chk.Violations = check.Violations
		return
	}

	check = checkCollectionItem(mi.Value, GetType(mi).GetValueType(), mi.Position)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return
}

// is it possible to enter this function with valueType=NullType or AnyType?
func checkCollectionItem(value interface{}, valueType Type, position *filepos.Position) (chk TypeCheck) {
	switch typedValue := value.(type) {
	case *yamlmeta.Map:
		check := checkMap(typedValue)
		chk.Violations = append(chk.Violations, check.Violations...)
	case *yamlmeta.Array:
		check := checkArray(typedValue)
		chk.Violations = append(chk.Violations, check.Violations...)
	default:
		chk = valueType.CheckType(&yamlmeta.Scalar{Value: value, Position: position})
	}
	return chk
}

func checkArray(a *yamlmeta.Array) (chk TypeCheck) {
	for _, item := range a.Items {
		check := checkArrayItem(item)
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}

func checkArrayItem(ai *yamlmeta.ArrayItem) (chk TypeCheck) {
	if GetType(ai) == nil {
		return
	}
	// TODO: This check only ensures that the ai is of ArrayItem type
	//       which we know because if it was not we would not assign
	//       the type to it.
	//       Given this maybe we can completely remove this check
	//       Lets not forget that the check of the type of the item
	//       is done by checkCollectionItem
	chk = GetType(ai).CheckType(ai)
	if chk.HasViolations() {
		return
	}

	check := checkCollectionItem(ai.Value, GetType(ai).GetValueType(), ai.Position)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
	}
	return chk
}

// TypeCheck is the result of checking a yamlmeta.Node structure against a given Type, recursively.
type TypeCheck struct {
	Violations []error
}

// Error generates the error message composed of the total set of TypeCheck.Violations.
func (tc TypeCheck) Error() string {
	if !tc.HasViolations() {
		return ""
	}

	msg := ""
	for _, err := range tc.Violations {
		msg += err.Error() + "\n"
	}
	return msg
}

// HasViolations indicates whether this TypeCheck contains any violations.
func (tc *TypeCheck) HasViolations() bool {
	return len(tc.Violations) > 0
}
