#!/bin/bash

set -e -x -u

GOOS=darwin GOARCH=amd64 go build -o ytt-darwin-amd64 ./cmd/ytt
GOOS=linux GOARCH=amd64 go build -o ytt-linux-amd64 ./cmd/ytt
GOOS=windows GOARCH=amd64 go build -o ytt-windows-amd64.exe ./cmd/ytt

shasum -a 256 ./ytt-*-amd64*
