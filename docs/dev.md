# Development References

For those interested in extending and improving `ytt`, below is a quick reference on the structure of the source code:

- [pkg/yamlmeta](https://github.com/k14s/ytt/tree/master/pkg/yamlmeta) handles paraing of yaml and dealing with tracking metadata for yaml nodes
- [pkg/yamltemplate](https://github.com/k14s/ytt/tree/master/pkg/yamltemplate) deals with assembling template in starlark from yaml
- [pkg/template](https://github.com/k14s/ytt/tree/master/pkg/template) deals with evaluating the generic starlark template
  - more specifically [Starlark](https://github.com/google/starlark-go) compilation happens in: [pkg/template/compiled_template.go](https://github.com/k14s/ytt/blob/master/pkg/template/compiled_template.go)
  - ultimately it uses [Starlark go library](https://github.com/google/starlark-go) to run through the template code it builds up
- [pkg/workspace](https://github.com/k14s/ytt/tree/master/pkg/workspace) represents in memory representation of files given to ytt command
load directive goes through `templateloader` to load in what is necessary
- [pkg/cmd/template/cmd.go](https://github.com/k14s/ytt/blob/master/pkg/cmd/template/cmd.go#L95) is the top level call site
- [pkg/yttlibrary](https://github.com/k14s/ytt/tree/master/pkg/yttlibrary) is the bundled `@ytt` library but you can also make your own libraries as explained in the [k14s/k14s/k8s-lib](https://github.com/k14s/k8s-lib) example
