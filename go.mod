module github.com/k14s/ytt

go 1.13

require (
	github.com/aws/aws-lambda-go v1.8.2
	github.com/cppforlife/cobrautil v0.0.0-20180924214100-a39a1714c920
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	go.starlark.net v0.0.0-20190219202100-4eb76950c5f0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200207164905-fd8842955e4e // ytt branch
