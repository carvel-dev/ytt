### ytt Library: Overlay module

Overlay module provides a way to combine two structures together with the help of annotations. You can use overlay functionality:

- [programmatically via `overlay.apply` function](#programmatic-access)
- [by specifying overlays as standalone YAML documents](#overlays-as-files)

When combining two structures together we refer to the structure is modified as "left-side" (or base) and to the structure that specifies modifications as "right-side" (or overlay). Modifications are described via overlay annotations on the right-side structures.

Annotations on the "right-side" nodes:

- `@overlay/match [by=matcher,expects=Int|String|List|Function,missing_ok=Bool]`
  - valid for documents, map and array items
  - specifies how to find node on the "left-side"
  - Defaults:
    - for array items, there is no default matcher
    - for map items, default matching is key equality
    - one "left-side" node is expected to be found
  - Examples:
    - `#@overlay/match by="name"`: expects to find 1 "left-side" node that has a key name with "right-side" node value (for that key)
    - `#@overlay/match by=overlay.map_key("name")`: (same as above)
    - `#@overlay/match by=overlay.all,expects="0+"`: expects to find all "left-side" nodes
    - `#@overlay/match missing_ok=True`: expects to find 0 or 1 matching "left-side" node
    - `#@overlay/match expects=2`: expects to find 2 "left-side" nodes
    - `#@overlay/match expects="2+"`: expects to find 2 or more "left-side" nodes
    - `#@overlay/match expects=[0,1,4]`: expects to find 0, 1 or 4 "left-side" nodes
    - `#@overlay/match expects=lambda x: return x < 10`: expects to less than 10 "left-side" nodes
- `@overlay/match-child-defaults [expects=...,missing_ok=Bool]`
  - valid for documents, map and array items
  - specifies `overlay/match` defaults for child nodes (does not apply to current node)
    - useful to avoid repeating `@overlay/match missing_ok=True` on each child node in maps
- `@overlay/merge` (*default operation*)
  - valid for both map and array items
  - merges node content recursively
- `@overlay/remove`
  - valid for both map and array items
  - removes "left-side" node ("right-side" node value is ignored)
- `@overlay/replace [via=Function]`
  - valid for documents, map and array items
  - replaces "left-side" node value with "right-side" node value
  - `via` (optional) keyword argument takes a function which will receive two arguments (left and right value) and expects to return single value
    - also works with non-scalar values (available in v0.26.0+)
  - Examples:
    - `#@overlay/replace`: (default)
    - `#@overlay/replace via=lambda a,b: "prefix-"+a`: prefix value with `prefix-` string
- `@overlay/insert [before=Bool,after=Bool]`
	- valid only for array items
  - inserts array item at the matched index, before matched index, or after
- `@overlay/append`
  - valid only for array items
  - appends array item to the end of "left-side" array
- `@overlay/assert [via=Function]` (available in v0.24.0+)
  - valid for documents, map and array items
  - tests equality of "left-side" node value with "right-side" node value
  - `via` (optional) keyword argument takes a function which will receive two arguments (left and right value) and expects to return NoneType, Bool, or Tuple(Bool,String)
    - if NoneType is returned comparison is considered a success (typically assert module is used within)
    - if False is returned comparison is considered a failure
    - if Tuple(False,String) is returned comparison is considered a failure
    - also works with non-scalar values (available in v0.26.0+)
  - Examples:
    - `#@overlay/assert`: (default)
    - `#@overlay/assert via=lambda a,b: a > 0 and a < 1000`: check that value is within certain numeric constraint
    - `#@overlay/assert via=lambda a,b: regexp.match("[a-z0-9]+", a)`: check that value is lowercase alphanumeric
    
#### Functions

Several functions are provided by overlay module that are useful for executing overlay operation, and matching various structures.

- `apply(left, right1[, rightX...])` to combine two or more structures (see [Programmatic access](#programmatic-access) below for details)

    ```python
    overlay.apply(left(), right())
    overlay.apply(left(), one(), two())
    ```
  
- `map_key(name)` matcher matches array items or maps based on the value of map key `name` (it requires that all item values have specified map key, use `overlay.subset(...)` if that requirement is cannot be met):
   
    this example will successfully match array items that have a map with a key `name` of value `item2`:
    ```yaml
    #@overlay/match by=overlay.map_key("name")
    - name: item2
    ```
    likewise, this example matches items in a map that have a key `name` of value `item2` (note: the key name `_` is arbitrary and ignored) (as of v0.26.0+):
    ```yaml
    #@overlay/match by=overlay.map_key("name")
    _:
      name: item2
    ```

- `index(i)` matcher matches array item at given index

    ```yaml
    #@overlay/match by=overlay.index(0)
    - item10
    ```

- `all` matcher matches all:
 
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

- `subset` matcher matches document, array item, or map item that has content that matches given map

    ```yaml
    #@overlay/match by=overlay.subset({"metadata":{"name":"example-ingress1"}})
    ---
    spec:
      enabled: true
    ```

- `and_op` matcher takes one or more matchers and returns true if all matchers return true (as of v0.26.0+)

    ```yaml
    #@ not_sa = overlay.not_op(overlay.subset({"kind": "ServiceAccount"}))
    #@ inside_ns = overlay.subset({"metadata": {"namespace": "some-ns"}})
    #@overlay/match by=overlay.and_op(not_sa, inside_ns),expects="1+"
    ---
    #! ...
    ```

- `or_op` matcher takes one or more matchers and returns true if any matchers return true (as of v0.26.0+)

    ```yaml
    #@overlay/match by=overlay.or_op(overlay.subset({"kind": "ConfigMap"}), overlay.subset({"kind": "Secret"}))
    ---
    #! ...
    ```

- `not_op` matcher takes another matcher and returns opposite result (as of v0.26.0+)

    ```yaml
    #@overlay/match by=overlay.not_op(overlay.subset({"metadata": {"namespace": "app"}}))
    ---
    #! ...
    ```

#### Programmatic access

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

#### Overlays as files

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

#### Overlay order

More explicitly specified in v0.13.0.

Overlay order is determined by:

1. left-to-right for file flags
    - e.g. in `-f overlay1.yml -f overlay2.yml`, `overlay1.yml` will be applied first
1. if file flag is set to a directory, files are alphanumerically sorted
    - e.g. in `aaa/z.yml xxx/c.yml d.yml`, will be applied in following order `aaa/z.yml d.yml xxx/c.yml`
1. top-to-bottom order for overlay YAML documents within a single file
