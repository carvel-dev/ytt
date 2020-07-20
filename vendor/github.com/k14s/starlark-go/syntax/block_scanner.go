package syntax

import (
	"fmt"
)

var _ = fmt.Sprintf

type scannerInterface interface {
	nextToken(*tokenValue) Token
	error(pos Position, s string)
	errorf(pos Position, format string, args ...interface{})
	recover(*error)

	getLineComments() []Comment
	getSuffixComments() []Comment
	getPos() Position
}

// blockScanner changes INDENT/OUTDENT to be
// based on nesting depth (start->end) instead of whitespace
type blockScanner struct {
	scanner     *scanner
	nextTokens  []blockScannerToken
	prevTokens  []blockScannerToken
	indentStack []blockScannerToken
	debug       bool
}

var _ scannerInterface = &blockScanner{}

type blockScannerToken struct {
	val tokenValue
	tok Token
	alreadyOutdented bool
}

func newBlockScanner(s *scanner) *blockScanner {
	return &blockScanner{s, nil, nil, nil, false}
}

func (s *blockScanner) nextToken(outVal *tokenValue) Token {
	pair := s.nextTokenInner()
	s.prevTokens = append(s.prevTokens, pair)

	if s.debug {
		fmt.Printf("emit: %s => %#v\n", pair.tok.String(), pair.val)
	}

	s.copyTokenValue(pair.val, outVal)
	return pair.tok
}

func (s *blockScanner) nextTokenInner() blockScannerToken {
	if s.matchesNewBlock() {
		return s.buildIndent()
	}

	var currToken blockScannerToken
	var tokSource string

	if len(s.nextTokens) > 0 {
		tokSource = "read-buffer"
		currToken = s.nextTokens[0]
		s.nextTokens = s.nextTokens[1:]
	} else {
		tokSource = "read"
		currToken = s.popNextToken()
	}

	if s.debug {
		fmt.Printf("\n%s: %s => %#v\n", tokSource, currToken.tok.String(), currToken.val)
	}

	switch currToken.tok {
	// 'else' is a special cases when we need to
	// implicitly outdent since end is not specified in code
	case ELSE:
		if !currToken.alreadyOutdented {
			// if 'else' is not followed by the colon assume
			// this is an inline if-else hence no need to outdent
			maybeColonToken := s.popNextToken()

			if maybeColonToken.tok == COLON {
				currToken.alreadyOutdented = true
				s.putBackToken(currToken)
				s.putBackToken(maybeColonToken)
				return s.buildOutdent()
			}

			s.putBackToken(maybeColonToken)
		}

	// 'elif' is special cases when we need to
	// implicitly outdent since end is not specified in code
	case ELIF:
		if !currToken.alreadyOutdented {
			currToken.alreadyOutdented = true
			s.putBackToken(currToken)
			return s.buildOutdent()
		}

	// Skip parsed indent/outdent as we insert
	// our own "indention" at appropriate times
	case INDENT, OUTDENT:
		return s.nextTokenInner()

	case PASS:
		s.errorf(s.getPos(), "use of reserved keyword 'pass' is not allowed")

	// 'end' is identifier
	case IDENT:
		if currToken.val.raw == "end" {
			s.swallowNextToken(NEWLINE)
			return s.buildOutdent()
		}

	case EOF:
		if len(s.indentStack) != 0 {
			pos := s.indentStack[len(s.indentStack)-1].val.pos
			s.errorf(pos, "mismatched set of block openings (if/else/elif/for/def) and closing (end)")
		}

	default:
		// continue with curr token
	}

	return currToken
}

func (s *blockScanner) popNextToken() blockScannerToken {
	val := tokenValue{}
	tok := s.scanner.nextToken(&val)
	return blockScannerToken{tok: tok, val: val}
}

func (s *blockScanner) putBackToken(pair blockScannerToken) blockScannerToken {
	s.nextTokens = append(s.nextTokens, pair)
	return pair
}

func (s *blockScanner) swallowNextToken(tok Token) {
	token := s.popNextToken()
	if token.tok != tok {
		s.putBackToken(token)
	}
}

func (s *blockScanner) buildIndent() blockScannerToken {
	s.indentStack = append(s.indentStack, s.prevTokens[len(s.prevTokens)-1])
	return blockScannerToken{
		tok: Token(INDENT),
		val: tokenValue{pos: s.prevTokens[len(s.prevTokens)-1].val.pos},
	}
}

func (s *blockScanner) buildOutdent() blockScannerToken {
	if len(s.indentStack) == 0 {
		s.error(s.getPos(), "unexpected end")
	}
	s.indentStack = s.indentStack[:len(s.indentStack)-1]
	return blockScannerToken{
		tok: Token(OUTDENT),
		val: tokenValue{pos: s.prevTokens[len(s.prevTokens)-1].val.pos},
	}
}

func (s *blockScanner) matchesNewBlock() bool {
	if len(s.prevTokens) < 2 {
		return false
	}
	lastLastColon := s.prevTokens[len(s.prevTokens)-2].tok == COLON
	lastNewline := s.prevTokens[len(s.prevTokens)-1].tok == NEWLINE
	return lastLastColon && lastNewline
}

func (s *blockScanner) copyTokenValue(left tokenValue, right *tokenValue) {
	right.raw = left.raw
	right.int = left.int
	right.bigInt = left.bigInt
	right.float = left.float
	right.string = left.string
	right.pos = left.pos
}

// implement boring scanner methods
func (s *blockScanner) error(pos Position, str string) {
	s.scanner.error(pos, str)
}

func (s *blockScanner) errorf(pos Position, format string, args ...interface{}) {
	s.scanner.errorf(pos, format, args...)
}

func (s *blockScanner) recover(err *error) {
	s.scanner.recover(err)
}

func (s *blockScanner) getLineComments() []Comment {
	return s.scanner.getLineComments()
}

func (s *blockScanner) getSuffixComments() []Comment {
	return s.scanner.getSuffixComments()
}

func (s *blockScanner) getPos() Position { return s.scanner.getPos() }

// augment regular scanner
var _ scannerInterface = &scanner{}

func (s *scanner) getLineComments() []Comment { return s.lineComments }
func (s *scanner) getSuffixComments() []Comment { return s.suffixComments }
func (s *scanner) getPos() Position { return s.pos }
