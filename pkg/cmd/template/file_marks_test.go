// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"testing"

	"carvel.dev/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func newFile(path string) *files.File {
	return files.MustNewFileFromSource(files.NewBytesSource(path, []byte{}))
}

func TestFileMarksOpts_Apply_RelaxedBehavior(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		files   []*files.File
		marks   []string
		relaxed bool
		wantErr bool
	}{
		"strict no-glob": {
			files: []*files.File{
				newFile("bar.txt"),
			},
			marks:   []string{"foo.txt:type=text-plain"},
			wantErr: true,
		},
		"strict glob": {
			files: []*files.File{
				newFile("bar.yaml"),
			},
			marks:   []string{"*.txt:type=text-plain"},
			wantErr: true,
		},
		"relaxed no-glob": {
			files: []*files.File{
				newFile("baz.yaml"),
			},
			marks:   []string{"foo.txt:type=text-plain"},
			relaxed: true,
			wantErr: true,
		},
		"relaxed glob": {
			files: []*files.File{
				newFile("baz.yaml"),
			},
			marks:   []string{"*.txt:type=text-plain"},
			relaxed: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := &FileMarksOpts{
				FileMarks: tt.marks,
				Relaxed:   tt.relaxed,
			}

			_, err := opts.Apply(tt.files)
			if tt.wantErr {
				require.ErrorContains(
					t,
					err,
					"to match at least one file by path",
				)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
