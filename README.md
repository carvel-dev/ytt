# ytt (YAML Templating Tool)

- Website: https://get-ytt.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)
- [Docs](docs/README.md) with topics on language, security, ytt vs other tools, _blog posts and talks_ etc.
- Install: Grab prebuilt binaries from the [Releases page](https://github.com/k14s/ytt/releases)
- Editor Extensions: [vscode sytnax highlighting](https://marketplace.visualstudio.com/items?itemName=ewrenn.vscode-ytt)

`ytt` (pronounced spelled out) is a templating tool that understands YAML structure allowing you to focus on your data instead of how to properly escape it.

[![](docs/ytt-playground-screenshot.png)](https://get-ytt.io/#example:example-demo)

Features:

- templating works on YAML structure (instead of text)
  - which eliminates variety of problems such as invalid YAML formatting, escaping, etc.
- syntactic sugar for single YAML node conditionals and for loops
  - makes it easier to read densely conditioned templates
- templates are themselves valid YAML files
  - makes them friendly to existing editors and YAML tools
- includes *sandboxed* "fully featured" Python-like programming language
- allows configuration modularization via functions and libraries

## Try it

Try out [online playground](https://get-ytt.io) or download latest binaries from [Releases page](https://github.com/k14s/ytt/releases) and run it locally:

```
ytt -f examples/playground/example-demo/
ytt -f examples/playground/example-demo/ --output-directory tmp/
```

See [examples/playground/](examples/playground/) for examples shown on [get-ytt.io](https://get-ytt.io).

## Development

Consult [docs/dev.md](docs/dev.md) for build instructions, code structure details.
