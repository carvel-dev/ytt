package filepos

import (
	"fmt"
)

type Position struct {
	line  *int // 1 based
	file  string
	known bool
}

func NewPosition(line int) *Position {
	if line <= 0 {
		panic("Lines are 1 based")
	}
	return &Position{line: &line, known: true}
}

// NewUnknownPosition is equivalent of zero value *Position
func NewUnknownPosition() *Position {
	return &Position{}
}

func (p *Position) SetFile(file string) { p.file = file }

func (p *Position) IsKnown() bool { return p != nil && p.known }

func (p *Position) Line() int {
	if !p.IsKnown() {
		panic("Position is unknown")
	}
	if p.line == nil {
		panic("Position was not properly initialized")
	}
	return *p.line
}

func (p *Position) AsString() string {
	return "line " + p.AsCompactString()
}

func (p *Position) AsCompactString() string {
	filePrefix := p.file
	if len(filePrefix) > 0 {
		filePrefix += ":"
	}
	if p.IsKnown() {
		return fmt.Sprintf("%s%d", filePrefix, p.Line())
	}
	return fmt.Sprintf("%s?", filePrefix)
}

func (p *Position) AsIntString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%d", p.Line())
	}
	return "?"
}

func (p *Position) As4DigitString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%4d", p.Line())
	}
	return "????"
}

func (p *Position) DeepCopy() *Position {
	if p == nil {
		return nil
	}
	newPos := &Position{file: p.file, known: p.known}
	if p.line != nil {
		lineVal := *p.line
		newPos.line = &lineVal
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
	*newPos.line += offset
	return newPos
}
