package files

import (
	"io"
)

type UI interface {
	Printf(string, ...interface{})
	Debugf(string, ...interface{})
	DebugWriter() io.Writer
}
