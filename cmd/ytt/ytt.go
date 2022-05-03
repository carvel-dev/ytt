// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	uierrs "github.com/cppforlife/go-cli-ui/errors"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd"
	"github.com/vmware-tanzu/carvel-ytt/pkg/feature"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	enableExperimentalFeatures()

	command := cmd.NewDefaultYttCmd()

	err := command.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ytt: Error: %s\n", uierrs.NewMultiLineError(err))
		os.Exit(1)
	}
}

func enableExperimentalFeatures() {
	if experiments, isSet := os.LookupEnv("YTTEXPERIMENTS"); isSet && experiments != "" {
		// if we didn't ignore empty string, we'd attempt to Enable(""); that would panic (rather than ignore).
		for _, experiment := range strings.Split(experiments, ",") {
			feature.Flags().Enable(experiment)
			fmt.Fprintf(os.Stderr, "Experimental feature %q enabled.\n", experiment)
		}
	}
}
