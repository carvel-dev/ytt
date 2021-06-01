#!/bin/bash

set -e -x

diff <(cat ../../examples/k8s-relative-rolling-update/config.yml | ../../ytt -f-) ../../examples/k8s-relative-rolling-update/expected.txt