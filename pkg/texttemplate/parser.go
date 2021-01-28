// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package texttemplate

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
)

type Parser struct {
	associatedName string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(dataBs []byte, associatedName string) (*NodeRoot, error) {
	return p.parse(dataBs, associatedName, filepos.NewPosition(1))
}

func (p *Parser) ParseWithPosition(dataBs []byte, associatedName string, startPos *filepos.Position) (*NodeRoot, error) {
	return p.parse(dataBs, associatedName, startPos)
}

func (p *Parser) parse(dataBs []byte, associatedName string, startPos *filepos.Position) (*NodeRoot, error) {
	p.associatedName = associatedName

	var lastChar rune
	var currLine int = 1
	var currCol int = 1

	if startPos.IsKnown() {
		currLine = startPos.LineNum()
	}

	var lastNode interface{} = &NodeText{Position: p.newPosition(currLine)}
	var nodes []interface{}

	data := string(dataBs)

	for i, currChar := range data {
		if lastChar == '(' && currChar == '@' {
			switch typedLastNode := lastNode.(type) {
			case *NodeText:
				typedLastNode.Content = data[typedLastNode.startOffset : i-1]
				nodes = append(nodes, lastNode)
				lastNode = &NodeCode{
					Position:    p.newPosition(currLine),
					startOffset: i + 1,
				}
			case *NodeCode:
				return nil, fmt.Errorf(
					"Unexpected code opening '(@' at line %d col %d", currLine, currCol)
			default:
				panic(fmt.Sprintf("unknown string template piece %T", typedLastNode))
			}
		}

		if lastChar == '@' && currChar == ')' {
			switch typedLastNode := lastNode.(type) {
			case *NodeText:
				return nil, fmt.Errorf(
					"Unexpected code closing '@)' at line %d col %d", currLine, currCol)
			case *NodeCode:
				typedLastNode.Content = data[typedLastNode.startOffset : i-1]
				nodes = append(nodes, lastNode)
				lastNode = &NodeText{
					Position:    p.newPosition(currLine),
					startOffset: i + 1,
				}
			default:
				panic(fmt.Sprintf("unknown string template piece %T", typedLastNode))
			}
		}

		if currChar == '\n' {
			currLine++
			currCol = 1
		} else {
			currCol++
		}

		lastChar = currChar
	}

	// close last node
	switch typedLastNode := lastNode.(type) {
	case *NodeText:
		typedLastNode.Content = data[typedLastNode.startOffset:len(data)]
		nodes = append(nodes, lastNode)
	case *NodeCode:
		return nil, fmt.Errorf(
			"Missing code closing '@)' at line %d col %d", currLine, currCol)
	default:
		panic(fmt.Sprintf("unknown string template piece %T", typedLastNode))
	}

	return &NodeRoot{Items: nodes}, nil
}

func (p *Parser) newPosition(line int) *filepos.Position {
	pos := filepos.NewPosition(line)
	pos.SetFile(p.associatedName)
	return pos
}
