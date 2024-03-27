// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlfmt

import (
	"fmt"
	"io"
	"strings"
)

type writer struct {
	writer    io.Writer
	lastChunk writerChunk
}

type writerChunk struct {
	Content        string
	Indent         string
	AllowsInlining bool
	InliningSpacer string
	CanBeInlined   bool
	Spacer         bool
}

func newWriter(w io.Writer) *writer {
	return &writer{writer: w}
}

func (w *writer) AddContent(chunk writerChunk) {
	defer func() {
		w.lastChunk = chunk
	}()

	if chunk.Spacer {
		return
	}

	if w.lastChunk.Spacer {
		fmt.Fprintf(w.writer, "\n")
	}

	if w.lastChunk.AllowsInlining {
		if !chunk.CanBeInlined {
			fmt.Fprintf(w.writer, "\n")
			fmt.Fprintf(w.writer, chunk.Indent)
		} else {
			fmt.Fprintf(w.writer, w.lastChunk.InliningSpacer)
			// continue with content
		}
	} else {
		fmt.Fprintf(w.writer, chunk.Indent)
	}

	fmt.Fprintf(w.writer, "%s", w.indentMultiline(chunk))

	if !chunk.AllowsInlining {
		fmt.Fprintf(w.writer, "\n")
	}
}

func (w *writer) indentMultiline(chunk writerChunk) string {
	if !strings.Contains(chunk.Content, "\n") {
		return chunk.Content
	}

	result := []string{}
	for i, piece := range strings.Split(chunk.Content, "\n") {
		if i != 0 {
			piece = chunk.Indent[0:len(chunk.Indent)-2] + piece
		}
		result = append(result, piece)
	}
	return strings.Join(result, "\n")
}
