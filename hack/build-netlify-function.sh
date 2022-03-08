#!/bin/bash

set -e -x -u

LATEST_GIT_TAG=$(git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
VERSION="${1:-$LATEST_GIT_TAG}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X github.com/vmware-tanzu/carvel-ytt/pkg/version.Version=$VERSION -buildid="

rm -f website/generated.go

go fmt $(go list ./... | grep -v yaml.v2)
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

export GOOS=linux GOARCH=amd64
go build -ldflags="$LDFLAGS" -trimpath -o ../netlify/functions/template ./cmd/ytt-lambda-website/...
