// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"io"
	"os"
)

var hasStdinBeenRead bool

// ReadStdin only read stdin once
func ReadStdin() ([]byte, error) {
	if hasStdinBeenRead {
		return nil, fmt.Errorf("Standard input has already been read, has the '-' argument been used in more than one flag?")
	}
	hasStdinBeenRead = true
	return io.ReadAll(os.Stdin)
}
