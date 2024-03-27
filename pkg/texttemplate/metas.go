// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package texttemplate

import (
	"strings"
)

type NodeCodeMeta struct {
	*NodeCode
}

// ShouldTrimSpaceLeft indicates whether leading spaces should be removed (because the left-trim token, `-`, was present)
func (p NodeCodeMeta) ShouldTrimSpaceLeft() bool {
	return strings.HasPrefix(p.Content, "-")
}

func (p NodeCodeMeta) ShouldTrimSpaceRight() bool {
	return strings.HasSuffix(p.Content, "-")
}

func (p NodeCodeMeta) ShouldPrint() bool {
	return strings.HasPrefix(p.Content, "=") || strings.HasPrefix(p.Content, "-=")
}

func (p NodeCodeMeta) Code() string {
	result := strings.TrimPrefix(p.Content, "-=") // longer first
	result = strings.TrimPrefix(result, "=")
	result = strings.TrimPrefix(result, "-")
	return strings.TrimSuffix(result, "-")
}
