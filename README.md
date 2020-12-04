
* Install: [See below](#installation), or grab prebuilt binaries from the [Releases page](https://github.com/vmware-tanzu/carvel-ytt/releases)
* Play: Jump right in by trying out the [online playground](https://get-ytt.io/#playground)
* For more information about annotations, data values, overlays and other features see [Docs](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/docs) page
* Slack: [#carvel in Kubernetes slack](https://slack.kubernetes.io/)

## Installation

### Via script (macOS or Linux)

```terminal 
$ wget -O- https://carvel.dev/install.sh | bash
# or with curl...
$ curl -L https://carvel.dev/install.sh | bash
```

### Via Homebrew (macOS or Linux)

Based on <a href="https://github.com/vmware-tanzu/homebrew-carvel">github.com/vmware-tanzu/homebrew-carvel</a>

```terminal
$ brew tap vmware-tanzu/carvel
$ brew install ytt kbld kapp imgpkg kwt vendir
```

### Specific version from a GitHub release
    
To download, click on one of the assets in a chosen [GitHub release](https://github.com/vmware-tanzu/carvel-ytt/releases), for example for 'ytt-darwin-amd64'.

```terminal
# Compare binary checksum against what's specified in the release notes
# (if checksums do not match, binary was not successfully downloaded)
$ shasum -a 256 ~/Downloads/ytt-darwin-amd64
08b25d21675fdc77d4281c9bb74b5b36710cc091f30552830604459512f5744c   /Users/pivotal/Downloads/ytt-darwin-amd64
# Move binary next to your other executables
$ mv ~/Downloads/ytt-darwin-amd64 /usr/local/bin/ytt
# Make binary executable
$ chmod +x /usr/local/bin/ytt
# Check its version
$ ytt version
```

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
- Download the latest binaries from the [releases page](https://github.com/vmware-tanzu/carvel-ytt/releases) and run the playground locally :

  `ytt website`

- See the examples used in the playground on the [examples](https://github.com/vmware-tanzu/carvel-ytt/tree/develop/examples/playground) page
- Editor Extensions: [vscode sytnax highlighting](https://marketplace.visualstudio.com/items?itemName=ewrenn.vscode-ytt)
