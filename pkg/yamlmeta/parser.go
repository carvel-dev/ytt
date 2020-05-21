package yamlmeta

// TODO json repr inside yaml?

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/yamlmeta/internal/yaml.v2"
)

var (
	docStartMarkerCheck = regexp.MustCompile(`\A\s*---\s+`)

	// eg "yaml: line 2: found character that cannot start any token"
	lineErrRegexp = regexp.MustCompile(`^(?P<prefix>yaml: line )(?P<num>\d+)(?P<suffix>: .+)$`)
)

type ParserOpts struct {
	WithoutMeta bool
	Strict      bool
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

	// YAML library uses 0 based line numbers for nodes
	// so by default move the numbering by one
	nodeLineCorrection := 1
	// Errors seem to use 1 based line numbers
	errLineCorrection := 0
	startsWithDocMarker := docStartMarkerCheck.Match(data)

	if !startsWithDocMarker {
		// Since we are adding doc marker that takes up first line
		// (0th line from YAML library perspective),
		// there is no need to line correct
		nodeLineCorrection = 0
		// For errors though, we do need to correct
		errLineCorrection = -1
		data = append([]byte("---\n"), data...)
	}

	docSet, err := p.parseBytes(data, nodeLineCorrection)
	if err != nil {
		return docSet, p.correctLineInErr(err, errLineCorrection)
	}

	// Change first document's line number to be 1
	// since we always present line numbers as 1 based
	// (note that first doc marker may be several lines down)
	if !startsWithDocMarker && !docSet.Items[0].Position.IsKnown() {
		docSet.Items[0].Position = filepos.NewPosition(1)
		docSet.Items[0].Position.SetFile(associatedName)
	}

	return docSet, nil
}

func (p *Parser) parseBytes(data []byte, lineCorrection int) (*DocumentSet, error) {
	docSet := &DocumentSet{Position: filepos.NewUnknownPosition()}

	var lastUnassingedMetas []*Meta

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.SetForceMapSlice(true)

	if p.opts.Strict {
		dec.SetStrictScalarResolve()
	}

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
			Metas:    lastUnassingedMetas,
			Value:    p.parse(rawVal, lineCorrection),
			Position: p.newDocPosition(dec.DocumentStartLine(), lineCorrection, len(docSet.Items) == 0),
		}

		allMetas, unassignedMetas := p.assignMetas(doc, dec.Comments(), lineCorrection)
		docSet.AllMetas = append(docSet.AllMetas, allMetas...)
		lastUnassingedMetas = unassignedMetas

		docSet.Items = append(docSet.Items, doc)
	}

	if len(lastUnassingedMetas) > 0 {
		endDoc := &Document{
			Metas:    lastUnassingedMetas,
			Value:    nil,
			Position: filepos.NewUnknownPosition(),
			injected: true,
		}
		docSet.Items = append(docSet.Items, endDoc)
	}

	return docSet, nil
}

func (p *Parser) parse(val interface{}, lineCorrection int) interface{} {
	switch typedVal := val.(type) {
	case yaml.MapSlice:
		result := &Map{Position: filepos.NewUnknownPosition()}
		for _, item := range typedVal {
			result.Items = append(result.Items, &MapItem{
				Key:      item.Key,
				Value:    p.parse(item.Value, lineCorrection),
				Position: p.newPosition(item.Line, lineCorrection),
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
					Value:    p.parse(typedItem.Value, lineCorrection),
					Position: p.newPosition(typedItem.Line, lineCorrection),
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

func (p *Parser) assignMetas(val interface{}, comments []yaml.Comment, lineCorrection int) ([]*Meta, []*Meta) {
	if p.opts.WithoutMeta {
		return nil, nil
	}

	nodesAtLines := map[int][]Node{}
	p.buildLineLocs(val, nodesAtLines)

	lineNums := p.buildLineNums(nodesAtLines)
	allMetas := []*Meta{}
	unassignedMetas := []*Meta{}

	for _, comment := range comments {
		meta := &Meta{
			Data:     comment.Data,
			Position: p.newPosition(comment.Line, lineCorrection),
		}
		allMetas = append(allMetas, meta)

		var foundOwner bool

		for _, lineNum := range lineNums {
			// Always looking at the same line or "above" (greater line number)
			if meta.Position.Line() > lineNum {
				continue
			}
			nodes, ok := nodesAtLines[lineNum]
			if ok {
				// Last node on the line is the one that owns inline meta
				// otherwise it's the first one (outermost one)
				// TODO any other better way to determine?
				if meta.Position.Line() == lineNum {
					nodes[len(nodes)-1].addMeta(meta)
				} else {
					nodes[0].addMeta(meta)
				}
				foundOwner = true
				break
			}
		}

		if !foundOwner {
			unassignedMetas = append(unassignedMetas, meta)
		}
	}

	return allMetas, unassignedMetas
}

func (p *Parser) buildLineLocs(val interface{}, nodeAtLines map[int][]Node) {
	if node, ok := val.(Node); ok {
		if node.GetPosition().IsKnown() {
			nodeAtLines[node.GetPosition().Line()] = append(nodeAtLines[node.GetPosition().Line()], node)
		}

		for _, childVal := range node.GetValues() {
			p.buildLineLocs(childVal, nodeAtLines)
		}
	}
}

func (p *Parser) buildLineNums(nodeAtLines map[int][]Node) []int {
	var result []int
	for lineNum, _ := range nodeAtLines {
		result = append(result, lineNum)
	}
	sort.Ints(result)
	return result
}

func (p *Parser) correctLineInErr(err error, correction int) error {
	submatches := lineErrRegexp.FindAllStringSubmatch(err.Error(), -1)
	if len(submatches) != 1 || len(submatches[0]) != 4 {
		return err
	}

	origLine, parseErr := strconv.Atoi(submatches[0][2])
	if parseErr != nil {
		return err
	}

	return fmt.Errorf("%s%d%s", submatches[0][1], p.newPosition(origLine, correction).Line(), submatches[0][3])
}

func (p *Parser) newDocPosition(actualLine, correction int, firstDoc bool) *filepos.Position {
	if firstDoc && actualLine+correction == 0 {
		return p.newUnknownPosition()
	}
	return p.newPosition(actualLine, correction)
}

func (p *Parser) newPosition(actualLine, correction int) *filepos.Position {
	pos := filepos.NewPosition(actualLine + correction)
	pos.SetFile(p.associatedName)
	return pos
}

func (p *Parser) newUnknownPosition() *filepos.Position {
	pos := filepos.NewUnknownPosition()
	pos.SetFile(p.associatedName)
	return pos
}
