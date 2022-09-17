package filetests

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
