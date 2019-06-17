# ytt (YAML Templating Tool)

- Website: https://get-ytt.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)

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

Try out [online playground](https://get-ytt.io) or download latest binaries from [Releases](https://github.com/k14s/ytt/releases) page (playground is included via `ytt website` command).

## Docs

- [Docs](docs/README.md) with topics language, security, ytt vs other tools, etc.
- Blog posts
  - [ytt: The YAML Templating Tool that simplifies complex configuration management](https://developer.ibm.com/blogs/yaml-templating-tool-to-simplify-complex-configuration-management/)
  - [Introducing k14s (Kubernetes Tools): Simple and Composable Tools for Application Deployment](https://content.pivotal.io/blog/introducing-k14s-kubernetes-tools-simple-and-composable-tools-for-application-deployment)
- Talks
  - [Introducing the YAML Templating Tool (ytt)](https://www.youtube.com/watch?v=KbB5tI_g3bo) on IBM Developer podcast
  - [TGI Kubernetes 079: YTT and Kapp](https://www.youtube.com/watch?v=CSglwNTQiYg) with Joe Beda

## Install

Grab prebuilt binaries from the [Releases page](https://github.com/k14s/ytt/releases).

## Try it

```
ytt -f examples/playground/example-demo/
ytt -f examples/playground/example-demo/ --output tmp/
```

See [examples/playground/](examples/playground/) for examples shown on [get-ytt.io](https://get-ytt.io).

## Development

```bash
./hack/build.sh
./hack/test-unit.sh
./hack/test-e2e.sh
./hack/test-all.sh

# include goog analytics in 'ytt website' command for https://get-ytt.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-ytt-io.yml ./hack/build.sh
```
