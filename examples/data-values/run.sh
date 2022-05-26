#!/bin/bash

set -e

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-value nothing=null \
  --data-value string=str \
  --data-value bool=true \
  --data-value int=123 \
  --data-value float=123.123

echo '***'

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-value-yaml nothing=null \
  --data-value-yaml string=str \
  --data-value-yaml bool=true \
  --data-value-yaml int=123 \
  --data-value-yaml float=123.123

echo '***'

export STR_VAL_nothing=null
export YAML_VAL_string=str
export YAML_VAL_bool=true
export YAML_VAL_int=123
export YAML_VAL_float=123.123

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-values-env STR_VAL \
  --data-values-env-yaml YAML_VAL

echo '***'

export YAML_VAL_string=[1,2,4]

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-value-yaml nothing=[1,2,3] \
  --data-values-env-yaml YAML_VAL

echo '***'

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-values-file examples/data-values/values-file.yml

echo '***'

./ytt -f examples/data-values/config.yml -f examples/data-values/values.yml \
  --data-value-file string=examples/data-values/file-as-value.txt
