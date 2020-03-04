### YAMLFragments

YAMLFragment is a type of value that is defined directly in YAML (instead of plain Starlark). For example, function `val()` returns a value of type `yamlfragment`.

```yaml
#@ def vals():
key1: val1
key2:
  subkey: val2
#@ end
```

YAMLFragment may contain:

- YAML document set (array of YAML documents)
- YAML array
- YAML map
- null

Given various contents it wraps, YAMLFragment currently exposes limited ways of accessing its contents directly. Following accessors are available in v0.26.0+.

#### YAML Document Set

```yaml
#@ def docs():
---
doc1
---
doc2
---
doc3
#@ end
```

- access contents of a document at particular index
```python
docs()[1] # returns "doc2"
```

- loop over each document, setting `val` to its contents
```python
for val in docs():
  val # ...
end
```

#### YAML Array

```yaml
#@ def vals():
- val1
- val2
#@ end
```

- access contents of an array item at particular index
```python
vals()[1] # returns "val2"
```

- loop over each array item, setting `val` to its contents
```python
for val in vals():
  val # ...
end
```

#### YAML Map

```yaml
#@ def vals():
key1: val1
key2:
  subkey: val2
#@ end
```

- access contents of a map item with particular key
```python
vals()["key1"] # returns "val1"
```

- check if map contains particular key
```python
"key1" in vals() # returns True
"key6" in vals() # returns False
```

- loop over each map item, setting `val` to its contents
```python
for key in vals():
  val = vals()[key] # ...
end
```

- convert to a dictinoary
```python
dict(**vals()) # returns {"key1": "val1", "key2": yamlfragment({"subkey": "val2"})}
```
