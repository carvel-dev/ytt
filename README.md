![logo](docs/CarvelLogo.png)

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7746/badge)](https://bestpractices.coreinfrastructure.org/projects/7746)

# ytt

* Play: Jump right in by trying out the [online playground](https://carvel.dev/ytt/#playground)
* Discover `ytt` in [video](https://youtu.be/WJw1MDFMVuk)
* For more information about annotations, data values, overlays and other features see [Docs](https://carvel.dev/ytt/docs/latest/) page
* Slack: [#carvel in Kubernetes slack](https://slack.kubernetes.io/)
* Install: Grab prebuilt binaries from the [Releases page](https://github.com/carvel-dev/ytt/releases) or [Homebrew Carvel tap](https://github.com/carvel-dev/homebrew)
* Backlog: [See what we're up to]([https://app.zenhub.com/workspaces/carvel-backlog-6013063a24147d0011410709/board?repos=173207060). (Note: we use ZenHub which requires GitHub authorization](https://github.com/orgs/carvel-dev/projects/1/views/1?filterQuery=repo%3A%22vmware-tanzu%2Fcarvel-ytt%22)).

## Overview

`ytt` (pronounced spelled out) is a templating tool that understands YAML structure. It helps you easily configure complex software via reusable templates and user provided values. Ytt includes the following features:
- Structural templating: understands yaml structure so users can focus on their configuration instead of issues associated with text templating, such as YAML value quoting or manual template indentation
- Built-in programming language: includes the "fully featured" Python-like programming language Starlark which helps ease the burden of configuring complex software through a richer set of functionality.
- Reusable configuration: You can reuse the same configuration in different environments by applying environment-specific values.
- Custom validations: coupled with the fast and deterministic execution, allows you to take advantage of faster feedback loops when creating and testing templates
- Overlays: this advanced configuration helps users manage the customization required for complex software. For more, see [this example](https://carvel.dev/ytt/#example:example-overlay-files) in the online playground.
- Sandboxing: provides a secure, deterministic environment for execution of templates

## Try it

To get started with `ytt` and to see examples, you use the online playground or download the binaries and run the playground locally.

- Try out the [online playground](https://carvel.dev/ytt/#playground)
- Download the latest binaries from the [releases page](https://github.com/carvel-dev/ytt/releases) and run the playground locally: `ytt website`
- See the examples used in the playground on the [examples](https://github.com/carvel-dev/ytt/tree/develop/examples/playground) page
- Editor Extensions: [vscode syntax highlighting](https://marketplace.visualstudio.com/items?itemName=ewrenn.vscode-ytt)

### Join the Community and Make Carvel Better
Carvel is better because of our contributors and maintainers. It is because of you that we can bring great software to the community. Please join us during our online community meetings. Details can be found on our [Carvel website](https://carvel.dev/community/).

You can chat with us on Kubernetes Slack in the #carvel channel and follow us on Twitter at @carvel_dev.

Check out which organizations are using and contributing to Carvel: [Adopter's list](https://github.com/carvel-dev/carvel/blob/master/ADOPTERS.md)

### Integrating with ytt

If you want to integrate `ytt` within your own tooling, review our [APIs](examples/integrating-with-ytt/apis.md).
