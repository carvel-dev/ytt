#!/bin/bash

set -e

./hack/test-unit.sh
./hack/test-e2e.sh

echo ALL SUCCESS
