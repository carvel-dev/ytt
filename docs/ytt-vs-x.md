## ytt vs x

### ytt vs go/template (and other text templating tools)

Most generic templating tools do not understand content that they are templating and consider it just plain text. ytt operates on YAML structures, hence typical escaping and formatting problems common to text templating tools are eliminated. Additionally, ytt provides a very easy way to make structures reusable in a much more readable way that's possible with some text templating tools.

### ytt vs jsonnet

ytt conceptually is very close to [jsonnet](https://jsonnet.org/). Both operate on data structures instead of text, hence are able to provide a better way to construct, compose and reuse structures. jsonnet introduces a custom language to help perform structure operations. ytt on the other hand, builds upon a Python-like language, which we think will be more familiar to the larger community.

### ytt vs using Ruby/Python/etc

Key advantages for `ytt`:

- provides an easy way to operate on structures (maps, lists, etc.). One can definitely use a regular language to do data manipulation. However, this is not what the language is optimized for, especially if data is heavily nested, typically leading to very verbose and less readable code.
- provides an _easy and safe_ way to execute templates without worrying that template code may be malicious. One can Dockerize execution of regular language templates but of course that brings in pretty heavy dependency.
- provides an _easy_ way to customize any part of configuration via [overlays](https://github.com/get-ytt/ytt/blob/master/docs/lang-ref-ytt-overlay.md). This is not possible to do with a regular language without parameterizing everything (a general anti-pattern) or bringing in an additional tool (e.g. BOSH ops files).
