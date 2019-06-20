### ytt Library: Overlay

Overlay package provides a way to combine two structures together with the help of annotations.

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

Functions:

- `apply(left, right1[, rightX...])`
```python
overlay.apply(left(), right())
overlay.apply(left(), one(), two())
```

- `map_key(name)` matcher matches array items based on the value of map key `name`; in the below example it will succesfully match array items that have a map with a key `name` of value `item2`
```yaml
#@overlay/match by=overlay.map_key("name")
- name: item2
```

- `index(i)` matcher matches array item at given index
```yaml
#@overlay/match by=overlay.index(0)
- item10
```

- `all` matcher matches all documents or array items
```yaml
#@overlay/match by=overlay.all
- item10
```

- `subset` matcher matches document or array item that has content that matches given map
```yaml
#@overlay/match by=overlay.subset({"metadata":{"name":"example-ingress1"}})
---
spec:
  enabled: true
```

Annotations on the "right-side" nodes:

- `@overlay/match [by=matcher,expects=Int|String|List|Function,missing_ok=Bool]`
  - valid for both map and array items
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
- `@overlay/merge` (*default operation*)
  - valid for both map and array items
  - merges node content recursively
- `@overlay/remove`
  - valid for both map and array items
  - removes "left-side" node ("right-side" node value is ignored)
- `@overlay/replace`
  - valid for both map and array items
  - replaces "left-side" node value with "right-side" node value
- `@overlay/insert [before=Bool,after=Bool]`
	- valid only for array items
  - inserts array item at the matched index, before matched index, or after
- `@overlay/append`
  - valid only for array items
  - appends array item to the end of "left-side" array

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

Overlay order is determined by:

1. left-to-right for file flags
  - e.g. in `-f overlay1.yml -f overlay2.yml`, `overlay1.yml` will be applied first
1. if file flag is set to a directory, files are alphanumerically sorted
  - e.g. in `aaa/z.yml xxx/c.yml d.yml`, will be applied in following order `aaa/z.yml d.yml xxx/c.yml`
1. top-to-bottom order for overlay YAML documents within a single file
