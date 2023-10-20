# Application Programming Interfaces of ytt

There are two strategies for integrating with ytt:

1. [as an executable](#as-an-executable) — invoke `ytt` out-of-process as an executable running as a separate process.
2. [as a Go module](#as-a-go-module) — construct, configure, execute, and inspect the results of a `ytt` command object, in-process.

**For most use-cases, the as-an-executable integration is more appropriate**:
- decouples versioning of your tooling from `ytt` itself making it easier to upgrade "in the field."
- maintains well-defined integration interfaces making development and troubleshooting far easier.
- provides lots of safety checks to ensure that inputs are well-formed and valid.
- allows your tooling to be implemented in any language/platform that supports shelling out a process.

Integrating with `ytt` as a Go module is more appropriate when: 
- your use-case is narrow enough that few (if any) of the `ytt`-isms are useful to the end-user.
- _your_ tooling must be distributed/integrated via a single binary (e.g. supplying a container image is undesirable)

Performance is typically _not_ a concern when deciding whether to go out-of-process or in-process: with reasonable computing resources and sources residing on-machine, `ytt` can execute over a reasonably complex library typically less than a second.


## As an Executable

By "integrate with `ytt` as an executable" we simply mean invoking the `ytt` binary, passing in the parameters required.

How one does this depends largely on the platform from which you're integrating.

Examples of how this can be done in Go:

- https://github.com/carvel-dev/kapp-controller/blob/develop/pkg/template/ytt.go
- https://github.com/carvel-dev/terraform-provider-carvel/blob/develop/pkg/ytt/ytt.go


## As a Go Module

`ytt` does have a programmatic means of being invoked. However, there are no formal APIs: we can commit to backwards compatible _behavior_, but interfaces can and will change. Integrators need to be prepared to migrate with version bumps.

### Overview
1. add `ytt` as a dependency:
   ```go.mod
   ...
   require  (
       ...
       carvel.dev/ytt v0.40.0
       ...
   )
   ```

2. create and populate an instance of the template command:
   ```go
   opts := template.NewOptions()
   ```

3. invoke the command to evaluate:
   ```go
   output := opts.RunWithFiles(inputs, ui)
   ```

### Examples
- [./internal-templating/main.go](internal-templating/main.go) — "Hello, world" with ytt.
- https://github.com/carvel-dev/kapp/blob/develop/pkg/kapp/yttresmod/overlay_contract_v1_mod.go — using `ytt` to overlay some existing YAML (here, a Kubernetes resource).

# Next Steps

We are always happy to support you in your use of `ytt`. You can:
- find us at [Kubernetes Slack: #carvel](https://kubernetes.slack.com/archives/CH8KCCKA5) (if your not yet in the workspace, get an [invite](http://slack.k8s.io/)).
- [pose a question](https://github.com/carvel-dev/ytt/discussions) or [report an issue](https://github.com/carvel-dev/ytt/issues/new/choose) in our GitHub repo.
