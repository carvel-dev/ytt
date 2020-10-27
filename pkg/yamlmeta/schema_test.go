// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"fmt"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"testing"
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
	schema := yamlmeta.NewDocumentSchema(schemaDocSet.GetValues()[1].(*yamlmeta.Document))
	dataValuesDocSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(valuesYAML), "dataValues.yml")
	if err != nil {
		t.Fatalf("Unable to parse data values file: %s", err)
	}
	dataValueDoc := dataValuesDocSet.GetValues()[1].(*yamlmeta.Document)
	schema.AssignType(dataValueDoc)
	typeCheck := dataValueDoc.Check()

	const expectedErrorMessage = "{[Map item 'not_in_schema' at dataValues.yml:5 is not defined in schema]}"
	if !typeCheck.HasViolations() {
		t.Fatalf("Expected schema validation to fail with: %s. But got success", expectedErrorMessage)
	}
	errorMessage := fmt.Sprintf("%v", typeCheck)
	if errorMessage != expectedErrorMessage {
		t.Fatalf("Expected schema validation to fail with: %s. But got: %s", expectedErrorMessage, errorMessage)
	}
}
