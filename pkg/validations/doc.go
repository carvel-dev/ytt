// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

/*
Package validations enriches YAML structures by attaching user-defined
constraints (that is, validation rules) onto individual yamlmeta.Node's.

# Validations on Data Values

While "@data/values" can technically be annotated with "@assert/validate"
annotations, it is expected that authors will use "@schema/validation" in
"@data/values-schema" documents instead.
*/
package validations
