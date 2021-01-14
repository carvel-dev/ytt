// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"
	"io"
	"os"
)

type TTY struct {
	debug  bool
	stdout io.Writer
	stderr io.Writer
}

var _ UI = TTY{}

func NewTTY(debug bool) TTY {
	return TTY{debug, os.Stdout, os.Stderr}
}

func (t TTY) Printf(str string, args ...interface{}) {
	fmt.Fprintf(t.stdout, str, args...)
}

func (t TTY) Warnf(str string, args ...interface{}) {
	fmt.Fprintf(t.stderr, str, args...)
}

func (t TTY) Debugf(str string, args ...interface{}) {
	if t.debug {
		fmt.Fprintf(t.stderr, str, args...)
	}
}

func (t TTY) DebugWriter() io.Writer {
	if t.debug {
		return os.Stderr
	}
	return noopWriter{}
}

type noopWriter struct{}

var _ io.Writer = noopWriter{}

func (w noopWriter) Write(data []byte) (int, error) { return len(data), nil }

// Used for testing whether TTY writes correct output to stdout/stderr
func NewCustomWriterTTY(debug bool, stdout, stderr io.Writer) TTY {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return TTY{debug, stdout, stderr}
}
