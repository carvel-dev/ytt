// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

// TODO json repr inside yaml?

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

var (
	docStartMarkerCheck = regexp.MustCompile(`\A\s*---\s+`)

	// eg "yaml: line 2: found character that cannot start any token"
	lineErrRegexp = regexp.MustCompile(`^(?P<prefix>yaml: line )(?P<num>\d+)(?P<suffix>: .+)$`)
)

type ParserOpts struct {
	WithoutComments bool
	Strict          bool
}

type Parser struct {
	opts           ParserOpts
	associatedName string
}

func NewParser(opts ParserOpts) *Parser {
	return &Parser{opts, ""}
}

func (p *Parser) ParseBytes(data []byte, associatedName string) (*DocumentSet, error) {
	p.associatedName = associatedName

	// YAML library uses 0-based line numbers for nodes (but, first line in a text file is typically line 1)
	nodeLineCorrection := 1
	// YAML library uses 1-based line numbers for errors
	errLineCorrection := 0
	startsWithDocMarker := docStartMarkerCheck.Match(data)

	if !startsWithDocMarker {
		data = append([]byte("---\n"), data...)

		// we just prepended a line to the original input, correct for that:
		nodeLineCorrection--
		errLineCorrection--
	}

	docSet, err := p.parseBytes(data, nodeLineCorrection)
	if err != nil {
		return docSet, p.correctLineNumInErr(err, errLineCorrection)
	}

	// Change first document's line number to be 1
	// since we always present line numbers as 1 based
	// (note that first doc marker may be several lines down)
	if !startsWithDocMarker && !docSet.Items[0].Position.IsKnown() {
		docSet.Items[0].Position = filepos.NewPosition(1)
		docSet.Items[0].Position.SetFile(associatedName)
	}
	setPositionOfCollections(docSet, nil)

	return docSet, nil
}

func (p *Parser) parseBytes(data []byte, lineCorrection int) (*DocumentSet, error) {
	docSet := &DocumentSet{Position: filepos.NewUnknownPosition()}

	var lastUnassignedComments []*Comment

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.SetForceMapSlice(true)

	if p.opts.Strict {
		dec.SetStrictScalarResolve()
	}

	lines := strings.Split(string(data), "\n")

	for {
		var rawVal interface{}

		err := dec.Decode(&rawVal)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		doc := &Document{
			Comments: lastUnassignedComments,
			Value:    p.parse(rawVal, lineCorrection, lines),
			Position: p.newDocPosition(dec.DocumentStartLine(), lineCorrection, len(docSet.Items) == 0, lines),
		}

		allComments, unassignedComments := p.assignComments(doc, dec.Comments(), lineCorrection)
		docSet.AllComments = append(docSet.AllComments, allComments...)
		lastUnassignedComments = unassignedComments

		docSet.Items = append(docSet.Items, doc)
	}

	if len(lastUnassignedComments) > 0 {
		endDoc := &Document{
			Comments: lastUnassignedComments,
			Value:    nil,
			Position: filepos.NewUnknownPosition(),
			injected: true,
		}
		docSet.Items = append(docSet.Items, endDoc)
	}

	return docSet, nil
}

// setPositionOfCollections assigns the Position of Maps and Arrays to their parent
//   these kinds of nodes are not visible and therefore technically don't have a position.
//   However, it is useful when communicating certain error cases to be able to reference
//   a collection by line number.
//   The position of the parent matches well with what the user sees. E.g. the MapItem that
//   holds an Array is a great place to point at when referring to the entire array.
func setPositionOfCollections(node Node, parent Node) {
	if !node.GetPosition().IsKnown() {
		if parent != nil {
			if mapNode, ok := node.(*Map); ok {
				mapNode.Position = parent.GetPosition()
			}
			if arrayNode, ok := node.(*Array); ok {
				arrayNode.Position = parent.GetPosition()
			}
		}
	}
	for _, val := range node.GetValues() {
		child, isNode := val.(Node)
		if isNode {
			setPositionOfCollections(child, node)
		}
	}
}

