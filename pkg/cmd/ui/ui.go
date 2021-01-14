// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"io"
)

type UI interface {
	Printf(string, ...interface{})
	Debugf(string, ...interface{})
	Warnf(str string, args ...interface{})
	DebugWriter() io.Writer
}
