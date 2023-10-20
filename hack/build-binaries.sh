#!/bin/bash

set -e -x -u

function get_latest_git_tag {
  git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+'
}

VERSION="${1:-$(get_latest_git_tag)}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X carvel.dev/ytt/pkg/version.Version=$VERSION"

./hack/generate-website-assets.sh

GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o ytt-darwin-amd64 ./cmd/ytt
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -trimpath -o ytt-darwin-arm64 ./cmd/ytt
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o ytt-linux-amd64 ./cmd/ytt
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -trimpath -o ytt-linux-arm64 ./cmd/ytt
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o ytt-windows-amd64.exe ./cmd/ytt

shasum -a 256 ./ytt-darwin-amd64 ./ytt-darwin-arm64 ./ytt-linux-amd64 ./ytt-linux-arm64 ./ytt-windows-amd64.exe > ./go-checksums
