#!/bin/bash

set -e

./ytt -f examples/data-values-multiple-envs/config/ -v version=123
echo '***'
./ytt -f examples/data-values-multiple-envs/config/ -f examples/data-values-multiple-envs/envs/dev.yml
echo '***'
./ytt -f examples/data-values-multiple-envs/config/ -f examples/data-values-multiple-envs/envs/staging.yml
echo '***'
./ytt -f examples/data-values-multiple-envs/config/ -f examples/data-values-multiple-envs/envs/prod.yml
