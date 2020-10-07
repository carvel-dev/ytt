// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import (
	"fmt"
)

type Schema interface {
	AssignType(document *Document)
}

type AnySchema struct {
}

type DocumentSchema struct {
	Allowed *DocumentType
}

type TypeCheck struct {
	Violations []string
}

func (tc *TypeCheck) HasViolations() bool {
	return len(tc.Violations) > 0
}

type DocumentType struct {
	*Document
}

type MapType struct {
	Map
}

type MapItemType struct {
	MapItem
}

func (mt *MapType) AllowsKey(key interface{}) bool {
	for _, item := range mt.Items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func (mt MapType) CheckAllows(item *MapItem) TypeCheck {
	typeCheck := TypeCheck{}

	if !mt.AllowsKey(item.Key) {
		typeCheck.Violations = append(typeCheck.Violations, fmt.Sprintf("Map item '%s' at %s is not defined in schema", item.Key, item.Position.AsCompactString()))
	}
	return typeCheck
}

func NewDocumentSchema(doc *Document) *DocumentSchema {
	schemaDoc := &DocumentSchema{Allowed: &DocumentType{&Document{Value: nil}}}

	switch typedContent := doc.Value.(type) {
	case *Map:
		mapType := &MapType{}
		for _, mapItem := range typedContent.Items {
			mapType.Items = append(mapType.Items, &MapItem{Key: mapItem.Key, Value: mapItem.Value, Type: "string"})
		}
		schemaDoc.Allowed = &DocumentType{&Document{Value: mapType}}
	}

	return schemaDoc
}

func (d *Document) Check() TypeCheck {
	var typeCheck TypeCheck

	switch typedContents := d.Value.(type) {
	case Node:
		typeCheck = typedContents.Check()
	}

	return typeCheck
}

func (m *Map) Check() TypeCheck {
	typeCheck := TypeCheck{}

	for _, item := range m.Items {
		check := m.Type.CheckAllows(item)
		if check.HasViolations() {
			typeCheck.Violations = append(typeCheck.Violations, check.Violations...)
			continue
		}

		check = item.Check()
		if check.HasViolations() {
			typeCheck.Violations = append(typeCheck.Violations, check.Violations...)
		}
	}
	return typeCheck
}

func (d *DocumentSet) Check() TypeCheck { return TypeCheck{} }
func (d *MapItem) Check() TypeCheck     { return TypeCheck{} }
func (d *Array) Check() TypeCheck       { return TypeCheck{} }
func (d *ArrayItem) Check() TypeCheck   { return TypeCheck{} }

func (as AnySchema) AssignType(doc *Document) {
	doc.Type = DocumentType{}
}

func (s DocumentSchema) AssignType(doc *Document) {
	switch typedNode := doc.Value.(type) {
	case *Map:
		mapType, ok := s.Allowed.Value.(*MapType)
		if !ok {
			typedNode.Type = &MapType{}
			// during typing we dont report error
			break
		}
		// set the type on the map
		typedNode.Type = mapType
		for _, item := range typedNode.Items {
			for _, mapTypeItem := range mapType.Items {
				if item.Key == mapTypeItem.Key {
					item.Type = mapTypeItem.Type
				}
			}
		}
	}
}
