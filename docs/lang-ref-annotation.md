### Annotation

#### Format

```
@ann1-name [ann-arg1, ann-arg2, ..., keyword-ann-arg1=val1]
```

- content between brackets is optional.

- annotation names are typically namespaced, for example, `overlay/merge` is an annotation within an `overlay` namespace. Annotation namespaces are there for general organization, they are not associated with loaded packages (from `load` keyword).

- annotation arguments (positional and keyword) is just plain code

#### Shared templating annotations

- `@template/code [code]` or just `@ [code]` (on its own line) represents plain code line

```yaml
#@ a = calculate(100)
key: value
```

- `@template/value [code]` or just `@ [code]` (at the end of line) represents a value associated structure

```yaml
key: #@ value
array:
- #@ value
```

- `@data/values` (no args) specifies values accessible via `data.values` from `@ytt:data` package (see [ytt @data/values](ytt-data-values.md) for more details.)

#### Text templating annotations

- `@text/trim-left` trims space to the left of code node
- `@text/trim-right` trims space to the right of code node

#### YAML templating annotations

- `@yaml/map-key-override` (no args)
  - necessary to indicate that map key is being intentionally overriden

- `@yaml/text-templated-strings` (no args)
  - necessary to indicate that node contents (including map key and map value) should be text templated (ie `(@` indicates start of text templating) (see [text templating](ytt-text-templating.md) for more details.)

#### Overlay annotations

See [@ytt:overlay Library](lang-ref-ytt-overlay.md) for list of annotations.
