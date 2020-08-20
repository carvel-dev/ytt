# ytt Overlays Overview

Sometimes it makes more sense to patch some YAML rather than template it.

For example, when:
- the file should not be edited directly (e.g. from a third party);
- the edit will apply to most or all documents; or
- the specific variable is less commonly configured.

Given a sample target YAML file:

> `config.yml`
> ```yaml
> ---
> id: 1
> contents:
> - apple
> ---
> id: 2
> contents:
> - pineapple
> ```
... this overlay ...

> `add-content.yml`
> ```yaml
> #@ load("@ytt:overlay", "overlay")
> 
> #@overlay/match by=overlay.all, expects="1+"
> ---
> contents:
> #@overlay/append
> - pen
> ```

_read as..._
1. _"match all YAML documents, expecting to match _at least_ one;"_
2. _"within _each_ document, merge the key named `contents`;"_
3. _"append an array item with the content `"pen"`"_


... when processed by `ytt` ...

```console
$ ytt -f config.yml -f add-content.yml
```

... produces ...

> `config.yml` _(edited)_
> ```yaml
> id: 1
> contents:
> - apple
> - pen
> ---
> id: 2
> contents:
> - pineapple
> - pen
> ```

### Next Steps

- [Overlay example](https://get-ytt.io/#example:example-overlay-files) in the ytt Playground to try it out, yourself.
- [ytt Library: Overlay module](lang-ref-ytt-overlay.md) for reference of all overlay annotations and functions.
- [Data Values vs Overlays](data-values-vs-overlays.md) for when to use one mechanism over the other.
