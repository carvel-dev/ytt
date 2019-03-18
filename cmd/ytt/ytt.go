package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/k14s/ytt/pkg/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := cmd.NewDefaultYttCmd()

	err := command.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
