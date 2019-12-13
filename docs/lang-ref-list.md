### Lists

```yaml
#@ nums = [123, 374, 490]
first: #@ nums[0]
```

Copied here for convenience from [Starlark specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#listappend).

- [list·append](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·append) (`L.append(x)`)
```python
x = []
x.append(1)  # None
x.append(2)  # None
x.append(3)  # None
x            # [1, 2, 3]
```

- [list·clear](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·clear) (`L.clear()`)
```python
x = [1, 2, 3]
x.clear()  # None
x          # []
```

- [list·extend](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·extend) (`L.extend(x)`)
```python
x = []
x.extend([1, 2, 3])  # None
x.extend(["foo"])    # None
x                    # [1, 2, 3, "foo"]
```

- [list·index](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·index) (`L.index(x[, start[, end]])`)
```python
x = list("banana".codepoints())
x.index("a")      # 1 (bAnana)
x.index("a", 2)   # 3 (banAna)
x.index("a", -2)  # 5 (bananA)
```

- [list·insert](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·insert) (`L.insert(i, x)`)
```python
x = ["b", "c", "e"]
x.insert(0, "a")   # None
x.insert(-1, "d")  # None
x                  # ["a", "b", "c", "d", "e"]
```

- [list·pop](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·pop) (`L.pop([index])`)
- [list·remove](https://github.com/google/starlark-go/blob/master/doc/spec.md#list·remove) (`L.remove(x)`)
```python
x = [1, 2, 3, 2]
x.remove(2)  # None (x == [1, 3, 2])
x.remove(2)  # None (x == [1, 3])
x.remove(2)  # error: element not found
```
