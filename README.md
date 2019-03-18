# ytt (YAML Templating Tool)

`ytt` is a templating tool that understands YAML structure allowing you to focus on your data instead of how to properly escape it.

[![](docs/ytt-playground-screenshot.png)](https://get-ytt.io/#example:example-demo)

Features:

- templating works on YAML structure (instead of text)
  - eliminates variety of problems such as invalid YAML formatting, escaping, etc.
- syntactic sugar for single YAML node conditionals and for loops
  - makes it easier to read densely conditioned templated
- templates are themselves valid YAML files
  - makes them friendly to existing editors and YAML tools
- includes *sandboxed* "fully featured" _Pythonic_ programming language
  - compared to what's exposed in go/template for example

Try out [online playground](https://get-ytt.io) or download latest binaries from [Releases](https://github.com/k14s/ytt/releases) page (playground is included via `ytt playground` command).

## Docs

- [Language](docs/lang.md)
- [Security](docs/security.md)
- [ytt vs X: How ytt is different from other tools / frameworks](docs/ytt-vs-x.md)

## Try it

```
ytt template -R -f examples/playground/example-demo/
ytt template -R -f examples/playground/example-demo/ --output tmp/
```

See [examples/playground/](examples/playground/) for examples shown on [get-ytt.io](https://get-ytt.io).

## Development

```bash
./hack/build.sh
./hack/test-unit.sh
./hack/test-e2e.sh
./hack/test-all.sh
```
