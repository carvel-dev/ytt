## ytt vs x

### ytt vs Go text/template (and other text templating tools)

- [Go's text/template](https://golang.org/pkg/text/template/)
- [Jinnja](http://jinja.pocoo.org/)

Most generic templating tools do not understand content that they are templating and consider it just plain text. ytt operates on YAML structures, hence typical escaping and formatting problems common to text templating tools are eliminated. Additionally, ytt provides a very easy way to make structures reusable in a much more readable way that's possible with some text templating tools.

### ytt vs jsonnet

- [Jsonnet](https://jsonnet.org/)

ytt conceptually is very close to [jsonnet](https://jsonnet.org/). Both operate on data structures instead of text, hence are able to provide a better way to construct, compose and reuse structures. jsonnet introduces a custom language to help perform structure operations. ytt on the other hand, builds upon a Python-like language, which we think will be more familiar to the larger community.

We also believe that transitioning from plain YAML to templated YAML with `ytt` is very easy and natural.

### ytt vs Dhall

- [Dhall](https://dhall-lang.org/)

Dhall language is a configuration language that can output YAML, and JSON. One of its strong points is ability to provide scripting environment that is "hermetically sealed" and safe, even against malicious templates. `ytt` also embraces same goal (and builds upon the great work of Starlark community) by exposing small API in the template context. For example, there is no way to make network calls, read from file system, _or currently, even get time_.

### ytt vs Kustomize (and CF BOSH ops files)

- [Kustomize](https://kubernetes.io/blog/2018/05/29/introducing-kustomize-template-free-configuration-customization-for-kubernetes/)
- [CF BOSH's ops files](https://bosh.io/docs/cli-ops-files)

Configuration customization tools are unique in a sense that they don't allow templating but rather build upon "base" configuration. ytt offers its own take on configuration customization via the ['overlay' feature](https://github.com/get-ytt/ytt/blob/master/docs/lang-ref-ytt-overlay.md). Unlike other tools, overlay operations (remove, replace, merge) in `ytt` mimic structure of the base configuration. For example in Kustomize to remove a particular map key, one has to use JSON patch syntax which is quite different from the normal document structure. On the other hand, `ytt` uses its ability to annotate YAML structures, hence it can mark map key that should be deleted. All in all, we think that `ytt`'s approach is superior.

### ytt vs Orchestration Tools (Pulumi / HELM)

- [Pulumi](https://www.pulumi.com/)
- [HELM](https://helm.sh/)

Orchestration tools like Pulumi, and HELM, have combined configuration management and workflow management into the same tool. There are advantages and disadvantages to that. `ytt` is designed specifically to only focus on configuration management. Though, YAML output can be used with HELM, Pulumi, or other tools.

### ytt vs plain Ruby/Python/etc

Key advantages for `ytt`:

- provides an easy way to operate on structures (maps, lists, etc.). One can definitely use a regular language to do data manipulation. However, this is not what the language is optimized for, especially if data is heavily nested, typically leading to very verbose and less readable code.
- provides an _easy and safe_ way to execute templates without worrying that template code may be malicious. One can Dockerize execution of regular language templates but of course that brings in pretty heavy dependency.
- provides an _easy_ way to customize any part of configuration via [overlays](https://github.com/get-ytt/ytt/blob/master/docs/lang-ref-ytt-overlay.md). This is not possible to do with a regular language without parameterizing everything (a general anti-pattern) or bringing in an additional tool (e.g. BOSH ops files).
