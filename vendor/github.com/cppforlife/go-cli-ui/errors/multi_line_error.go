package errors

import (
	"regexp"
	"strings"
)

var (
	errColonSep = regexp.MustCompile("(: )([A-Z])")
)

type MultiLineError struct {
	err error
}

func NewMultiLineError(err error) MultiLineError {
	return MultiLineError{err}
}

func (e MultiLineError) Error() (result string) {
	// Be conservative in not swallowing underlying error if any
	// error processing fails (shouldn't happen, but let's be certain)
	defer func() {
		if rec := recover(); rec != nil {
			result = e.err.Error()
		}
	}()

	lines := strings.Split(e.err.Error(), "\n")

	firstLineRootPiece, _ := NewErrorPiecesFromString(lines[0])

	for _, piece := range firstLineRootPiece.Pieces {
		// Do not split withing container pieces
		if len(piece.Value) > 0 {
			piece.Value = errColonSep.ReplaceAllString(piece.Value, ":\n$2")
		}
	}

	var firstLines []string
	for i, chunk := range strings.Split(firstLineRootPiece.AsString(), "\n") {
		firstLines = append(firstLines, strings.Repeat("  ", i)+chunk)
	}

	lines[0] = strings.Join(firstLines, "\n")

	return strings.Join(lines, "\n")
}
