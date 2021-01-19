// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})
