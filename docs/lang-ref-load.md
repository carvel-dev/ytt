### Load Statement

#### Terminology

- `module`: single file; can export variables, functions, or be templated => some type of result e.g. yaml structure, or string, or None)
- `package`: single directory; contains modules
- `library`: collection of packages

#### Usage

Load statement allows to load functions from other modules (such as ones from [builtin `ytt` library](lang-ref-ytt.md)).

- [load](https://github.com/google/starlark-go/blob/master/doc/spec.md#load-statements)
```python
load("@ytt:overlay", "overlay")                # load overlay module from builtin ytt library
load("@ytt:overlay", "overlay"=ov)             # load overlay symbol under a different alias
load("helpers.star", "func1", "func2")         # load func1, func2 from Starlark file
load("helpers.lib.yml", "func1", "func2")      # load func1, func2 from YAML file
load("helpers.lib.txt", "func1", "func2")      # load func1, func2 from text file
load("/dir/helpers.lib.yml", "func1")          # load func1 from file relative to root of library
load("sub-dir/helpers.lib.txt", "func1")       # load func1 from a sub-directory
load("@project:dir/helpers.lib.txt", "func1")  # load func1 from a project located under _ytt_lib
```

`load` arguments are as follows:

1. location which takes following shape `[@[library]:][package/]{0,n}module`, where,
    - `library` could be `ytt` or local path under `_ytt_lib` directory
      - examples: [`ytt`](lang-ref-ytt.md), `github.com/k14s/k8s-lib`, `common`
    - `package` could be a directory path
      - examples: `overlay`, `regexp`, `app/`, `/app/something`
    - `module` is a file name or predefined name (included in `ytt` library)
      - examples: `module.lib.yml`
1. one or more symbols to import with optional aliases
    - examples: `func1`, `func1="as_func1"`

Files can be loaded from current or child directories. As of ytt v0.24.0, `/` package prefix can be used to load files relative to the root of the current library.

Note that there is a distinction between using `load("@project:dir/helpers.lib.txt", "func1")` or `load("@project/dir:helpers.lib.txt", "func1")`, in that, `:` signifies what ytt considers a self-contained library (i.e. is `dir` simply a package in `project` or `project/dir` a standalone library). This allows files to load other files relative to the library root.

To load a set of functions from a single file, you can create a `struct` that contains references to the functions. For example:
`funcs.star`:
```
load("@ytt:struct", "struct")

def testfunc():
  return 123
end

def otherfunc():
  return 456
end

mod = struct.make(testfunc=testfunc, otherfunc=otherfunc)
```

`config.yml`:
```
#@ load("funcs.star", "mod")

result: #@ mod.testfunc()
other_result: #@ mod.otherfunc()
```

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
#### Files

To make files available to `load` statement they have to be given to ytt CLI via `--file` (`-f`) option. The argument of that option can be a path to either of:

- a **file**: in which case the file can be loaded by its name.
- a **directory**: in which case all the files found can be loaded by using paths relative to the directory. If the directory contains a `_ytt_lib` folder, then libraries in it can also be loaded.

For example, given following directory structure:

```
app1.yml
helpers.lib.yml
_ytt_lib/apps/apps.lib.yml
sub-dir/more-helpers.lib.yml
sub-dir/_ytt_lib/weird-lib/funcs.lib.yml
```

- `ytt -f .` will make it possible for `app1.yml` to load:
  - `helpers.lib.yml`
  - `@apps:apps.lib.yml`
  - `sub-dir/more-helpers.lib.yml`
- `ytt -f helpers.lib.yml -f sub-dir -f app1.yml` will make it possible for `app1.yml` to load:
  - `helpers.lib.yml`
  - `more-helpers.lib.yml` (not `sub-dir/more-helpers.lib.yml`)
  - `@weird-lib:funcs.lib.yml`

#### Examples

- [Load](https://get-ytt.io/#example:example-load)
- [Load ytt library](https://get-ytt.io/#example:example-load-ytt-library)
- [Load custom library](https://get-ytt.io/#example:example-load-custom-library)
