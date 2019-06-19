# Source Code Structure

For those interested in extending and improving `ytt`, below is a quick reference on the structure of the source code:

- [pkg/cmd/template/cmd.go](https://github.com/k14s/ytt/blob/master/pkg/cmd/template/cmd.go#L95) is the top level call site for the template command
- [pkg/files](https://github.com/k14s/ytt/tree/master/pkg/files) allows reading files from filesystem
- [pkg/workspace](https://github.com/k14s/ytt/tree/master/pkg/workspace) keeps read files from the filesystem in memory for later access
  - `load(...)` directive goes through `TemplateLoader` to load files
- [pkg/yamlmeta](https://github.com/k14s/ytt/tree/master/pkg/yamlmeta) parses YAML and tracks YAML node annotations (metas)
- [pkg/yamltemplate](https://github.com/k14s/ytt/tree/master/pkg/yamltemplate) generates Starlark template based on yamlmeta package structures
- [pkg/yamltemplate](https://github.com/k14s/ytt/tree/master/pkg/texttemplate) parses text templates and generates Starlark template
- [pkg/template](https://github.com/k14s/ytt/tree/master/pkg/template)
  - `InstructionSet` provides generic template instruction set for building Starlark templates
  - `CompiledTemplate` uses [Starlark Go library](https://github.com/google/starlark-go) to evaluate Starlark code
- [pkg/yttlibrary](https://github.com/k14s/ytt/tree/master/pkg/yttlibrary) is bundled `@ytt` library
  - you can also make your own libraries as exemplified by [k14s/k8s-lib](https://github.com/k14s/k8s-lib)
