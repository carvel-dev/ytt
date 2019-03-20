#!/bin/bash

set -e -x

./hack/build.sh

mkdir -p ./tmp

# check stdin reading
cat examples/eirini/config.yml | ./ytt template -f - -f examples/eirini/input.yml > ./tmp/config.yml
diff ./tmp/config.yml examples/eirini/config-result.yml

./ytt template -f examples/eirini/config.yml -f examples/eirini/input.yml > ./tmp/config.yml
diff ./tmp/config.yml examples/eirini/config-result.yml

./ytt template -f examples/eirini/config-alt1.yml -f examples/eirini/input.yml > ./tmp/config-alt1.yml
diff ./tmp/config-alt1.yml examples/eirini/config-result.yml

# check directory reading
./ytt template -R -f examples/eirini/ --output=tmp/eirini
diff ./tmp/eirini/config-alt2.yml examples/eirini/config-result.yml

# check playground examples
for name in $(ls examples/playground/); do
  if [ "$name" != "example-assert" ] && [ "$name" != "example-load-custom-library" ]; then
    ./ytt template -R -f examples/playground/${name} > /dev/null
  fi
done

./ytt template -R -f examples/overlay > /dev/null
./ytt template -R -f examples/overlay-files > /dev/null
./ytt template -R -f examples/overlay-regular-files --file-mark file.yml:type=yaml-plain > /dev/null

echo E2E SUCCESS
