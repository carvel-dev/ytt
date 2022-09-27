#!/bin/bash

set -e

./hack/build.sh

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

go test ./... "$@"

# run a "contract test" to smoke Go module integration
(
  cd examples/integrating-with-ytt/internal-templating
  # in the case where a dependency is being bumped, update that dependency in this 👆 module.
  go mod tidy
  go test ./...
)

# error out if -run is given but no test is run
if [[ "$@" == *"-run "* ]]; then
  num_pkgs_with_tests=$( go test ./... "$@"  | grep "^\(ok  \|FAIL\)\tgithub.com" | grep -v "no test" | wc -l )
  if [[ num_pkgs_with_tests -eq 0 ]]; then
    echo
    echo "NO TESTS RUN"
    echo "  go test ./... "$@""
    exit 1
  fi
fi

echo ALL SUCCESS
