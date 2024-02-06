module example_internal_templating

go 1.17

// ensure example works with this copy of ytt; remove before use
replace carvel.dev/ytt => ../../../

require carvel.dev/ytt v0.44.1

require (
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/k14s/starlark-go v0.0.0-20200720175618-3a5c849cc368 // indirect
)
