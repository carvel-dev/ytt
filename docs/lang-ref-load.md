### Load Statement

Load statement allows to load functions from other modules (such as ones from (builtin `ytt` library](lang-ref-ytt.md)).

- [load](https://github.com/google/starlark-go/blob/master/doc/spec.md#load-statements)
```python
load("@ytt:overlay", "overlay")                # load overlay module from builtin ytt library
load("@ytt:overlay", "overlay"=ov)             # load overlay symbol under a different alias
load("helpers.star", "func1", "func2")         # load func1, func2 from Starlark file
load("helpers.lib.yml", "func1", "func2")      # load func1, func2 from YAML file
load("helpers.lib.txt", "func1", "func2")      # load func1, func2 from text file
load("sub-dir/helpers.lib.txt", "func1")       # load func1 from a sub-directory
load("@project:dir/helpers.lib.txt", "func1")  # load func1 from a project located under _ytt_lib
```

`load` arguments are as follows:

1. location which takes following shape `[@[library]:][package/]{0,n}module`, where,
  - `library` could be `ytt` or local path under `_ytt_lib` directory
    - examples: [`ytt`](lang-ref-ytt.md), `github.com/k14s/k8s-lib`, `common`
  - `package` could be a directory path
    - examples: `overlay`, `regexp`, `app/`
  - `module` is a file name or predefined name (included in `ytt` library)
    - examples: `module.lib.yml`
1. one or more symbols to import with optional aliases
  - examples: `func1`, `func1="as_func1"`

#### Files

To make files available to `load` statement they have to be given to ytt CLI via `--file` (-f) flag.

For example given following directory structure:

```
app1.yml
helpers.lib.yml
_ytt_lib/apps/apps.lib.yml
```

- `ytt -f .` will make it possible for `app1.yml` to load `helpers.lib.yml` and `@apps:apps.lib.yml`
- equivalently, `ytt -f app1.yml -f helpers.lib.yml -f _ytt_lib/apps/apps.lib.yml` will also make above possible

#### _ytt_lib directory

`_ytt_lib` directory allows to keep private dependencies from consumers of libraries.

For example given following directory structure:

```
app1.yml
_ytt_lib/big-corp/sre.lib.yml
_ytt_lib/big-corp/_ytt_lib/big-corp/common/deployments.lib.yml
_ytt_lib/big-corp/_ytt_lib/big-corp/common/services.lib.yml
```

- `app1.yml` _can_ load `big-corp/sre.lib.yml` via `@big-corp:sre.lib.yml`
- `app1.yml` _cannot_ load `big-corp/_ytt_lib/big-corp/common/services.lib.yml` as it is a private dependency of anything inside `_ytt_lib/big-corp/` directory (e.g. `sre.lib.yml`)

hence making it possible for `big-corp/sre.lib.yml` module to keep its `big-corp/common` library dependency private.

#### Examples

- [Load](https://get-ytt.io/#example:example-load)
- [Load ytt library](https://get-ytt.io/#example:example-load-ytt-library)
- [Load custom library](https://get-ytt.io/#example:example-load-custom-library)
