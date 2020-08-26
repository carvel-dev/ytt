# ytt Library: Overlay module

## Contents

- [Overview](#overview)
- [Overlays as files](#overlays-as-files) — the primary way overlays are used: declaratively
- [Programmatic access](#programmatic-access) — applying overlays in a more precise way: programmatically
- [`@overlay` Annotations](#overlay-annotations) — how to declare overlays
- [Functions](#functions) — the contents of the `@ytt:overlay` module

---

## Overview

`ytt`'s Overlay feature provides a way to combine YAML structures together with the help of annotations.

There are two (2) structures involved in an overlay operation:
- the "left" — the YAML document(s) (and/or contained maps and arrays) being modified, and
- the "right" — the YAML document (and/or contained maps and arrays) that is the overlay, describing the modification.

Each modification is composed of:
- a matcher (via an [`@overlay/(match)`](#matching-annotations) annotation), identifying which node(s) on the "left" are the target(s) of the edit, and
- an action (via an [`@overlay/(action)`](#action-annotations) annotation), describing the edit.

Once written, an overlay can be applied in one of two ways:

- on all rendered templates, [declaratively, via YAML documents annotated with `@overlay/match`](#overlays-as-files); this is the most common approach.
- on selected documents, [programmatically, via `overlay.apply()`](#programmatic-access).

---
## Overlays as files

As `ytt` scans input files, it pulls aside any YAML Document that is annotated with `@overlay/match`, and considers it an overlay.

After YAML templates are rendered, the collection of identified overlays are applied. Each overlay executes, one-at-a-time over the entire set of the rendered YAML documents.

Order matters: modifications from earlier overlays are seen by later overlays. Overlays are applied in the order detailed in [Overlay order](#overlay-order), below.
 
In the example below, the last YAML document is an overlay (it has the `@overlay/match` annotation).
That overlay matcher's selects the first YAML document *only*: it's the only one that has a `metadata.name` of `example-ingress`.

```yaml
#@ load("@ytt:overlay", "overlay")

apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    ingress.kubernetes.io/rewrite-target: /
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: another-example-ingress
  annotations:
    ingress.kubernetes.io/rewrite-target: /

#@overlay/match by=overlay.subset({"metadata":{"name":"example-ingress"}})
---
metadata:
  annotations:
    #@overlay/remove
    ingress.kubernetes.io/rewrite-target:
```

yields:

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

See also: [Overlay files example](https://get-ytt.io/#example:example-overlay-files) in online playground.

__
### Overlay order

(as of v0.13.0)

Overlays are applied, in sequence, by:

1. left-to-right for file flags
    - e.g. in `-f overlay1.yml -f overlay2.yml`, `overlay1.yml` will be applied first
1. if file flag is set to a directory, files are alphanumerically sorted
    - e.g. in `aaa/z.yml xxx/c.yml d.yml`, will be applied in following order `aaa/z.yml d.yml xxx/c.yml`
1. top-to-bottom order for overlay YAML documents within a single file

__
### Next Steps

Familiarize yourself with the [overlay annotations](#overlay-annotations).

---
## Programmatic access

Overlays need not apply to the entire set of rendered YAML documents (as is the case with [the declarative approach](#overlays-as-files)).

Instead, the declared modifications can be captured in a function and applied to a specific set of documents via [`overlay.apply()`](#overlayapply) in Starlark code.

In this example we have `left()` function that returns target structure and `right()` that returns the overlay, specifying the modifications.

`overlay.apply(...)` will execute the execute the overlay and return a new structure.

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

yields:

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
__
### Next Steps

Familiarize yourself with the two kinds of overlay annotations:
- matchers (via an [`@overlay/(match)`](#matching-annotations) annotation), and
- actions (via an [`@overlay/(action)`](#action-annotations) annotation).

---
## `@overlay` Annotations

There are two groups of overlay annotations:

- [Matching Annotations](#matching-annotations)
- [Action Annotations](#action-annotations)


### Matching Annotations

These annotations are used to select which structure(s) will be modified:

- [@overlay/match](#overlaymatch)
- [@overlay/match-child-defaults](#overlaymatch-child-defaults)

__
#### @overlay/match

Specifies which nodes on the "left" to modify.

**Valid on:** Document, Map Item, Array Item.

```
@overlay/match [by=Function|String, expects=Int|String|List|Function, missing_ok=Bool, when=Int|String|List]
```
- **`by=`**`Function|String` — criteria for matching nodes on the "left"
   - `Function` — predicate of whether a given node is a match
       - provided matcher functions (supplied by the `@ytt:overlay` module):
         - [`overlay.all()`](#overlayall)
         - [`overlay.subset()`](#overlaysubset)
         - [`overlay.index()`](#overlayindex)
         - [`overlay.map_key()`](#overlaymap_key)
       - [Custom matcher function](#custom-overlay-matcher-functions) can also be used
   - `String` — short-hand for [`overlay.map_key()`](#overlaymap_key) with the same argument 
   - Defaults (depends on the type of the annotated node):
     - document or array item: none (i.e. `by` is required)
     - map item: key equality (i.e. [`overlay.map_key()`](#overlaymap_key))
- **`expects=`**`Int|String|List|Function` — (optional) expected number of nodes to be found in the "left." If not satisfied, raises an error.
   - `Int` — must match this number, exactly
   - `String` (e.g. `"1+"`) — must match _at least_ the number
   - `Function(found):Bool` — predicate of whether the expected number of matches were found
      - `found` (`Int`) — number of actual matches
   - `List[Int|String|Function]` — must match one of the given criteria
   - Default: `1` (`Int`) (i.e. expecting to match exactly one (1) node on the "left").
- **`missing_ok=`**`Bool` (optional) shorthand syntax for `expects="0+"`
- **`when=`**`Int|String|List` (optional) criteria for when the overlay should apply. If the criteria is met, the overlay applies; otherwise, nothing happens.
   - `Int` — must equal this number, exactly
   - `String` (e.g. `"1+"`) — must match _at least_ the number
   - `List[Int|String]` — must match one of the given criteria

**Notes:**
- `expects`, `missing_ok`, and `when` are mutually-exclusive parameters.
- take care when `expects` includes zero (0); matching none is indistinguishable from a mistakenly written match (e.g. a misspelling of a key name)

**Examples:**

- `#@overlay/match by="id"`: expects to find one (1) node on the "left" that has the key `id` and value matching the same-named item on the "right."
- `#@overlay/match by=`[`overlay.map_key("name")`](#overlaymap_key): (same as above)
- `#@overlay/match by=`[`overlay.all`](#overlayall)`,expects="0+"`: has no effective matching expectations
- `#@overlay/match missing_ok=True`: expects to find 0 or 1 matching nodes on the "left"
- `#@overlay/match expects=2`: expects to find exactly two (2) matching nodes on the "left"
- `#@overlay/match expects="2+"`: expects to find two (2) or more matching nodes on the "left"
- `#@overlay/match expects=[0,1,4]`: expects to find 0, 1 or 4 matching nodes on the "left"
- `#@overlay/match expects=lambda x: return x < 10`: expects 9 or fewer matching nodes on the "left"
- `#@overlay/match when=2`: applies changes only if two (2) nodes are found on the "left" and will not error otherwise
- `#@overlay/match when=2+`: applies changes only if two (2) or more nodes are found on the "left" and will not error otherwise
- `#@overlay/match when=[0,1,4]`: applies changes only if there were exactly 0, 1 or 4 nodes found on the "left" and will not error otherwise

**History:**
- v0.28.0+ — added `when` keyword argument.

__
##### Custom Overlay Matcher Functions

The matcher functions from `@ytt:overlay` cover many use-cases. From time-to-time, more precise matching is required.

A matcher function has the following signature:

`Function(index,left,right):Boolean`
   - `index` (`Int`) — the potential match's position in the list of all potential matches (zero-based)
   - `left` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the potential match/target node
   - `right` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the value of the annotated node in the overlay
   - returns `True` if `left` should be considered a match; `False` otherwise.

Most custom matchers can be written as a Lambda expression.

Lambda expressions start with the keyword `lambda`, followed by a parameter list, then a `:`, and a single expression that is the body of the function.

**Examples:**

_Example 1: Key presence or partial string match_

`left` contains `right`:
```python
lambda index, left, right: right in left
```
Returns `True` when `right` is "a member of" `left`

(see also: [Starlark Spec: Membership tests](https://github.com/google/starlark-go/blob/master/doc/spec.md#membership-tests) for more details)

__

_Example 2: Precise string matching_

`left` contains a key of the same name as the value of `right`:
```python
lambda index, left, right: left["metadata"]["name"].endswith("server")
```
See also:
- [Language: String](lang-ref-string.md) for more built-in functions on strings.
- [@ytt:regexp Library](lang-ref-ytt.md#regexp) for regular expression matching.



__
#### @overlay/match-child-defaults

Sets default values for `expects`, `missing_ok`, or `when` for the children of the annotated node.
Does not set these values for the annotated node, itself.

Commonly used to avoid repeating `@overlay/match missing_ok=True` on each child node in maps.

**Valid on:** Document, Map Item, Array Item.

```
@overlay/match-child-defaults [expects=Int|String|List|Function, missing_ok=Bool, when=Int|String|List]
```

_(see [@overlay/match](#overlaymatch) for parameter specifications.)_

**Examples:**

Without the `#@overlay/match-child-defaults`, _each_ of the four new annotation would have needed an `@overlay/match missing_ok=True` to apply successfully:
```yaml
---
metadata:
  annotations:
    ingress.kubernetes.io/rewrite-target: /

#@overlay/match by=overlay.all
---
metadata:
  #@overlay/match-child-defaults missing_ok=True
  annotations:
    nginx.ingress.kubernetes.io/limit-rps: 2000
    nginx.ingress.kubernetes.io/enable-access-log: "true"
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/client-body-buffer-size: 1M
```

---
### Action Annotations

The following annotations describe how to modify the matched "left" node.

They are:
  - [`@overlay/merge`](#overlaymerge) — (default) combine left and right nodes
  - [`@overlay/remove`](#overlayremove) — delete nodes from left
  - [`@overlay/replace`](#overlayreplace) — replace the left node
  - [`@overlay/insert`](#overlayinsert) — insert right node into left
  - [`@overlay/append`](#overlayappend) — add right node at end of collection on left
  - [`@overlay/assert`](#overlayassert) — declare an invariant on the left node

__
#### @overlay/merge

Merge the value of "right" node with the corresponding "left" node.

**Valid on:** Map Item, Array Item.

```
@overlay/merge
```
_(this annotation has no parameters.)_

**Note:** This is the default action; for each node in an overlay, either the action is explicitly specified or it is `merge`.


__
#### @overlay/remove

Deletes the matched "left" node.

**Valid on:** Document, Map Item, Array Item.

```
@overlay/remove
```
_(this annotation has no parameters.)_


__
#### @overlay/replace

Substitutes matched "left" node with the value of the "right" node (or by that of a provided function).

**Valid on:** Document, Map Item, Array Item.

```
@overlay/replace [via=Function]
```

- **`via=`**`Function(left, right): (any)` _(optional)_ determines the value to substitute in. If omitted, the value is `right`.
   - `left` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the matched node's value
   - `right` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the value of the annotated node

**History:**
- v0.26.0 — works with [`yamlfragment`](lang-ref-yaml-fragment.md) values.

**Examples:**

_Example 1: Use value from "right"_

Replaces the corresponding "left" with the value `"v1"`
```yaml
#@overlay/replace
apiVersion: v1
```
__

_Example 2: Edit string value_ 

```yaml
#@overlay/replace via=lambda left, right: "prefix-"+left
```

See also:
- `ytt` modules that export functions useful for manipulating values:
    - [base64 module](lang-ref-ytt.md#base64)
    - [json module](lang-ref-ytt.md#json)
    - [md5 module](lang-ref-ytt.md#md5)
    - [sha256 module](lang-ref-ytt.md#sha256)
    - [url module](lang-ref-ytt.md#url)
    - [yaml module](lang-ref-ytt.md#yaml)
- [Language: String](lang-ref-string.md) for built-in string functions.
- Other Starlark language features that manipulate values:
    - [string interpolation](https://github.com/google/starlark-go/blob/master/doc/spec.md#string-interpolation)
    - [conditional expressions](https://github.com/google/starlark-go/blob/master/doc/spec.md#conditional-expressions)
    - [index expressions](https://github.com/google/starlark-go/blob/master/doc/spec.md#index-expressions)
    - [slice expressions](https://github.com/google/starlark-go/blob/master/doc/spec.md#slice-expressions)

__
#### @overlay/insert

Inserts "right" node before/after the matched "left" node.

**Valid on:** Document, Array Item.

```
@overlay/insert [before=Bool, after=Bool]
```
- **`before=`**`Bool` whether to insert the "right" node immediately in front of the matched "left" node.
- **`after=`**`Bool` whether to insert the "right" node immediately following the matched "left" node.


__
#### @overlay/append

Inserts the "right" node after the last "left" node.

**Valid on:** Document, Array Item.

```
@overlay/append
```
_(this annotation has no parameters.)_

**Note:** This action implies an `@overlay/match` selecting the last node. Any other `@overlay/match` annotation is ignored. 

__
#### @overlay/assert

Checks assertion that value of "left" matched node equals that of the annotated "right" node (_or_ a provided predicate).

**Valid on:** Document, Map Item, Array Item.

```
@overlay/assert [via=Function]
```

- Default: checks that the value of the matched "left" node equals the value of the annotated "right" node.
- **`via`**`=Function(left, right):(Bool|Tuple(Bool|String)|None)` _(optional)_ predicate indicating whether "left" passes the check 
   - `left` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the matched node's value
   - `right` ([`yamlfragment`](lang-ref-yaml-fragment.md) or scalar) — the value of the annotated node
   - Return types:
     - `Bool` — if `False`, the assertion fails; otherwise, nothing happens.
     - `Tuple(Bool|String)` — if `False`, the assertion fails and specified string is appended to the resulting error message; otherwise nothing happens.
     - `None` — the assertion assumes to succeed. In these situations, the function makes use of the [`@ytt:assert`](lang-ref-ytt.md#assert) module to effect the assertion.
     
**History:**
- v0.26.0 — works with [`yamlfragment`](lang-ref-yaml-fragment.md) values.
- v0.24.0 — introduced

**Examples:**

_Example 1: Range check_

Fails the execution if `left` not between 0 and 1000, exclusively.

```yaml
#@overlay/assert via=lambda left, right: left > 0 and left < 1000
```

__

_Example 2: Well-formedness check_

Fails the execution if `left` contains anything other than lowercase letters or numbers.

```yaml
#@overlay/assert via=lambda left, right: regexp.match("[a-z0-9]+", left)
```

__

See also:
- `ytt` hashing functions from:
    - [md5 module](lang-ref-ytt.md#md5)
    - [sha256 module](lang-ref-ytt.md#sha256)
- Boolean expression operators and built-in functions, including:
    - [`in`](https://github.com/google/starlark-go/blob/master/doc/spec.md#membership-tests) (aka "membership test")
    - [`and` and `or`](https://github.com/google/starlark-go/blob/master/doc/spec.md#or-and-and)
    - [`any()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#any) or [`all()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#all)
    - [`hasattr()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#hasattr)
    - [`len()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#len)
    - [`type()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#type)
- [Language: String](lang-ref-string.md) functions


---
## Functions

The `@ytt:overlay` module provides several functions that support overlay use.

To use these functions, include the `@ytt:overlay` module:
 
```python
#@ load("@ytt:overlay", "overlay")
```

The functions exported by this module are:
- [overlay.apply()](#overlayapply)
- [overlay.map_key()](#overlaymap_key)
- [overlay.index()](#overlayindex)
- [overlay.all()](#overlayall)
- [overlay.subset()](#overlaysubset)
- [overlay.and_op()](#overlayand_op)
- [overlay.or_op()](#overlayor_op)
- [overlay.not_op()](#overlaynot_op)

__
### overlay.apply()

Executes the supplied overlays on top of the given structure.

```python
overlay.apply(left, right1[, rightN...])
```

- `left` ([`yamlfragment`](lang-ref-yaml-fragment.md)) — the target of the overlays
- `right1` ([`yamlfragment`](lang-ref-yaml-fragment.md) annotated with [`@overlay/(action)`](#action-annotations)) — the (first) overlay to apply on `left`.
- `rightN` ([`yamlfragment`](lang-ref-yaml-fragment.md) annotated with [`@overlay/(action)`](#action-annotations)) — the Nth overlay to apply on the result so far (which reflects the changes made by prior overlays)

**Notes:**
- For details on how to use `apply()`, see [Programmatic access](#programmatic-access).

**Examples:** 

```python
overlay.apply(left(), right())
overlay.apply(left(), one(), two())
```

See also: [Overlay example](https://get-ytt.io/#example:example-overlay) in the ytt Playground.

__
### overlay.map_key()

An [Overlay matcher function](#overlaymatch) that matches when the collection (i.e. Map or Array) in the "left" contains a map item with the key of `name` and value equal to the corresponding map item from the "right."
 
```python
overlay.map_key(name)
```
- `name` (`String`) — the key of the contained map item on which to match

**Note:** this matcher requires that _all_ items in the target collection have a map item with the key `name`; if this requirement cannot be guaranteed, consider using [`overlay.subset()`](#overlaysubset), instead.  

**Examples:**
   
_Example 1: Over an Array_

With "left" similar to:
```yaml
clients:
- id: 1
- id: 2
```
the following matches on the second array item:
```yaml
clients:
#@overlay/match by=overlay.map_key("id")
- id: 2
```
__

_Example 2: Over a Map_

(as of v0.26.0+)

With "left" similar to:
```yaml
clients:
  clientA:
    id: 1
  clientB:
    id: 2
```
the following matches on the second map item:
```yaml
---
clients:
  #@overlay/match by=overlay.map_key("id")
  _:
    id: 2
```
(note: the key name `_` is arbitrary and ignored).


__
### overlay.index()

An [Overlay matcher function](#overlaymatch) that matches the array item at the given index

```python
overlay.index(i)
```
- `i` (`Int`) — the ordinal of the item in the array on the "left" to match (zero-based index) 

**Example:**

```yaml
#@overlay/match by=overlay.index(0)
- item10
```

__
### overlay.all()

An [Overlay matcher function](#overlaymatch) that matches all contained nodes from the "left", unconditionally.

```python
overlay.all()
```
_(this function has no parameters.)_

**Examples:**

_Example 1: Documents_

Matches each and every document:
```yaml
#@overlay/match by=overlay.all
---
metadata:
 annotations: ...
```

_Example 2: Array Items_

Matches each and every item in the array contained in `items` on the "left":
```yaml
items:
#@overlay/match by=overlay.all
- item10
```

_Example 3: Map Items_

(as of v0.26.0+)

Matches each and every item in the map contained in `items` on the "left":
```yaml
items:
  #@overlay/match by=overlay.all
  _:
    name: item10
```
(note: the key name `_` is arbitrary and ignored) 

__
### overlay.subset()

An [Overlay matcher function](#overlaymatch) that matches when the "left" node's structure and value equals the given `target`.
 
```python
overlay.subset(target)
```
- `target` (`any`) — value that the "left" node must equal.

**Examples**

_Example 1: Scalar_

To match, scalar values must be equal.

```yaml
#@overlay/match by=overlay.subset(1)
#@overlay/match by=overlay.subset("Entire string must match")
#@overlay/match by=overlay.subset(True)
```
(if a partial match is required, consider writing a [Custom Overlay Matcher function](#custom-overlay-matcher-functions))

__

_Example 2: Dict (aka "map")_

To match, dictionary literals must match the structure and value of `left`.

```yaml
#@overlay/match by=overlay.subset({"kind": "Deployment"})
#@overlay/match by=overlay.subset("metadata":{"name": "istio-system"})
```
__

_Example 3: YAML Fragment_

To match, [`yamlfragment`](lang-ref-yaml-fragment.md)'s must match structure and value of `left`.

```yaml
#@ def resource(kind, name):
kind: #@ kind
metadata:
  name: #@ name
#@ end

#@overlay/match by=overlay.subset(resource("Deployment", "istio-system"))
```  

__
### overlay.and_op()

(as of v0.26.0+)

An [Overlay matcher function](#overlaymatch) that matches when all given matchers return `True`.

```python
overlay.and_op(matcher1, matcher2, ...)
```
- `matcher1`, `matcher2`, ... — one or more other [Overlay matcher function](#overlaymatch)s.

**Examples:**

```yaml
#@ not_sa = overlay.not_op(overlay.subset({"kind": "ServiceAccount"}))
#@ inside_ns = overlay.subset({"metadata": {"namespace": "some-ns"}})

#@overlay/match by=overlay.and_op(not_sa, inside_ns),expects="1+"
```

__
### overlay.or_op()

(as of v0.26.0+)

An [Overlay matcher function](#overlaymatch) that matches when at least one of the given matchers return `True`.

```python
overlay.or_op(matcher1, matcher2, ...)
```
- `matcher1`, `matcher2`, ... — one or more other [Overlay matcher function](#overlaymatch)s.

**Examples:**

```yaml
#@ config_maps = overlay.subset({"kind": "ConfigMap"})
#@ secrets = overlay.subset({"kind": "Secret"})

#@overlay/match by=overlay.or_op(config_maps, secrets)
```

__
### overlay.not_op()

(as of v0.26.0+)

An [Overlay matcher function](#overlaymatch) that matches when the given matcher does not.

```pythons
not_op(matcher)
```
- `matcher` — another [Overlay matcher function](#overlaymatch).

```yaml
#@overlay/match by=overlay.not_op(overlay.subset({"metadata": {"namespace": "app"}}))
```
