// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"
	"io"
	"os"
)

type tty struct {
	debug  bool
	stdout io.Writer
	stderr io.Writer
}

var _ UI = tty{}

func NewTTY(debug bool) tty {
	return tty{debug, os.Stdout, os.Stderr}
}

func (t tty) Printf(str string, args ...interface{}) {
	fmt.Fprintf(t.stdout, str, args...)
}

func (t tty) Warnf(str string, args ...interface{}) {
	fmt.Fprintf(t.stderr, str, args...)
}

func (t tty) Debugf(str string, args ...interface{}) {
	if t.debug {
		fmt.Fprintf(t.stderr, str, args...)
	}
}

func (t tty) DebugWriter() io.Writer {
	if t.debug {
		return os.Stderr
	}
	return noopWriter{}
}

type noopWriter struct{}

var _ io.Writer = noopWriter{}

func (w noopWriter) Write(data []byte) (int, error) { return len(data), nil }
