// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func Check_Node(n yamlmeta.Node) TypeCheck {
	switch typed := n.(type) {
	case *yamlmeta.DocumentSet:
		return Check_DocumentSet(typed)
	case *yamlmeta.Document:
		return Check_Document(typed)
	case *yamlmeta.Map:
		return Check_Map(typed)
	case *yamlmeta.MapItem:
		return Check_MapItem(typed)
	case *yamlmeta.Array:
		return Check_Array(typed)
	case *yamlmeta.ArrayItem:
		return Check_ArrayItem(typed)
	default:
		panic(fmt.Sprintf("unknown Node type: %T", n))
	}
}

func Check_DocumentSet(ds *yamlmeta.DocumentSet) TypeCheck {
	return TypeCheck{}
}

func Check_Document(d *yamlmeta.Document) (chk TypeCheck) {
	switch typedContents := d.Value.(type) {
	case yamlmeta.Node:
		chk = Check_Node(typedContents)
	}

	return chk
}

func Check_Map(m *yamlmeta.Map) (chk TypeCheck) {
	if GetType(m) == nil {
		return
	}
	check := GetType(m).CheckType(m)
	if check.HasViolations() {
		chk.Violations = append(chk.Violations, check.Violations...)
		return
	}

	for _, item := range m.Items {
		check = Check_MapItem(item)
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}

func Check_MapItem(mi *yamlmeta.MapItem) (chk TypeCheck) {
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
		check := Check_Map(typedValue)
		chk.Violations = append(chk.Violations, check.Violations...)
	case *yamlmeta.Array:
		check := Check_Array(typedValue)
		chk.Violations = append(chk.Violations, check.Violations...)
	default:
		chk = valueType.CheckType(&yamlmeta.Scalar{Value: value, Position: position})
	}
	return chk
}

func Check_Array(a *yamlmeta.Array) (chk TypeCheck) {
	for _, item := range a.Items {
		check := Check_ArrayItem(item)
		if check.HasViolations() {
			chk.Violations = append(chk.Violations, check.Violations...)
		}
	}
	return
}

func Check_ArrayItem(ai *yamlmeta.ArrayItem) (chk TypeCheck) {
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

type TypeCheck struct {
	Violations []error
}

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

func (tc *TypeCheck) HasViolations() bool {
	return len(tc.Violations) > 0
}
