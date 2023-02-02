## Building

```bash
git clone https://github.com/carvel-dev/ytt ./src/github.com/carvel-dev/ytt
export GOPATH=$PWD
cd ./src/github.com/carvel-dev/ytt

./hack/build.sh
./hack/test-unit.sh
./hack/test-e2e.sh
./hack/test-all.sh
```

## Source Code Changes
To keep source code documentation up to date, ytt uses [godoc](https://go.dev/blog/godoc). To document a type, variable, constant, function, or a package, write a regular comment directly preceding its declaration that begins with the name of the element it describes. See the ["@ytt:ip" module](https://github.com/carvel-dev/ytt/blob/develop/pkg/yttlibrary/ip.go) for examples. When contributing new source code via a PR, the [GitHub Action linter](https://github.com/carvel-dev/ytt/blob/develop/.github/workflows/golangci-lint.yml) will ensure that godocs are included in the changes.

To view the docs
  1. install godoc: `go get -v  golang.org/x/tools/cmd/godoc`
  1. Start the server: `godoc -http=:6060` and visit [`http://localhost:6060/pkg/github.com/carvel-dev/ytt/`](http://localhost:6060/pkg/github.com/carvel-dev/ytt/).

The go module was renamed from `github.com/k14s` to `github.com/carvel-dev/ytt` in February 2022. Carvel started out as a suite named "k14s" short for Kubernetes Tools.
  
## Source Code Structure

For those interested in extending and improving `ytt`, below is a quick reference on the structure of the source code:

- [pkg/cmd/template/cmd.go](https://github.com/carvel-dev/ytt/blob/develop/pkg/cmd/template/cmd.go#L95) is the top level call site for the template command
- [pkg/files](https://github.com/carvel-dev/ytt/tree/develop/pkg/files) allows reading files from filesystem
- [pkg/workspace](https://github.com/carvel-dev/ytt/tree/develop/pkg/workspace) keeps read files from the filesystem in memory for later access
  - `load(...)` directive goes through `TemplateLoader` to load files
- [pkg/yamlmeta](https://github.com/carvel-dev/ytt/tree/develop/pkg/yamlmeta) parses YAML and tracks YAML node annotations (metas)
- [pkg/yamltemplate](https://github.com/carvel-dev/ytt/tree/develop/pkg/yamltemplate) generates Starlark template based on yamlmeta package structures
- [pkg/texttemplate](https://github.com/carvel-dev/ytt/tree/develop/pkg/texttemplate) parses text templates and generates Starlark template
- [pkg/template](https://github.com/carvel-dev/ytt/tree/develop/pkg/template)
  - `InstructionSet` provides generic template instruction set for building Starlark templates
  - `CompiledTemplate` uses [Starlark Go library](https://github.com/google/starlark-go) to evaluate Starlark code
- [pkg/yttlibrary](https://github.com/carvel-dev/ytt/tree/develop/pkg/yttlibrary) is bundled `@ytt` library
  - you can also make your own libraries as exemplified by [carvel-dev/ytt-library-for-kubernetes](https://github.com/carvel-dev/ytt-library-for-kubernetes)

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

## Contributing to the Standard Library

`ytt` prepares each template evaluation with a set of Starlark modules (e.g. `@ytt:data`, `@ytt:json`, `@ytt:template`).
The source for this library is at `pkg/yttlibrary`.

When contributing to this library note:

- **Begin with the API**
  - before delving too deep into implementation, solidify the Starlark API (usually a conversation with a maintainer).
  - this is codified in a couple of acceptance tests (`pkg/yamltemplate/filetests/ytt-library/...`)
  - the goal is to provide a consistent, goal-aligned feature set.
- **Design near Python and Go**
  - Starlark is a dialect of Python: to maintain the value of that heritage, be inspired by Python
  - much of the functionality exposed can come from the Go ecosystem: to keep solutions simple, be inspired by Go.
  - that said, this library is typically operating in a specific context: a hermetic and deterministic calculation of (YAML-formatted) configuration. Bend to adapt to that context.
- **Offer string() rather than assume string encoding**
  - concrete types whose instances template authors will handle should implement `UnconvertableStarlarkValue` rather than assume to convert to a string. 
  - in the conversion hint, suggest they use `string()`.
  - this preserves the ability to choose _other_ encoding strategies, later, if desired.
- **Names**
  - overall: look for conventions present in existing modules (e.g. `is_..()` for predicates,  `string()` to encode to string, ...)
  - **module** — consider in combination with function names: be as concrete as possible, reveal intention, and avoid redundancy (e.g. `ip.parse_addr()` > `ip.parse_ip_addr()`) 
  - **functions** — typically verb and verb phrases work best.
    - **built-in** — use the format "${module}.[${type_name}.]${function_name}" (e.g. IPAddrValue.IsIPv4 => "ip.addr.is_ipv4") (this value is not yet used)
    - Go definition of the Starlark function should almost always be the snakeCase format of the same name.
  - **Type()** — use the format "@ytt:${module}.${type}" (e.g. IPAddrValue => "@ytt:ip.addr" )
- **Fail fast** — users are usually better served by an error than an attempt to "guess" or "fix"; as soon as something seems ary, error out.
- **Make Objects Immutable** — return modified copies rather than mutating the receiver: this tends to make template code that much easier to reason.
- **built-in functions must return a valid `starlark.Value`**
  - Note: `nil` is _not_ a valid `starlark.Value`; use `starlark.None` instead.
- **Include Automated Tests** — include the ability to verify that your contribution (still) works, now and in the future
  - at least demonstrate that the intended functionality works (i.e. cover the happy path case)
  - should catch when a bump to underlying dependencies (if any) would cause a breaking change to your module
  - such tests typically live in `pkg/yamltemplate/filetests/ytt-library`

### Prior Efforts

For your convenience, here's a list of PRs of prior contributions to the standard library:
- [module for handling Internet Protocol data #433](https://github.com/carvel-dev/ytt/pull/433)
- [Add url type in the url module #372](https://github.com/carvel-dev/ytt/pull/372)
