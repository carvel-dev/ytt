#!/bin/bash

set -e -x

diff <(../../examples/data-values/run.sh) ../../examples/data-values/expected.txt