// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"carvel.dev/ytt/pkg/cmd"
	uierrs "github.com/cppforlife/go-cli-ui/errors"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := cmd.NewDefaultYttCmd()

	err := command.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ytt: Error: %s\n", uierrs.NewMultiLineError(err))
		os.Exit(1)
	}
}
