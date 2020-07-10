
* Install: Grab prebuilt binaries from the [Releases page](https://github.com/k14s/ytt/releases)
* Play: Jump right in by trying out the [online playground](https://get-ytt.io/#playground)
* For more information about annotations, data values, overlays and other features see [Docs](https://github.com/k14s/ytt/tree/develop/docs) page
* Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io/)

## Overview

`ytt` (pronounced spelled out) is a templating tool that understands YAML structure. It helps you easily configure complex software via reusable templates and user provided values. Ytt includes the following features:
- Structural templating: understands yaml structure so users can focus on their configuration instead of issues associated with text templating, such as YAML value quoting or manual template indentation
- Built-in programming language: includes the "fully featured" Python-like programming language Starklark which helps ease the burden of configuring complex software through a richer set of functionality.
- Reusable configuration: You can reuse the same configuration in different environments by applying environment-specific values.
- Custom validations: coupled with the fast and deterministic execution, allows you to take advantage of faster feedback loops when creating and testing templates
- Overlays: this advanced configuration helps users manage the customization required for complex software. For more, see [this example](https://get-ytt.io/#example:example-overlay-files) in the online playground.
- Sandboxing: provides a secure, deterministic environment for execution of templates

## Try it

To get started with `ytt` and to see examples, you use the online playground or download the binaries and run the playground locally.
- Try out the [online playground](https://get-ytt.io/#playground)
- Download the latest binaries from the [releases page](https://github.com/k14s/ytt/releases) and run the playground locally :

  `ytt website`

- See the examples used in the playground on the [examples](https://github.com/k14s/ytt/tree/develop/examples/playground) page
- Editor Extensions: [vscode sytnax highlighting](https://marketplace.visualstudio.com/items?itemName=ewrenn.vscode-ytt)
