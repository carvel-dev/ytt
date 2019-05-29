#!/bin/bash

set -e -x -u

rm -f website/generated.go

go fmt ./cmd/... ./pkg/...

# build without website assets
go build -o ytt ./cmd/ytt/...
./ytt version

mkdir -p tmp

(
	# Use newly built binary to template all website assets
	# into a single Go file
	cd pkg/website
	./../../ytt \
		-f . \
		-f ../../examples/playground \
		-f ../../hack/build-values.yml \
		--file-mark 'alt-example**/*:type=data' \
		--file-mark 'example**/*:type=data' \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--output ../../tmp/
)
mv tmp/generated.go.txt pkg/website/generated.go

# rebuild with website assets
rm -f ./ytt
go build -o ytt ./cmd/ytt/...
./ytt version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build -o ./tmp/ytt ./cmd/ytt/...
go build -o ./tmp/main ./cmd/ytt-lambda-website/...
(
	cd tmp
	chmod +x main ytt
	rm -f ytt-lambda-website.zip
	zip ytt-lambda-website.zip main ytt
)

# TODO ./hack/generate-docs.sh

echo SUCCESS
