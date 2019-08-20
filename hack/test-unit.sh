#!/bin/bash

set -e -x

go fmt github.com/k14s/ytt/...

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

go test -v `go list ./...|grep -v yaml.v2` "$@"

echo UNIT SUCCESS
