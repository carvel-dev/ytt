#!/bin/bash

set -e -x -u

LATEST_GIT_TAG=$(git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
VERSION="${1:-$LATEST_GIT_TAG}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X carvel.dev/ytt/pkg/version.Version=$VERSION"

rm -f website/generated.go

go fmt $(go list ./... | grep -v yaml.v2)
go mod vendor
go mod tidy

# build without website assets
./hack/generate-website-assets.sh

# rebuild with website assets
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
