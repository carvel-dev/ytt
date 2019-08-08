## Text Templating

ytt supports text templating within YAML strings and `.txt` files.

Text templating is controlled via `(@` and `@)` directives. These directives can be combined with following markers:

- `=` to output result; result must be of type string
- `-` to trim space either to the left (if next to opening directive) or right (if next to closing directive)

Examples:

- `before (@ 123 @) middle (@= "tpl" @) after` produces `before  middle tpl after`
- `before    (@- 123 -@)   middle   (@-= "tpl" -@)   after` produces `beforemiddletplafter`

### Inside YAML strings

`+` operand or [`string·format(...)`](lang-ref-string.md) method provide a good way to build strings:

- `name: #@ name_prefix + "-secret"`
- `name: #@ name_prefix + "-" + str(1234)`
- `name: #@ "{}-secret-{}".format(name_prefix, name_suffix)`

However, occasionally it might be useful to use text templating directly in YAML strings. To do so YAML node must be annotated with `@yaml/text-templated-strings` (v0.17.0+). Annotation will apply to node and its child nodes.

Examples:

- basic use
```yaml
#@ val1 = "val1"
#@ val2 = "val2"

#@yaml/text-templated-strings
templated: "before (@= val1 @) middle (@= val2 @) after"

non_templated: "(@ something"
```

- nested nodes
```yaml
#@ val1 = "val1"
#@ val2 = "val2"

#@yaml/text-templated-strings
---
key:
  nested_key_(@= val1 @): "middle (@= val2 @) after"
```

See [Text template example](https://get-ytt.io/#example:example-text-template) in online playground.
