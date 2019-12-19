#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

BUILD_VALUES= ./hack/build.sh

GOOS=darwin GOARCH=amd64 go build $repro_flags -o ytt-darwin-amd64 ./cmd/ytt
GOOS=linux GOARCH=amd64 go build $repro_flags -o ytt-linux-amd64 ./cmd/ytt
GOOS=windows GOARCH=amd64 go build $repro_flags -o ytt-windows-amd64.exe ./cmd/ytt

shasum -a 256 ./ytt-*-amd64*
