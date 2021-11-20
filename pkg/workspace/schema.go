// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package workspace

import (
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type Schema interface {
	AssignType(TypeWithValues schema.TypeWithValues) schema.TypeCheck

	// DefaultDataValues should yield the default values for Data Values...
	//   if schema was built by schema.NewNullSchema (i.e. no schema was provided), returns nil
	DefaultDataValues() *yamlmeta.Document

	// GetDocumentType should return a reference to the DocumentType that is the root of this Schema.
	GetDocumentType() *schema.DocumentType
}

var _ Schema = &schema.DocumentSchema{}
