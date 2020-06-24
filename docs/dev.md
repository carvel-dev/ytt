## Building

```bash
git clone https://github.com/k14s/ytt ./src/github.com/k14s/ytt
export GOPATH=$PWD
cd ./src/github.com/k14s/ytt

./hack/build.sh
./hack/test-unit.sh
./hack/test-e2e.sh
./hack/test-all.sh

# include goog analytics in 'ytt website' command for https://get-ytt.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-ytt-io.yml ./hack/build.sh
```

## Source Code Structure

For those interested in extending and improving `ytt`, below is a quick reference on the structure of the source code:

- [pkg/cmd/template/cmd.go](https://github.com/k14s/ytt/blob/develop/pkg/cmd/template/cmd.go#L95) is the top level call site for the template command
- [pkg/files](https://github.com/k14s/ytt/tree/develop/pkg/files) allows reading files from filesystem
- [pkg/workspace](https://github.com/k14s/ytt/tree/develop/pkg/workspace) keeps read files from the filesystem in memory for later access
  - `load(...)` directive goes through `TemplateLoader` to load files
- [pkg/yamlmeta](https://github.com/k14s/ytt/tree/develop/pkg/yamlmeta) parses YAML and tracks YAML node annotations (metas)
- [pkg/yamltemplate](https://github.com/k14s/ytt/tree/develop/pkg/yamltemplate) generates Starlark template based on yamlmeta package structures
- [pkg/texttemplate](https://github.com/k14s/ytt/tree/develop/pkg/texttemplate) parses text templates and generates Starlark template
- [pkg/template](https://github.com/k14s/ytt/tree/develop/pkg/template)
  - `InstructionSet` provides generic template instruction set for building Starlark templates
  - `CompiledTemplate` uses [Starlark Go library](https://github.com/google/starlark-go) to evaluate Starlark code
- [pkg/yttlibrary](https://github.com/k14s/ytt/tree/develop/pkg/yttlibrary) is bundled `@ytt` library
  - you can also make your own libraries as exemplified by [k14s/k8s-lib](https://github.com/k14s/k8s-lib)

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

`pkg/website` contains HTML/JS/CSS source for get-ytt.io. `./hack/build.sh` combines those assets into a single file `pkg/website/generated.go` (using ytt itself ;) ) before building final binary. `ytt website` command can serve website locally. `./hack/build.sh && ytt website` can be used for pretty quick local iteration.

Ultimately `get-ytt.io` runs on AWS Lambda, hence is wrapped with Lambda specific library. See `cmd/ytt-lambda-website/main.go` for details.

## Terminology

- library path: string that describes location of a library under _ytt_lib in format "@lib1@nested-lib2". used in:
  - `library.get("@...")`
  - `load("@...")`
- library alias: string that is used in place of a path when referencing an instances of a library. Assigned via the alias kwarg to library.get(...). For example, the instance returned from `library.get("/github.com/folder/my-lib", alias=my-lib)` will be addressable in the `@library/ref` annotations using just `my-lib`.
- library ref: string that describes  reference to a library under _ytt_lib. The ref can be either the library path or library alias. For example, "@lib1@~foo". used in:
  - `library/ref` annotation, e.g. `#@library/ref "@lib1@~foo"`
  - data value flag, e.g. `-v @~foo:key=value`
