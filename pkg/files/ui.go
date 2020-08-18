// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"io"
)

type UI interface {
	Printf(string, ...interface{})
	Debugf(string, ...interface{})
	DebugWriter() io.Writer
}
