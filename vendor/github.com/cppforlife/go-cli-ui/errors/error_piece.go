package errors

import (
	"regexp"
	"strings"
)

const (
	braketOpen = "["
	parenOpen  = "("
	curlyOpen  = "{"
	rootOpen   = "*"
	customOpen = "~"
	rawOpen    = "...|"

	braketClose = "]"
	parenClose  = ")"
	curlyClose  = "}"
	rawClose    = "|..."
)

var (
	sepOpeners = map[string]string{
		braketOpen: string(braketOpen),
		parenOpen:  string(parenOpen),
		curlyOpen:  string(curlyOpen),
		rootOpen:   "",
		customOpen: "",
		rawOpen:    "",
	}
	sepClosers = map[string]string{
		braketOpen: string(braketClose),
		parenOpen:  string(parenClose),
		curlyOpen:  string(curlyClose),
		rootOpen:   "",
		customOpen: "",
		rawOpen:    "",
	}
)

type ErrorPiece struct {
	Value string

	ContainerType string // eg *,{,[,(
	Pieces        []*ErrorPiece

	ReorganizePiecesAroundCommas bool
	FormatPiecesAsList           bool

	closed bool
}

func NewErrorPiecesFromString(str string) (*ErrorPiece, bool) {
	rootPiece := &ErrorPiece{
		ContainerType: rootOpen,
	}
	stack := []*ErrorPiece{rootPiece}

	for i, f := range str {
		if len(stack) == 0 {
			return rootPiece, false
		}

		charStr := string(f)

		if !stack[len(stack)-1].IsLeafContainer() {
			switch charStr {
			case braketOpen, curlyOpen, parenOpen:
				piece := &ErrorPiece{ContainerType: charStr}
				stack[len(stack)-1].AddPiece(piece)
				stack = append(stack, piece)
				continue

			case braketClose, curlyClose, parenClose:
				stack[len(stack)-1].closed = true
				stack = stack[:len(stack)-1]
				continue
			}
		}

		switch {
		case checkForward(str, i, rawOpen):
			piece := &ErrorPiece{ContainerType: rawOpen}
			piece.AddStr(charStr)
			stack[len(stack)-1].AddPiece(piece)
			stack = append(stack, piece)

		case checkBackward(str, i, rawClose):
			stack[len(stack)-1].AddStr(charStr)
			stack[len(stack)-1].closed = true
			stack = stack[:len(stack)-1]

		default:
			stack[len(stack)-1].AddStr(charStr)
		}
	}

	return rootPiece, len(stack) == 1
}

func checkForward(str string, i int, sep string) bool {
	endIdx := i + len(sep)
	if endIdx < len(str) {
		return str[i:endIdx] == sep
	}
	return false
}

func checkBackward(str string, i int, sep string) bool {
	startIdx := i - len(sep) + 1
	if startIdx >= 0 && i < len(str) {
		return str[startIdx:i+1] == sep
	}
	return false
}

func (p *ErrorPiece) IsContainer() bool {
	return len(p.ContainerType) > 0
}

func (p *ErrorPiece) IsLeafContainer() bool {
	return p.ContainerType == rawOpen
}

func (p *ErrorPiece) AddPiece(piece *ErrorPiece) {
	p.Pieces = append(p.Pieces, piece)
}

func (p *ErrorPiece) AddStr(str string) {
	var lastPiece *ErrorPiece
	if len(p.Pieces) > 0 {
		lastPiece = p.Pieces[len(p.Pieces)-1]
	}
	if lastPiece == nil || lastPiece.IsContainer() {
		lastPiece = &ErrorPiece{}
		p.Pieces = append(p.Pieces, lastPiece)
	}
	lastPiece.Value += str
}

var (
	listItemSep         = ", "
	likelyListItemStart = regexp.MustCompile(`^[a-z\.\[\]:\s]{3,10}`)
)

func (p *ErrorPiece) ReorganizePieces() {
	for _, piece := range p.Pieces {
		piece.ReorganizePieces()
	}

	if p.ReorganizePiecesAroundCommas {
		p.ReorganizePiecesAroundCommas = false

		var newPieces []*ErrorPiece

		for _, piece := range p.Pieces {
			switch {
			case len(piece.Value) > 0:
				for i, val := range strings.Split(piece.Value, listItemSep) {
					switch {
					case len(newPieces) > 0 && i == 0:
						newPieces[len(newPieces)-1].AddStr(val)

					case len(newPieces) > 0 && !likelyListItemStart.MatchString(val):
						newPieces[len(newPieces)-1].AddStr(listItemSep + val)

					default:
						newPieces = append(newPieces, &ErrorPiece{
							ContainerType: customOpen,
							Pieces:        []*ErrorPiece{{Value: val}},
						})
					}
				}

			case p.IsContainer():
				if len(newPieces) > 0 {
					newPieces[len(newPieces)-1].AddPiece(piece)
				} else {
					newPieces = append(newPieces, &ErrorPiece{
						ContainerType: customOpen,
						Pieces:        []*ErrorPiece{piece},
					})
				}
			}
		}

		if len(newPieces) > 1 {
			p.Pieces = newPieces
			p.FormatPiecesAsList = true
		}
	}
}

func (p *ErrorPiece) AsString() string {
	if len(p.Value) > 0 {
		return p.Value
	}

	var result string

	if p.FormatPiecesAsList {
		if len(p.Pieces) > 0 {
			result += "\n\n"
		}
		for _, piece := range p.Pieces {
			result += "  - " + piece.AsString() + "\n\n"
		}
	} else {
		result += sepOpeners[p.ContainerType]
		for _, piece := range p.Pieces {
			result += piece.AsString()
		}
		if p.closed {
			result += sepClosers[p.ContainerType]
		}
	}

	return result
}

func (p *ErrorPiece) AsIndentedString(indent string) string {
	if len(p.Value) > 0 {
		return indent + "+ " + p.Value + "\n"
	}
	var result string
	if p.IsContainer() {
		result += indent + "+ ...\n"
	}
	for _, piece := range p.Pieces {
		result += piece.AsIndentedString(indent + "  ")
	}
	return result
}
