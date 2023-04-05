#!/bin/bash

set -e -x -u

# build website assets
rm -f pkg/website/generated.go
(
	# Use newly built binary to template all website assets
	# into a single Go file
	cd pkg/website
	go run ./../../cmd/ytt \
		-f . \
		-f ../../examples/playground/basics \
		-f ../../examples/playground/overlays \
		-f ../../examples/playground/getting-started \
		--file-mark 'alt-example**/*:type=data' \
		--file-mark 'example**/*:type=data' \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--dangerous-emptied-output-directory ../../tmp/
)
mv tmp/generated.go.txt pkg/website/generated.go
