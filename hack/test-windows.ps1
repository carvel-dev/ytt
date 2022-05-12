#! /usr/bin/pwsh

# makes builds reproducible
$env:CGO_ENABLED = 0
$repro_flags = "-ldflags=-trimpath"

echo Formatting
go fmt ./cmd/... ./pkg/...

echo 'Building Binary...'
$env:GOOS = "windows"
$env:GOARCH = 386
go build -o ytt.exe ./cmd/ytt/...

$env:PATH = "$env:PATH;$pwd"

echo 'Running Tests...'
ytt version

# We only care about exit code
ytt -f examples\load-paths
