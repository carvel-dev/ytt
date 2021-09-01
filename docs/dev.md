## Building

```bash
git clone https://github.com/vmware-tanzu/carvel-ytt ./src/github.com/vmware-tanzu/carvel-ytt
export GOPATH=$PWD
cd ./src/github.com/vmware-tanzu/carvel-ytt

./hack/build.sh
./hack/test-unit.sh
./hack/test-e2e.sh
./hack/test-all.sh
```

## Source Code Changes
To keep source code documentation up to date, ytt uses [godoc](https://go.dev/blog/godoc). To document a type, variable, constant, function, or a package, write a regular comment directly preceding its declaration that begins with the name of the element it describes. See the ["@ytt:ip" module](https://github.com/vmware-tanzu/carvel-ytt/blob/develop/pkg/yttlibrary/ip.go) for examples. When contributing new source code via a PR, the [GitHub Action linter](https://github.com/vmware-tanzu/carvel-ytt/blob/develop/.github/workflows/golangci-lint.yml) will ensure that godocs are included in the changes.

To view the docs
  1. install godoc: `go get -v  golang.org/x/tools/cmd/godoc`
  1. Start the server: `godoc -http=:6060` and visit [`http://localhost:6060/pkg/github.com/k14s/ytt/`](http://localhost:6060/pkg/github.com/k14s/ytt/).
  
## Source Code Structure

For those interested in extending and improving `ytt`, below is a quick reference on the structure of the source code:

- [pkg/cmd/template/cmd.go](https://github.com/vmware-tanzu/carvel-ytt/blob/develop/pkg/cmd/template/cmd.go#L95) is the top level call site for the template command
- [pkg/files](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/files) allows reading files from filesystem
- [pkg/workspace](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/workspace) keeps read files from the filesystem in memory for later access
  - `load(...)` directive goes through `TemplateLoader` to load files
- [pkg/yamlmeta](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/yamlmeta) parses YAML and tracks YAML node annotations (metas)
- [pkg/yamltemplate](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/yamltemplate) generates Starlark template based on yamlmeta package structures
- [pkg/texttemplate](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/texttemplate) parses text templates and generates Starlark template
- [pkg/template](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/template)
  - `InstructionSet` provides generic template instruction set for building Starlark templates
  - `CompiledTemplate` uses [Starlark Go library](https://github.com/google/starlark-go) to evaluate Starlark code
- [pkg/yttlibrary](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/pkg/yttlibrary) is bundled `@ytt` library
  - you can also make your own libraries as exemplified by [vmware-tanzu/carvel-ytt-library-for-kubernetes](https://github.com/vmware-tanzu/carvel-ytt-library-for-kubernetes)

### Tests

- `./hack/test-unit.sh` executes various basic validation tests
  - Notable test locations:
    - `pkg/cmd/template/*_test.go`: functional testing of `ytt` command (as a combination of various high level features e.g. data values, overlays, templating)
    - `pkg/template/*_test.go`: mostly generic templating functionality
    - `pkg/yamlmeta/*_test.go`: mostly YAML parsing tests
    - `pkg/texttemplate/filetests/*`: functional testing of text templating (e.g. function definition, control flow)
    - `pkg/yamltemplate/filetests/*`: functional testing of YAML templating (e.g. function definition, control flow)
    - `pkg/yamltemplate/filetests/ytt-library/*`: functional testing of ytt provided modules (e.g. `base64`, `regexp`)
- `./hack/test-e2e.sh` executes various `examples/` directory content as end-to-end tests

## Website changes

`pkg/website` contains HTML/JS/CSS source for ytt playground. `./hack/build.sh` combines those assets into a single file `pkg/website/generated.go` (using ytt itself ;) ) before building final binary. `ytt website` command can serve website locally. `./hack/build.sh && ytt website` can be used for pretty quick local iteration.

Ultimately `get-ytt.io` runs on AWS Lambda, hence is wrapped with Lambda specific library. See `cmd/ytt-lambda-website/main.go` for details.

## Terminology

- library path: string that describes location of a library under _ytt_lib in format "@lib1@nested-lib2". used in:
  - `library.get("@...")`
  - `load("@...")`
- library alias: string that is used in place of a path when referencing an instances of a library. Assigned via the alias kwarg to library.get(...). For example, the instance returned from `library.get("/github.com/folder/my-lib", alias=my-lib)` will be addressable in the `@library/ref` annotations using just `my-lib`.
- library ref: string that describes  reference to a library under _ytt_lib. The ref can be either the library path or library alias. For example, "@lib1@~foo". used in:
  - `library/ref` annotation, e.g. `#@library/ref "@lib1@~foo"`
  - data value flag, e.g. `-v @~foo:key=value`
