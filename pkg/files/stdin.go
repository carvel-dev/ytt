// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"io/ioutil"
	"os"
)

var hasStdinBeenRead bool

//ReadStdin only read stdin once
func ReadStdin() ([]byte, error) {
	if hasStdinBeenRead {
		return nil, fmt.Errorf("Standard input has been already read once (has the '-' argument been used in more than one flags?)")
	}
	hasStdinBeenRead = true
	return ioutil.ReadAll(os.Stdin)
}
