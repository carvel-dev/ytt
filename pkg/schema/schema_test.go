// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"fmt"
	"testing"

	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func TestValueNotInSchemaErr(t *testing.T) {
	schemaYAML := `#@schema/match data_values=True
---
one: ""
two: ""
`
	valuesYAML := `#@data/values
---
one: cow
two: cows
not_in_schema: "this must fail validation."
`

	schemaDocSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(schemaYAML), "schema.yml")
	if err != nil {
		t.Fatalf("Unable to parse schema file: %s", err)
	}
	schemaVar, err := schema.NewDocumentSchema(schemaDocSet.GetValues()[1].(*yamlmeta.Document))
	if err != nil {
		t.Fatalf("Unable to create schema from file %v: %s", schemaDocSet.GetValues()[1].(*yamlmeta.Document).Position, err)
	}
	dataValuesDocSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(valuesYAML), "dataValues.yml")
	if err != nil {
		t.Fatalf("Unable to parse data values file: %s", err)
	}
	dataValueDoc := dataValuesDocSet.GetValues()[1].(*yamlmeta.Document)
	schemaVar.AssignType(dataValueDoc)
	typeCheck := dataValueDoc.Check()

	const expectedErrorMessage = `
dataValues.yml:5 | not_in_schema
                 | ^^^
                 |  unexpected key in map (as defined at schema.yml:2)
`
	if !typeCheck.HasViolations() {
		t.Fatalf("Expected schema validation to fail with: %s. But got success", expectedErrorMessage)
	}
	errorMessage := fmt.Sprintf("%v", typeCheck)
	if errorMessage != expectedErrorMessage {
		t.Fatalf("Expected schema validation to fail with: %s. But got: %s", expectedErrorMessage, errorMessage)
	}
}
