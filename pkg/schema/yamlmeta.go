// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package schema

import "carvel.dev/ytt/pkg/yamlmeta"

const schemaType = "schema/type"

// GetType retrieves schema metadata from `n`, set previously via SetType().
func GetType(n yamlmeta.Node) Type {
	t := n.GetMeta(schemaType)
	if t == nil {
		return nil
	}
	return t.(Type)
}

// SetType attaches schema metadata to `n`, later retrieved via GetType().
func SetType(n yamlmeta.Node, t Type) {
	n.SetMeta(schemaType, t)
}
