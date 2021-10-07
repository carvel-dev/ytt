#!/bin/bash

set -e

./hack/build.sh

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

go test -v `go list ./...|grep -v yaml.v2` "$@"

echo ALL SUCCESS
