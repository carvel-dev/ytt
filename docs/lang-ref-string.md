### Strings

```yaml
name1: #@ name + "-deployment"
name2: #@ "{}-deployment".format("name")
```

Copied here for convenience from [Starlark specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#stringelem_ords).

- [string·elem_ords](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·elem_ords)
- [string·capitalize](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·capitalize) (`S.capitalize()`)
```python
"hello, world!".capitalize()  # "Hello, world!"`
```

- [string·codepoint_ords](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·codepoint_ords)
- [string·count](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·count) (`S.count(sub[, start[, end]])`)
```python
"hello, world!".count("o")         # 2
"hello, world!".count("o", 7, 12)  # 1  (in "world")
```

- [string·endswith](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·endswith) (`S.endswith(suffix[, start[, end]])`)
```python
"filename.star".endswith(".star")  # True
'foo.cc'.endswith(('.cc', '.h'))   # True
```

- [string·find](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·find) (`S.find(sub[, start[, end]])`)
```python
"bonbon".find("on")        # 1
"bonbon".find("on", 2)     # 4
"bonbon".find("on", 2, 5)  # -1
```

- [string·format](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·format) (`S.format(*args, **kwargs)`)
```python
"a{x}b{y}c{}".format(1, x=2, y=3)   # "a2b3c1"
"a{}b{}c".format(1, 2)              # "a1b2c"
"({1}, {0})".format("zero", "one")  # "(one, zero)"
```

- [string·index](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·index) (`S.index(sub[, start[, end]])`)
```python
"bonbon".index("on")        # 1
"bonbon".index("on", 2)     # 4
"bonbon".index("on", 2, 5)  # error: substring not found (in "nbo")
```

- [string·isalnum](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·isalnum)
```python
"base64".isalnum()    # True
"Catch-22".isalnum()  # False
```

- [string·isalpha](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·isalpha)
- [string·isdigit](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·isdigit)
- [string·islower](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·islower)
- [string·isspace](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·isspace)
- [string·istitle](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·istitle)
- [string·isupper](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·isupper)
- [string·join](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·join) (`S.join(iterable)`)
```python
", ".join(["one", "two", "three"])  # "one, two, three"
"a".join("ctmrn".codepoints())      # "catamaran"
```

- [string·lower](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·lower)
- [string·lstrip](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·lstrip)
```python
"  hello  ".lstrip()  # "  hello"
```

- [string·partition](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·partition)
- [string·replace](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·replace) (`S.replace(old, new[, count])`)
```python
"banana".replace("a", "o")     # "bonono"
"banana".replace("a", "o", 2)  # "bonona"
```

- [string·rfind](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·rfind) (`S.rfind(sub[, start[, end]])`)
```python
"bonbon".rfind("on")           # 4
"bonbon".rfind("on", None, 5)  # 1
"bonbon".rfind("on", 2, 5)     # -1
```

- [string·rindex](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·rindex) (`S.rindex(sub[, start[, end]])`)
- [string·rpartition](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·rpartition)
- [string·rsplit](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·rsplit)
- [string·rstrip](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·rstrip) (`S.rstrip()`)
```python
"  hello  ".rstrip()  # "hello  "
```

- [string·split](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·split) (`S.split([sep [, maxsplit]])`)
```python
"one two  three".split()         # ["one", "two", "three"]
"one two  three".split(" ")      # ["one", "two", "", "three"]
"one two  three".split(None, 1)  # ["one", "two  three"]
"banana".split("n")              # ["ba", "a", "a"]
"banana".split("n", 1)           # ["ba", "ana"]
```

- [string·elems](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·elems)
- [string·codepoints](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·codepoints)
- [string·splitlines](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·splitlines) (`S.splitlines([keepends])`)
```python
"one\n\ntwo".splitlines()      # ["one", "", "two"]
"one\n\ntwo".splitlines(True)  # ["one\n", "\n", "two"]
```

- [string·startswith](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·startswith) (`S.startswith(prefix[, start[, end]])`)
```python
"filename.star".startswith("filename")  # True`
```

- [string·strip](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·strip) (`S.strip()`)
```python
"  hello  ".strip()  # "hello"
```

- [string·title](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·title)
- [string·upper](https://github.com/google/starlark-go/blob/master/doc/spec.md#string·upper) (`S.upper()`)
```python
"Hello, World!".upper()  # "HELLO, WORLD!"
```
