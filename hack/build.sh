#!/bin/bash

set -e -x -u

rm -f playground/generated.go

go fmt ./cmd/... ./pkg/...

# build without playground assets
go build -o ytt ./cmd/ytt/...
./ytt version

mkdir -p tmp

(
	# Use newly built binary to template all playground assets
	# into a single Go file
	cd pkg/playground; 
	./../../ytt \
		-f . \
		-f ../../examples/playground \
		-f ../../hack/build-values.yml \
		--file-mark 'alt-example**/*:type=data' \
		--file-mark 'example**/*:type=data' \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--output ../../tmp/
)
mv tmp/generated.go.txt pkg/playground/generated.go

# rebuild with playground assets
rm -f ./ytt
go build -o ytt ./cmd/ytt/...
./ytt version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build -o ./tmp/ytt ./cmd/ytt/...
go build -o ./tmp/main ./cmd/lambda-playground/...
(
	cd tmp
	chmod +x main ytt
	rm -f lambda-playground.zip
	zip lambda-playground.zip main ytt
)

# TODO ./hack/generate-docs.sh

echo SUCCESS
