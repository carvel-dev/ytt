## Documentation

- [Language](lang.md) includes description of types
  - [Strings](lang-ref-string.md)
  - [Lists](lang-ref-list.md)
  - [Dictionaries (a.k.a.: maps, hashes)](lang-ref-dict.md)
  - [If conditional](lang-ref-if.md)
  - [For loop](lang-ref-for.md)
  - [Functions](lang-ref-def.md)
  - [Load statement](lang-ref-load.md)
  - [Starlark specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#contents) from google/starlark-go repo
  - [@ytt library](lang-ref-ytt.md) describes builtin modules in ytt
    - [overlay module](lang-ref-ytt-overlay.md) describes how to perform configuration patching on top of plain templating
    - [library module](lang-ref-ytt-library.md) describes how to programmatically template libraries
    - [version module](lang-ref-ytt-version.md) describes how to require minimum ytt version
  - [Annotation format](lang-ref-annotation.md) describes how annotations are built
    - [@data/values annotation](ytt-data-values.md) describes how to inject input into templates
  - [Text templating](ytt-text-templating.md) describes how to deal with complex text templating
- [@data/values vs overlays](data-values-vs-overlays.md)
- [Injecting secrets](injecting-secrets.md)
- [Outputs](outputs.md)
- [File marks](file-marks.md)
- [Security](security.md)
- [FAQ](faq.md)
- [Known limitations](known-limitations.md)
- [Strict YAML subset](strict.md)
- [ytt vs X: How ytt is different from other tools / frameworks](ytt-vs-x.md)

### Blog posts

- [ytt: The YAML Templating Tool that simplifies complex configuration management](https://developer.ibm.com/blogs/yaml-templating-tool-to-simplify-complex-configuration-management/)
- [Introducing k14s (Kubernetes Tools): Simple and Composable Tools for Application Deployment](https://content.pivotal.io/blog/introducing-k14s-kubernetes-tools-simple-and-composable-tools-for-application-deployment)

### Talks

- [Introducing the YAML Templating Tool (ytt)](https://www.youtube.com/watch?v=KbB5tI_g3bo) on IBM Developer podcast
- [TGI Kubernetes 079: YTT and Kapp](https://www.youtube.com/watch?v=CSglwNTQiYg) with Joe Beda
- [Helm Summit 2019 - ytt: An Alternative to Text Templating of YAML Configuration](https://www.youtube.com/watch?v=7-PqgpkxC7E)
  - [presentation slides](https://github.com/k14s/meetups/blob/develop/ytt-2019-sep-helm-summit.pdf)

### Misc

- [Development details](dev.md)
