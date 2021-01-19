#!/bin/bash

set -e -x -u

./hack/build.sh
./hack/linter.sh
