package core

import (
	"fmt"
	"io"
	"os"

	"github.com/k14s/ytt/pkg/files"
)

type PlainUI struct {
	debug bool
}

var _ files.UI = PlainUI{}

func NewPlainUI(debug bool) PlainUI { return PlainUI{debug} }

func (ui PlainUI) Printf(str string, args ...interface{}) {
	fmt.Printf(str, args...)
}

func (ui PlainUI) Debugf(str string, args ...interface{}) {
	if ui.debug {
		fmt.Fprintf(os.Stderr, str, args...)
	}
}

func (ui PlainUI) DebugWriter() io.Writer {
	if ui.debug {
		return os.Stderr
	}
	return noopWriter{}
}

type noopWriter struct{}

var _ io.Writer = noopWriter{}

func (w noopWriter) Write(data []byte) (int, error) { return len(data), nil }
