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
./ytt -f examples/eirini/ --output=tmp/eirini
diff ./tmp/eirini/config-alt2.yml examples/eirini/config-result.yml

# check playground examples
for name in $(ls examples/playground/); do
  case "$name" in
    example-load-template-library) ;;
    example-assert) ;;
    example-load-custom-library) ;;
    *)
      ./ytt -f examples/playground/${name} > /dev/null
      ;;
  esac
done

./ytt -f examples/overlay > /dev/null
./ytt -f examples/overlay-files > /dev/null
./ytt -f examples/overlay-regular-files --file-mark file.yml:type=yaml-plain > /dev/null

diff <(./ytt -f examples/k8s-adjust-rbac-version)     examples/k8s-adjust-rbac-version/expected.txt
diff <(./ytt -f examples/k8s-docker-secret)           examples/k8s-docker-secret/expected.txt
diff <(./ytt -f examples/k8s-relative-rolling-update) examples/k8s-relative-rolling-update/expected.txt
diff <(./ytt -f examples/concourse-overlay)           examples/concourse-overlay/expected.txt

echo E2E SUCCESS
