// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package filepos

import (
	"fmt"
)

type Position struct {
	lineNum    *int // 1 based
	file       string
	line       string
	known      bool
	fromMemory bool
}

func NewPosition(lineNum int) *Position {
	if lineNum <= 0 {
		panic("Lines are 1 based")
	}
	return &Position{lineNum: &lineNum, known: true}
}

// NewPositionInFile returns the Position of line "lineNum" within the file "file"
func NewPositionInFile(lineNum int, file string) *Position {
	p := NewPosition(lineNum)
	p.file = file
	return p
}

// NewUnknownPosition is equivalent of zero value *Position
func NewUnknownPosition() *Position {
	return &Position{}
}

// NewUnknownPositionInFile produces a Position of a known file at an unknown line.
func NewUnknownPositionInFile(file string) *Position {
	return &Position{file: file}
}

func NewUnknownPositionWithKeyVal(k, v interface{}, separator string) *Position {
	return &Position{line: fmt.Sprintf("%v%v %#v", k, separator, v), fromMemory: true}
}

func (p *Position) SetLine(line string) { p.line = line }

func (p *Position) IsKnown() bool { return p != nil && p.known }

func (p *Position) FromMemory() bool { return p.fromMemory }

func (p *Position) LineNum() int {
	if !p.IsKnown() {
		panic("Position is unknown")
	}
	if p.lineNum == nil {
		panic("Position was not properly initialized")
	}
	return *p.lineNum
}

func (p *Position) GetLine() string {
	return p.line
}

func (p *Position) AsString() string {
	return "line " + p.AsCompactString()
}

func (p *Position) GetFile() string {
	return p.file
}

func (p *Position) AsCompactString() string {
	filePrefix := p.file
	if len(filePrefix) > 0 {
		filePrefix += ":"
	}
	if p.IsKnown() {
		return fmt.Sprintf("%s%d", filePrefix, p.LineNum())
	}
	return fmt.Sprintf("%s?", filePrefix)
}

func (p *Position) AsIntString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%d", p.LineNum())
	}
	return "?"
}

func (p *Position) As4DigitString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%4d", p.LineNum())
	}
	return "????"
}

func (p *Position) DeepCopy() *Position {
	if p == nil {
		return nil
	}
	newPos := &Position{file: p.file, known: p.known, line: p.line}
	if p.lineNum != nil {
		lineVal := *p.lineNum
		newPos.lineNum = &lineVal
	}
	return newPos
}

func (p *Position) DeepCopyWithLineOffset(offset int) *Position {
	if !p.IsKnown() {
		panic("Position is unknown")
	}
	if offset < 0 {
		panic("Unexpected line offset")
	}
	newPos := p.DeepCopy()
	*newPos.lineNum += offset
	return newPos
}

// IsNextTo compares the location of one position with another.
func (p *Position) IsNextTo(otherPosition *Position) bool {
	if p.IsKnown() && otherPosition.IsKnown() {
		if p.GetFile() == otherPosition.GetFile() {
			diff := p.LineNum() - otherPosition.LineNum()
			if -1 <= diff && 1 >= diff {
				return true
			}
		}
	}
	return false
}
