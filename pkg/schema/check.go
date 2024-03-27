// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"carvel.dev/ytt/pkg/yamlmeta"
)

// CheckNode attempts type check of root node and its children.
//
// If `n` has "schema/type" metadata (typically attached using SetType()), `n` is checked against that schema, recursively.
// `chk` contains all the type violations found in the check.
func CheckNode(n yamlmeta.Node) TypeCheck {
	checker := newTypeChecker()

	err := yamlmeta.Walk(n, checker)
	if err != nil {
		panic(err)
	}

	return *checker.chk
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

func newTypeChecker() *typeChecker {
	return &typeChecker{chk: &TypeCheck{}}
}

type typeChecker struct {
	chk *TypeCheck
}

func (t *typeChecker) Visit(node yamlmeta.Node) error {
	nodeType := GetType(node)
	if nodeType == nil {
		return nil
	}

	chk := nodeType.CheckType(node)
	if chk.HasViolations() {
		t.chk.Violations = append(t.chk.Violations, chk.Violations...)
	}

	return nil
}

var _ yamlmeta.Visitor = &typeChecker{}

// CheckType checks the type of `node` against this DocumentType.
//
// If `node` is not a yamlmeta.Document, `chk` contains a violation describing this mismatch
// If this document's value is a scalar, checks its type against this DocumentType.ValueType
func (t *DocumentType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	doc, ok := node.(*yamlmeta.Document)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, t))
		return chk
	}

	if _, isNode := doc.Value.(yamlmeta.Node); !isNode {
		valChk := t.ValueType.CheckType(doc) // ScalarType checks doc.Value
		if valChk.HasViolations() {
			chk.Violations = append(chk.Violations, valChk.Violations...)
		}
	}
	return chk
}

// CheckType checks the type of `node` against this MapType.
//
// If `node` is not a yamlmeta.Map, `chk` contains a violation describing this mismatch
// If a contained yamlmeta.MapItem is not allowed by this MapType, `chk` contains a corresponding violation
func (m *MapType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	nodeMap, ok := node.(*yamlmeta.Map)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, m))
		return chk
	}

	for _, item := range nodeMap.Items {
		if !m.AllowsKey(item.Key) {
			chk.Violations = append(chk.Violations, NewUnexpectedKeyAssertionError(item, m.Position, m.AllowedKeys()))
		}
	}
	return chk
}

// AllowsKey determines whether this MapType permits a MapItem with the key of `key`
func (m *MapType) AllowsKey(key interface{}) bool {
	for _, item := range m.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

// AllowedKeys returns the set of keys (in string format) permitted in this map.
func (m *MapType) AllowedKeys() []string {
	var keysAsString []string

	for _, item := range m.Items {
		keysAsString = append(keysAsString, fmt.Sprintf("%s", item.Key))
	}

	return keysAsString
}

// CheckType checks the type of `node` against this MapItemType.
//
// If `node` is not a yamlmeta.MapItem, `chk` contains a violation describing this mismatch
// If this map item's value is a scalar, checks its type against this MapItemType.ValueType
func (t *MapItemType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	mapItem, ok := node.(*yamlmeta.MapItem)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, t))
		return chk
	}

	if _, isNode := mapItem.Value.(yamlmeta.Node); !isNode {
		valChk := t.ValueType.CheckType(mapItem) // ScalarType checks mapItem.Value
		if valChk.HasViolations() {
			chk.Violations = append(chk.Violations, valChk.Violations...)
		}
	}
	return chk
}

// CheckType checks the type of `node` against this ArrayType
//
// If `node` is not a yamlmeta.Array, `chk` contains a violation describing the mismatch
func (a *ArrayType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	_, ok := node.(*yamlmeta.Array)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, a))
	}
	return chk
}

// CheckType checks the type of `node` against this ArrayItemType.
//
// If `node` is not a yamlmeta.ArrayItem, `chk` contains a violation describing this mismatch
// If this array item's value is a scalar, checks its type against this ArrayItemType.ValueType
func (a *ArrayItemType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	arrayItem, ok := node.(*yamlmeta.ArrayItem)
	if !ok {
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, a))
		return chk
	}

	if _, isNode := arrayItem.Value.(yamlmeta.Node); !isNode {
		valChk := a.ValueType.CheckType(arrayItem) // ScalarType checks arrayItem.Value
		if valChk.HasViolations() {
			chk.Violations = append(chk.Violations, valChk.Violations...)
		}
	}
	return chk
}

// CheckType checks the type of `node`'s `value`, which is expected to be a scalar type.
//
// If the value is not a recognized scalar type, `chk` contains a corresponding violation
// If the value is not of the type specified in this ScalarType, `chk` contains a violation describing the mismatch
func (s *ScalarType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	if len(node.GetValues()) < 1 {
		panic(fmt.Sprintf("Expected a node that could hold a scalar value, but was %#v", node))
	}
	value := node.GetValues()[0]
	switch value.(type) {
	case string:
		if s.ValueType != StringType {
			chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, s))
		}
	case float64:
		if s.ValueType != FloatType {
			chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, s))
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if s.ValueType != IntType && s.ValueType != FloatType {
			// integers between -9007199254740992 and 9007199254740992 fits in a float64 with no loss of precision.
			chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, s))
		}
	case bool:
		if s.ValueType != BoolType {
			chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, s))
		}
	default:
		chk.Violations = append(chk.Violations, NewMismatchedTypeAssertionError(node, s))
	}
	return chk
}

// CheckType is a no-op because AnyType allows any value.
//
// Always returns an empty TypeCheck.
func (a AnyType) CheckType(_ yamlmeta.Node) TypeCheck {
	return TypeCheck{}
}

// CheckType checks the type of `node` against this NullType
//
// If `node`'s value is null, this check passes
// If `node`'s value is not null, then it is checked against this NullType's wrapped Type.
func (n NullType) CheckType(node yamlmeta.Node) TypeCheck {
	chk := TypeCheck{}
	node, isNode := node.(yamlmeta.Node)
	if !isNode {
		if node == nil {
			return chk
		}
	}
	if len(node.GetValues()) == 1 && node.GetValues()[0] == nil {
		return chk
	}

	check := n.GetValueType().CheckType(node)
	chk.Violations = check.Violations

	return chk
}
