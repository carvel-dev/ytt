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

- `all` matcher matches all array items
```yaml
#@overlay/match by=overlay.all
- item10
```

Annotations on the "right-side" nodes:

- `@overlay/match [by=matcher,exists=Int|String|List|Function,missing_ok=Bool]`
  - valid for both map and array items
  - specifies how to find node on the "left-side"
  - Defaults:
    - for array items, there is no default matcher
    - for map items, default matching is key equality
    - one "left-side" node is expected to be found
  - Examples:
    - `#@overlay/match by="name"`: expects to find 1 "left-side" node that has a key name with "right-side" node value (for that key)
    - `#@overlay/match by=overlay.map_key("name")`: (same as above)
    - `#@overlay/match by=overlay.all,exists="0+"`: expects to find all "left-side" nodes
    - `#@overlay/match missing_ok=True`: expects to find 0 or 1 matching "left-side" node
    - `#@overlay/match exists=2`: expects to find 2 "left-side" nodes
    - `#@overlay/match exists="2+"`: expects to find 2 or more "left-side" nodes
    - `#@overlay/match exists=[0,1,4]`: expects to find 0, 1 or 4 "left-side" nodes
    - `#@overlay/match exists=lambda x: return x < 10`: expects to less than 10 "left-side" nodes
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
