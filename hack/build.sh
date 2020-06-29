#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

rm -f website/generated.go

go fmt ./cmd/... ./pkg/...
go mod vendor
go mod tidy

# build without website assets
rm -f pkg/website/generated.go
go build -o ytt ./cmd/ytt/...
./ytt version

build_values_path="../../${BUILD_VALUES:-./hack/build-values-default.yml}"

(
	# Use newly built binary to template all website assets
	# into a single Go file
	cd pkg/website
	./../../ytt \
		-f . \
		-f ../../examples/playground/basics \
		-f ../../examples/playground/getting-started \
		-f $build_values_path \
		--file-mark 'alt-example**/*:type=data' \
		--file-mark 'example**/*:type=data' \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--dangerous-emptied-output-directory ../../tmp/
)
mv tmp/generated.go.txt pkg/website/generated.go

# rebuild with website assets
rm -f ./ytt
go build $repro_flags -o ytt ./cmd/ytt/...
./ytt version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build $repro_flags -o ./tmp/ytt ./cmd/ytt/...
go build $repro_flags -o ./tmp/main ./cmd/ytt-lambda-website/...
(
	cd tmp
	chmod +x main ytt
	rm -f ytt-lambda-website.zip
	zip ytt-lambda-website.zip main ytt
)

# TODO ./hack/generate-docs.sh

echo SUCCESS
