// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

/*
Package schema enhances yamlmeta.Node structures with fine-grain data types.

# Type and Checking

Within a schema document (which itself is a yamlmeta.Document), type metadata is
attached to nodes via @schema/... annotations. Those annotations are parsed and
digested into a composite tree of schema.Type.

With such a schema.Type hierarchy in hand, other YAML documents can be
"typed" (via Type.AssignTypeTo()). This process walks the target
yamlmeta.Document and attaches the corresponding type as metadata on each
corresponding yamlmeta.Node.

"Typed" documents can then be "checked" (via schema.CheckNode()) to determine if
their fine-grain types conform to the assigned schema.

# Documenting

Other @schema/... annotations are used to describe the exact syntax and semantic
of values.

# Other Schema Formats

Like other Carvel tools, ytt aims to interoperate well with other tooling. In
this vein, ytt can export schema defined within ytt as an OpenAPI v3 document.
*/
package schema
