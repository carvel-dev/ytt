## Documentation

### What is ytt?
`ytt` is a templating tool that understands YAML structure. It helps to easily configure complex software via
 reusable templates and user provided values.

A comparison to similar tools can be found in the [ytt vs X doc](ytt-vs-x.md).
  
#### Templating
`ytt` performs templating by accepting "templates" (which are annotated YAML files) and "data values" 
(also annotated YAML files) and rendering those templates with the given values. 

For more:
* see it in action in the [load data files](https://get-ytt.io/#example:example-load-data-values) 
example on the playground
* check out the [data values reference docs](ytt-data-values.md)

#### Patching
`ytt` can also be used to patch YAML files. It does this by accepting overlay files (also in YAML) which describe
edits. This allows for configuring values beyond what was exposed as data values. 

For more:
* navigate to the [overlay files](https://get-ytt.io/#example:example-overlay-files)
example on the playground
* take a look at the [ytt Overlays overview](ytt-overlays.md)

### Next steps
- [Reference](#reference) — learn more of what can be done with `ytt`
- [Tips and Techniques](#tips-and-techniques) — ideas around common use cases
- [Additional Resources](#additional-resources) — read and watch more about `ytt` and [Carvel](https://carvel.dev)
- [Frequently Asked Questions](faq.md) — how to avoid common pitfalls
- [Getting Started Tutorial](https://get-ytt.io/#example:example-hello-world) — thorough stepwise introduction to `ytt` features

---

### Reference

#### Templating
- [Language](lang.md) — about the integration of `ytt`'s templating language, Starlark
  - [Strings](lang-ref-string.md)
  - [Lists](lang-ref-list.md)
  - [Dictionaries (a.k.a.: maps, hashes)](lang-ref-dict.md)
  - [Structs](lang-ref-structs.md)
  - [YAML Fragments](lang-ref-yaml-fragment.md)
  - [If conditional](lang-ref-if.md)
  - [For loop](lang-ref-for.md)
  - [Functions](lang-ref-def.md)
  - [Load statement](lang-ref-load.md)
  - [Full Starlark specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#contents)
- [@ytt library](lang-ref-ytt.md) — modules included with ytt
  - General
    - [assert module](lang-ref-ytt.md#assert)
    - [struct module](lang-ref-ytt.md#struct)
    - [data module](lang-ref-ytt.md#data)
    - [regexp module](lang-ref-ytt.md#regexp)
    - [url module](lang-ref-ytt.md#url)
    - [version module](lang-ref-ytt-version.md) — requiring a minimum ytt version
  - Serialization
    - [base64 module](lang-ref-ytt.md#base64)
    - [json module](lang-ref-ytt.md#json)
    - [yaml module](lang-ref-ytt.md#yaml)
  - Hashing
    - [md5 module](lang-ref-ytt.md#md5)
    - [sha256 module](lang-ref-ytt.md#sha256)
  - [overlay module](lang-ref-ytt-overlay.md) — patching on top of plain templating
  - [library module](lang-ref-ytt-library.md) — programmatically including template libraries
- [Annotation format](lang-ref-annotation.md) — how annotations look and work
  - [@data/values annotation](ytt-data-values.md) — injecting input into templates
- [Text templating](ytt-text-templating.md) — templating non-YAML files and strings

#### Command Line Options
- [Outputs](outputs.md) — manage where `ytt` outputs rendered templates
- [File marks](file-marks.md) — tweak how `ytt` handles individual files
- [Strict YAML subset](strict.md)

#### Misc
- [Known limitations](known-limitations.md)

---

### Tips and Techniques
- [@data/values vs overlays](data-values-vs-overlays.md) — when to use one over the other
- [Security](security.md) — sketch of a threat model of `ytt`
- [Injecting secrets](injecting-secrets.md) — thoughts around secrets management and distribution

---

### Additional Resources
#### Blog posts

- [ytt: The YAML Templating Tool that simplifies complex configuration management](https://developer.ibm.com/blogs/yaml-templating-tool-to-simplify-complex-configuration-management/)
- [Introducing k14s (Kubernetes Tools): Simple and Composable Tools for Application Deployment](https://content.pivotal.io/blog/introducing-k14s-kubernetes-tools-simple-and-composable-tools-for-application-deployment)

#### Talks

- [Introducing the YAML Templating Tool (ytt)](https://www.youtube.com/watch?v=KbB5tI_g3bo) on IBM Developer podcast
- [TGI Kubernetes 079: YTT and Kapp](https://www.youtube.com/watch?v=CSglwNTQiYg) with Joe Beda
- [Helm Summit 2019 - ytt: An Alternative to Text Templating of YAML Configuration](https://www.youtube.com/watch?v=7-PqgpkxC7E)
  - [presentation slides](https://github.com/k14s/meetups/blob/develop/ytt-2019-sep-helm-summit.pdf)

### Contributing

- [Contributing Guidelines](../CONTRIBUTING.md)
- [Development details](dev.md) — overview of the `ytt` codebase
