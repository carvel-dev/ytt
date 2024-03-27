// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

// ytt library extensions are defined in this package.
// They have been separated from "yttlibrary" package because
// they depend on functionality outside of Go standard library.
// Users who import ytt as a library in their Go programs may not
// want to depend on such functionality.

package yttlibraryext

import (
	_ "carvel.dev/ytt/pkg/yttlibraryext/toml" // include toml
)
