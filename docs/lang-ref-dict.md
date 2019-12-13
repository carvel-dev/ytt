### Dictionaries

```yaml
#@ color = {"red": 123, "yellow": 100, "blue": "245"}
red: #@ color["red"]
```

Copied here for convenience from [Starlark specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#dictclear).

- [dict·clear](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·clear) (`D.clear()`)
```python
x = {"one": 1, "two": 2}
x.clear()  # None
print(x)   # {}
```

- [dict·get](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·get) (`D.get(key[, default])`)
```python
x = {"one": 1, "two": 2}
x.get("one")       # 1
x.get("three")     # None
x.get("three", 0)  # 0
```

- [dict·items](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·items) (`D.items()`)
```python
x = {"one": 1, "two": 2}
x.items()  # [("one", 1), ("two", 2)]
```

- [dict·keys](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·keys) (`D.keys()`)
```python
x = {"one": 1, "two": 2}
x.keys()  # ["one", "two"]
```

- [dict·pop](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·pop) (`D.pop(key[, default])`)
```python
x = {"one": 1, "two": 2}
x.pop("one")       # 1
x                  # {"two": 2}
x.pop("three", 0)  # 0
x.pop("four")      # error: missing key
```

- [dict·popitem](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·popitem) (`D.popitem()`)
```python
x = {"one": 1, "two": 2}
x.popitem()  # ("one", 1)
x.popitem()  # ("two", 2)
x.popitem()  # error: empty dict
```

- [dict·setdefault](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·setdefault) (`D.setdefault(key[, default])`)
```python
x = {"one": 1, "two": 2}
x.setdefault("one")       # 1
x.setdefault("three", 0)  # 0
x                         # {"one": 1, "two": 2, "three": 0}
x.setdefault("four")      # None
x                         # {"one": 1, "two": 2, "three": None}
```

- [dict·update](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·update) (`D.update([pairs][, name=value[, ...])`)
```python
x = {}
x.update([("a", 1), ("b", 2)], c=3)
x.update({"d": 4})
x.update(e=5)
x  # {"a": 1, "b": "2", "c": 3, "d": 4, "e": 5}
```

- [dict·values](https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·values) (`D.values()`)
```python
x = {"one": 1, "two": 2}
x.values()  # [1, 2]
```
