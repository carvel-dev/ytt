#!/bin/bash

set -e -x

./hack/build.sh

mkdir -p ./tmp

# check stdin reading
cat examples/eirini/config.yml | ./ytt -f - -f examples/eirini/input.yml > ./tmp/config.yml
diff ./tmp/config.yml examples/eirini/config-result.yml

./ytt -f examples/eirini/config.yml -f examples/eirini/input.yml > ./tmp/config.yml
diff ./tmp/config.yml examples/eirini/config-result.yml

./ytt -f examples/eirini/config-alt1.yml -f examples/eirini/input.yml > ./tmp/config-alt1.yml
diff ./tmp/config-alt1.yml examples/eirini/config-result.yml

# check directory reading
./ytt -f examples/eirini/ --dangerous-emptied-output-directory=tmp/eirini
diff ./tmp/eirini/config-alt2.yml examples/eirini/config-result.yml

# check playground examples
for name in $(ls examples/playground/basics/); do
  if [ "$name" != "example-assert" ] && \
    [ "$name" != "example-load-custom-library-module" ] && \
    [ "$name" != "example-ytt-library-module" ]; then
      ./ytt -f examples/playground/basics/${name} > /dev/null
  fi
done

./ytt -f examples/overlay > /dev/null
./ytt -f examples/overlay-files > /dev/null
./ytt -f examples/overlay-regular-files --file-mark file.yml:type=yaml-plain > /dev/null

diff <(./ytt -f examples/k8s-add-global-label)          examples/k8s-add-global-label/expected.txt
diff <(./ytt -f examples/k8s-adjust-rbac-version)       examples/k8s-adjust-rbac-version/expected.txt
diff <(./ytt -f examples/k8s-docker-secret)             examples/k8s-docker-secret/expected.txt
diff <(./ytt -f examples/k8s-relative-rolling-update)   examples/k8s-relative-rolling-update/expected.txt
diff <(./ytt -f examples/k8s-config-map-files)          examples/k8s-config-map-files/expected.txt
diff <(./ytt -f examples/k8s-update-env-var)            examples/k8s-update-env-var/expected.txt
diff <(./ytt -f examples/k8s-overlay-all-containers)    examples/k8s-overlay-all-containers/expected.txt
diff <(./ytt -f examples/k8s-overlay-remove-resources)  examples/k8s-overlay-remove-resources/expected.txt
diff <(./ytt -f examples/k8s-overlay-in-config-map)     examples/k8s-overlay-in-config-map/expected.txt
diff <(./ytt -f examples/concourse-overlay)             examples/concourse-overlay/expected.txt
diff <(./ytt -f examples/overlay-not-matcher)           examples/overlay-not-matcher/expected.txt

# test json output
./ytt -f examples/k8s-adjust-rbac-version -o json > /dev/null

# test pipe stdin
diff <(cat examples/k8s-relative-rolling-update/config.yml | ./ytt -f-) examples/k8s-relative-rolling-update/expected.txt

# test pipe redirect (on Linux, pipe is symlinked)
diff <(./ytt -f pipe.yml=<(cat examples/k8s-relative-rolling-update/config.yml)) examples/k8s-relative-rolling-update/expected.txt

# test data values
diff <(./examples/data-values/run.sh) examples/data-values/expected.txt

# test data values required
diff <(./ytt -f examples/data-values-required/inline -v version=123) examples/data-values-required/expected.txt
diff <(./ytt -f examples/data-values-required/function -v version=123) examples/data-values-required/expected.txt
diff <(./ytt -f examples/data-values-required/bulk -v version=123) examples/data-values-required/expected.txt

# test data values multiple envs
diff <(./examples/data-values-multiple-envs/run.sh) examples/data-values-multiple-envs/expected.txt

echo E2E SUCCESS
