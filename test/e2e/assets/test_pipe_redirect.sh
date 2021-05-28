#!/bin/bash

set -e -x

# test pipe redirect (on Linux, pipe is symlinked)
diff <(../../ytt -f pipe.yml=<(cat ../../examples/k8s-relative-rolling-update/config.yml)) ../../examples/k8s-relative-rolling-update/expected.txt