func (p *Parser) parse(val interface{}, lineCorrection int, lines []string) interface{} {
	switch typedVal := val.(type) {
	case yaml.MapSlice:
		result := &Map{Position: p.newUnknownPosition()}
		for _, item := range typedVal {
			result.Items = append(result.Items, &MapItem{
				Key:      item.Key,
				Value:    p.parse(item.Value, lineCorrection, lines),
				Position: p.newPosition(item.Line, lineCorrection, lines[item.Line]),
			})
		}
		return result

	// As a precaution against yaml library returning non-ordered maps
	case map[interface{}]interface{}:
		panic("Unexpected map[interface{}]interface{} when parsing YAML, expected MapSlice")

	case []interface{}:
		result := &Array{Position: filepos.NewUnknownPosition()}
		for _, item := range typedVal {
			if typedItem, ok := item.(yaml.ArrayItem); ok {
				result.Items = append(result.Items, &ArrayItem{
					Value:    p.parse(typedItem.Value, lineCorrection, lines),
					Position: p.newPosition(typedItem.Line, lineCorrection, lines[typedItem.Line]),
				})
			} else {
				panic("unknown item")
			}
		}
		return result

	default:
		return val
	}
}

func (p *Parser) assignComments(val interface{}, comments []yaml.Comment, lineCorrection int) ([]*Comment, []*Comment) {
	if p.opts.WithoutComments {
		return nil, nil
	}

	nodesAtLines := map[int][]Node{}
	p.buildLineLocs(val, nodesAtLines)

	lineNums := p.buildLineNums(nodesAtLines)
	allComments := []*Comment{}
	unassignedComments := []*Comment{}

	for _, comment := range comments {
		comment := &Comment{
			Data:     comment.Data,
			Position: p.newPosition(comment.Line, lineCorrection, comment.Data),
		}
		allComments = append(allComments, comment)

		var foundOwner bool

		for _, lineNum := range lineNums {
			// Always looking at the same line or "above" (greater line number)
			if comment.Position.LineNum() > lineNum {
				continue
			}
			nodes, ok := nodesAtLines[lineNum]
			if ok {
				// Last node on the line is the one that owns inline comment
				// otherwise it's the first one (outermost one)
				// TODO any other better way to determine?
				if comment.Position.LineNum() == lineNum {
					nodes[len(nodes)-1].addComments(comment)
				} else {
					nodes[0].addComments(comment)
				}
				foundOwner = true
				break
			}
		}

		if !foundOwner {
			unassignedComments = append(unassignedComments, comment)
		}
	}

	return allComments, unassignedComments
}

func (p *Parser) buildLineLocs(val interface{}, nodeAtLines map[int][]Node) {
	if node, ok := val.(Node); ok {
		if node.GetPosition().IsKnown() {
			nodeAtLines[node.GetPosition().LineNum()] = append(nodeAtLines[node.GetPosition().LineNum()], node)
		}

		for _, childVal := range node.GetValues() {
			p.buildLineLocs(childVal, nodeAtLines)
		}
	}
}

func (p *Parser) buildLineNums(nodeAtLines map[int][]Node) []int {
	var result []int
	for lineNum := range nodeAtLines {
		result = append(result, lineNum)
	}
	sort.Ints(result)
	return result
}

func (p *Parser) correctLineNumInErr(err error, correction int) error {
	submatches := lineErrRegexp.FindAllStringSubmatch(err.Error(), -1)
	if len(submatches) != 1 || len(submatches[0]) != 4 {
		return err
	}

	actualLineNum, parseErr := strconv.Atoi(submatches[0][2])
	if parseErr != nil {
		return err
	}

	return fmt.Errorf("%s%d%s", submatches[0][1], p.newPosition(actualLineNum, correction, "").LineNum(), submatches[0][3])
}

func (p *Parser) newDocPosition(actualLineNum, correction int, firstDoc bool, lines []string) *filepos.Position {
	if firstDoc && actualLineNum+correction == 0 {
		return p.newUnknownPosition()
	}
	return p.newPosition(actualLineNum, correction, lines[actualLineNum])
}

func (p *Parser) newPosition(actualLineNum, correction int, line string) *filepos.Position {
	pos := filepos.NewPosition(actualLineNum + correction)
	pos.SetFile(p.associatedName)
	pos.SetLine(line)
	return pos
}

func (p *Parser) newUnknownPosition() *filepos.Position {
	pos := filepos.NewUnknownPosition()
	pos.SetFile(p.associatedName)
	return pos
}
