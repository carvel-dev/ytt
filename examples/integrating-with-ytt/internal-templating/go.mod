module example_internal_templating

go 1.17

replace github.com/vmware-tanzu/carvel-ytt => ../../../

require github.com/vmware-tanzu/carvel-ytt v0.40.1

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/hashicorp/go-version v1.4.0 // indirect
	github.com/k14s/starlark-go v0.0.0-20200720175618-3a5c849cc368 // indirect
)
