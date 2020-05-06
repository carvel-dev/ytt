package errors

type SemiStructuredError struct {
	err error
}

func NewSemiStructuredError(err error) SemiStructuredError {
	return SemiStructuredError{err}
}

func (e SemiStructuredError) Error() (result string) {
	// Be conservative in not swallowing underlying error if any
	// error processing fails (shouldn't happen, but let's be certain)
	defer func() {
		if rec := recover(); rec != nil {
			result = e.err.Error()
		}
	}()

	rootPiece, complete := NewErrorPiecesFromString(e.err.Error())
	if !complete {
		// Cannot deconstruct error message, return original
		return e.err.Error()
	}

	for _, piece := range rootPiece.Pieces {
		// k8s list of errors is wrapped with [] and separated by comma
		// (https://github.com/kubernetes/kubernetes/blob/a5e6ac0a959b059513c1e7908fbb0713467839c4/staging/src/k8s.io/apimachinery/pkg/util/errors/errors.go#L64)
		if piece.ContainerType == braketOpen {
			piece.ReorganizePiecesAroundCommas = true
		}
	}

	rootPiece.ReorganizePieces()

	return rootPiece.AsString()
}
