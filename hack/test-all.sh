#!/bin/bash

set -e

./hack/build.sh

go fmt github.com/k14s/ytt/...

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

go test -v `go list ./...|grep -v yaml.v2` "$@"

echo ALL SUCCESS
