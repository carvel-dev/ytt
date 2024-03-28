// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package filetests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimTrailingMultilineWhitespace(t *testing.T) {
	for _, testcase := range []struct {
		give, want string
	}{
		{
			give: `we want yaml`,
			want: `we want yaml`,
		},
		{
			give: `we want yaml `,
			want: `we want yaml`,
		},
		{
			give: `we want yaml	`,
			want: `we want yaml`,
		},
		{
			give: `we want yaml
`,
			want: `we want yaml`,
		},
		{
			give: `
we 
want	
yaml  `,
			want: `
we
want
yaml`,
		},
		{
			give: `
we

  want	
	yaml

`,
			want: `
we

  want
	yaml`,
		},
	} {
		assert.Equal(t, testcase.want, TrimTrailingMultilineWhitespace(testcase.give))
	}
}
