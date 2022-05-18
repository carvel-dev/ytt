#!/usr/bin/env bash

set -e

./ytt -f examples/data-values-directory/config/ \
      --data-values-file examples/data-values-directory/values/
