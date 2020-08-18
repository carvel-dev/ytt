// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})
