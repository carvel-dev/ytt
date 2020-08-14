// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
)

type StarlarkNoop struct{}

var _ starlark.Value = &StarlarkNoop{}

func (s *StarlarkNoop) String() string        { return "noop" }
func (s *StarlarkNoop) Type() string          { return "noop" }
func (s *StarlarkNoop) Freeze()               {}
func (s *StarlarkNoop) Truth() starlark.Bool  { return false }
func (s *StarlarkNoop) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: noop") }
