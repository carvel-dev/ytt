// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package texttemplate

import (
	"strings"
)

type NodeCodeMeta struct {
	*NodeCode
}

func (p NodeCodeMeta) ShoudTrimSpaceLeft() bool {
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
