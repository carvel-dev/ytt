#!/bin/bash

set -e -x -u

LATEST_GIT_TAG=$(git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
VERSION="${1:-$LATEST_GIT_TAG}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X github.com/k14s/ytt/pkg/version.Version=$VERSION -buildid="

rm -f website/generated.go

go fmt ./cmd/... ./pkg/... ./test/...
go mod vendor
go mod tidy

# build without website assets
rm -f pkg/website/generated.go
go build -ldflags="$LDFLAGS" -trimpath -o ytt ./cmd/ytt/...
./ytt version

(
	# Use newly built binary to template all website assets
	# into a single Go file
	cd pkg/website
	./../../ytt \
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

# rebuild with website assets
rm -f ./ytt
go build -ldflags="$LDFLAGS" -trimpath -o ytt ./cmd/ytt/...
./ytt version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build -ldflags="$LDFLAGS" -trimpath -o ./tmp/ytt ./cmd/ytt/...
go build -ldflags="$LDFLAGS" -trimpath -o ./tmp/main ./cmd/ytt-lambda-website/...
(
	cd tmp
	chmod +x main ytt
	rm -f ytt-lambda-website.zip
	zip ytt-lambda-website.zip main ytt
)

# TODO ./hack/generate-docs.sh

echo SUCCESS
