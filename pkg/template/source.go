// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"

	"github.com/k14s/ytt/pkg/filepos"
)

type Line struct {
	Instruction Instruction
	SourceLine  *SourceLine
}

type SourceLine struct {
	Position  *filepos.Position
	Content   string
	Selection *SourceLine
}

func NewCodeFromBytes(bs []byte, instructions *InstructionSet) []Line {
	return NewCodeFromBytesAtPosition(bs, filepos.NewPosition(1), instructions)
}

func NewCodeFromBytesAtPosition(bs []byte, pos *filepos.Position, instructions *InstructionSet) []Line {
	var result []Line

	for i, line := range bytes.Split(bs, []byte("\n")) {
		result = append(result, Line{
			Instruction: instructions.NewCode(string(line)),
			SourceLine:  NewSourceLine(pos.DeepCopyWithLineOffset(i), string(line)),
		})
	}

	return result
}

func NewSourceLine(pos *filepos.Position, content string) *SourceLine {
	if !pos.IsKnown() {
		panic("Expected source line position to be known")
	}
	return &SourceLine{Position: pos, Content: content}
}
