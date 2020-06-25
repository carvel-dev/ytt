# ytt Library: Overlay module

Overlay module provides a way to combine two structures together with the help of annotations. You can use overlay functionality:

- [programmatically via `overlay.apply` function](#programmatic-access)
- [by specifying overlays as standalone YAML documents](#overlays-as-files)

When combining two structures together we refer to the structure is modified as "left-side" (or base) and to the structure that specifies modifications as "right-side" (or overlay). Modifications are described via overlay annotations on the right-side structures.

Each modification is made of one matching stage (as specified by `@overlay/match` annotation) and one action (e.g. `@overlay/merge`, `@overlay/replace`, etc.).

## Annotations on the "right-side" nodes

### @overlay/match

`@overlay/match [by=Function, expects=Int|String|List|Function, missing_ok=Bool, when=Int|String|List]`

Specifies how to find node on the "left-side". Valid for documents, map and array items.

- `by=Function` (optional) function used to match left-side node. See matchers section below for builtin matches. Can be used in conjunction with other keyword arguments. Default: (a) for array items or documents, no default (b) for map items, default matching is key equality.
- `expects=Int|String|List|Function` (optional) sets expected number of matched left-side nodes to be found; if more or less found, error is raised. Defaults to `1` (i.e. expecting to find exactly one left-side node).
- `missing_ok=Bool` (optional) shorthand syntax for `expects="0+"`
- `when=Int|String|List` (optional; available in v0.28.0+) checks if number of matched left-side nodes matches given criteria, and if it matches proceeds with overlay execution; otherwise does nothing. This keyword argument is similar to `expects` with an exception of raising an error making it useful to apply certain overlay operation conditionally.

`expects`, `missing_ok`, and `when` cannot be used simultaneously.

Examples:

- `#@overlay/match by="name"`: expects to find 1 "left-side" node that has a key name with "right-side" node value (for that key)
- `#@overlay/match by=overlay.map_key("name")`: (same as above)
- `#@overlay/match by=overlay.all,expects="0+"`: expects to find all "left-side" nodes
- `#@overlay/match missing_ok=True`: expects to find 0 or 1 matching "left-side" node
- `#@overlay/match expects=2`: expects to find 2 "left-side" nodes
- `#@overlay/match expects="2+"`: expects to find 2 or more "left-side" nodes
- `#@overlay/match expects=[0,1,4]`: expects to find 0, 1 or 4 "left-side" nodes
- `#@overlay/match expects=lambda x: return x < 10`: expects to less than 10 "left-side" nodes
- `#@overlay/match when=2`: applies changes when 2 "left-side" nodes are found but will not error otherwise
- `#@overlay/match when=2+`: applies changes when 2 or more "left-side" nodes are found but will not error otherwise
- `#@overlay/match when=[0,1,4]`: applies changes when 0, 1 or 4 "left-side" nodes are found but will not error otherwise

### @overlay/match-child-defaults

`@overlay/match-child-defaults [expects=..., missing_ok=Bool]`

Specifies `@overlay/match` defaults for child nodes (does not apply to current node). Mostly useful to avoid repeating `@overlay/match missing_ok=True` on each child node in maps. Valid for documents, map and array items.

### @overlay/merge

`@overlay/merge`

This is a default action. It merges node content recursively. Valid for both map and array items.

### @overlay/remove

`@overlay/remove`

Removes "left-side" node ("right-side" node value is ignored). Valid for documents, map and array items.

### @overlay/replace

`@overlay/replace [via=Function]`

Replaces "left-side" node value with "right-side" node value. Valid for documents, map and array items.

- `via=Function` (optional) takes a function which will receive two arguments (left-side and right-side value) and expects to return single new value. Works with non-scalar values as of v0.26.0+.

Examples:

- `#@overlay/replace`: uses right-side value (default)
- `#@overlay/replace via=lambda a,b: "prefix-"+a`: prefix left-side value with `prefix-` string

### @overlay/insert

`@overlay/insert [before=Bool, after=Bool]`

Inserts array item at the matched index, before matched index, or after. Valid only for documents and array items.

### @overlay/append

`@overlay/append`

Appends array item to the end of "left-side" array or document set. Valid only for documents and array items.

### @overlay/assert

`@overlay/assert [via=Function]` (available in v0.24.0+)

Tests equality of "left-side" node value with "right-side" node value. Valid for documents, map and array items.

- `via=Function` (optional) takes a function which will receive two arguments (left-side and right-side value) and expects to return NoneType, Bool, or Tuple(Bool,String)
  - if NoneType is returned comparison is considered a success (typically assert module is used within)
  - if False is returned comparison is considered a failure
  - if Tuple(False,String) is returned comparison is considered a failure
  - also works with non-scalar values (available in v0.26.0+)

Examples:

- `#@overlay/assert`: (default)
- `#@overlay/assert via=lambda a,b: a > 0 and a < 1000`: check that value is within certain numeric constraint
- `#@overlay/assert via=lambda a,b: regexp.match("[a-z0-9]+", a)`: check that value is lowercase alphanumeric

---
## Functions

Several functions are provided by overlay module that are useful for executing overlay operation, and matching various structures. Use `load("@ytt:overlay", "overlay")` to access these functions.

### overlay.apply

`overlay.apply(left, right1[, rightX...])` to combine two or more structures (see [Programmatic access](#programmatic-access) below for details)

```python
overlay.apply(left(), right())
overlay.apply(left(), one(), two())
```

### overlay.map_key

`overlay.map_key(name)` matcher matches array items or maps based on the value of map key `name` (it requires that all item values have specified map key, use `overlay.subset(...)` if that requirement is cannot be met)
   
This example will successfully match array items that have a map with a key `name` of value `item2`:

```yaml
#@overlay/match by=overlay.map_key("name")
- name: item2
```

Likewise, this example matches items in a map that have a key `name` of value `item2` (note: the key name `_` is arbitrary and ignored) (as of v0.26.0+):

```yaml
#@overlay/match by=overlay.map_key("name")
_:
  name: item2
```

### overlay.index

`overlay.index(i)` matcher matches array item at given index

```yaml
#@overlay/match by=overlay.index(0)
- item10
```

### overlay.all

`overlay.all` matcher matches all:
 
documents

```yaml
#@overlay/match by=overlay.all
---
metadata:
 annotations: ...
```

array items

```yaml
#@overlay/match by=overlay.all
- item10
```

or items in maps (note: the key name `_` is arbitrary and ignored) (as of v0.26.0+)

```yaml
#@overlay/match by=overlay.all
_:
  name: item10
```

### overlay.subset

`overlay.subset` matcher matches based on partial equality of given value. Value could be map or any scalar.

- when value is a map, left-side node must include all keys and values recursively to be a match
- when value is a scalar, left-side node must be of the same value to be a match

```yaml
#@overlay/match by=overlay.subset({"metadata":{"name":"example-ingress1"}})
---
spec:
  enabled: true
```

### overlay.and_op

`overlay.and_op(matcher1, matcher2, ...)` matcher takes one or more matchers and returns true if all matchers return true (as of v0.26.0+)

```yaml
#@ not_sa = overlay.not_op(overlay.subset({"kind": "ServiceAccount"}))
#@ inside_ns = overlay.subset({"metadata": {"namespace": "some-ns"}})
#@overlay/match by=overlay.and_op(not_sa, inside_ns),expects="1+"
---
#! ...
```

### overlay.or_op

`overlay.or_op(matcher1, matcher2, ...)` matcher takes one or more matchers and returns true if any matchers return true (as of v0.26.0+)

```yaml
#@overlay/match by=overlay.or_op(overlay.subset({"kind": "ConfigMap"}), overlay.subset({"kind": "Secret"}))
---
#! ...
```

### overlay.not_op

`not_op(matcher)` matcher takes another matcher and returns opposite result (as of v0.26.0+)

```yaml
#@overlay/match by=overlay.not_op(overlay.subset({"metadata": {"namespace": "app"}}))
---
#! ...
```

---
## Programmatic access

In this example we have `left()` function that returns left-side structure and `right()` that returns right-side structure that specifies modifications. `overlay.apply(...)` will execute modifications and return a new structure.

```yaml
#@ load("@ytt:overlay", "overlay")

#@ def left():
key1: val1
key2:
  key3:
    key4: val4
  key5:
  - name: item1
    key6: val6
  - name: item2
    key7: val7
#@ end

#@ def right():
#@overlay/remove
key1: val1
key2:
  key3:
    key4: val4
  key5:
  #@overlay/match by="name"
  - name: item2
    #@overlay/match missing_ok=True
    key8: new-val8
#@ end

result: #@ overlay.apply(left(), right())
```

with the result of

```yaml
result:
  key2:
    key3:
      key4: val4
    key5:
    - name: item1
      key6: val6
    - name: item2
      key7: val7
      key8: new-val8
```

---
## Overlays as files

ytt CLI treats YAML documents with `overlay/match` annotation as overlays. Such overlays could be specified in any file and are matched against all resulting YAML documents from all other files in the post-templating stage.

In the example below, last YAML document is considered to be an overlay because it has `overlay/match` annotation. It will match *only* first YAML document, which has `example-ingress` as its `metadata.name`.

```yaml
#@ load("@ytt:overlay", "overlay")
#@ some_path = "/"

apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    ingress.kubernetes.io/rewrite-target: #@ some_path
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: another-example-ingress
  annotations:
    ingress.kubernetes.io/rewrite-target: #@ some_path

#@overlay/match by=overlay.subset({"metadata":{"name":"example-ingress"}})
---
metadata:
  annotations:
    #@overlay/remove
    ingress.kubernetes.io/rewrite-target:
```

Result:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example-ingress
  annotations: {}
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: another-example-ingress
  annotations:
    ingress.kubernetes.io/rewrite-target: /
```

See [Overlay files example](https://get-ytt.io/#example:example-overlay-files) in online playground.

### Overlay order

Specified as follows in v0.13.0+.

Overlay order is determined by:

1. left-to-right for file flags
    - e.g. in `-f overlay1.yml -f overlay2.yml`, `overlay1.yml` will be applied first
1. if file flag is set to a directory, files are alphanumerically sorted
    - e.g. in `aaa/z.yml xxx/c.yml d.yml`, will be applied in following order `aaa/z.yml d.yml xxx/c.yml`
1. top-to-bottom order for overlay YAML documents within a single file
