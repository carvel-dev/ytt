#!/bin/bash

set -e

./hack/build.sh

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

go test ./... "$@"

echo ALL SUCCESS
