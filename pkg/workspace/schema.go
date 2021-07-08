// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package workspace

import (
	"github.com/k14s/ytt/pkg/schema"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

type Schema interface {
	AssignType(typeable yamlmeta.Typeable) yamlmeta.TypeCheck
	// DefaultDataValues yields default values for Data Values...
	//   if schema was built by schema.NewNullSchema (i.e. no schema was provided), returns nil
	DefaultDataValues() *yamlmeta.Document
}

var _ Schema = &schema.DocumentSchema{}
