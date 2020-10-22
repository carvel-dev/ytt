# ytt Library

Following modules are included with `ytt`.

## General modules

### struct

See [@ytt:struct module docs](lang-ref-ytt-struct.md).

### assert

```python
load("@ytt:assert", "assert")

assert.fail("expected value foo, but was {}".format(value)) # stops execution
x = data.values.env.mysql_password or assert.fail("missing env.mysql_password")
```

### data

See [ytt @data/values](ytt-data-values.md) for more details

```python
load("@ytt:data", "data")

data.values # struct that has input values

# relative to current package
data.list()                # ["template.yml", "data/data.txt"]
data.read("data/data.txt") # "data-txt contents"

# relative to library root (available in v0.27.1+)
data.list("/")              # list files 
data.list("/data/data.txt") # read file
```

### regexp

```python
load("@ytt:regexp", "regexp")

regexp.match("[a-z]+[0-9]+", "__hello123__") # True

regexp.replace("[a-z]+[0-9]+", "__hello123__", "foo")                 # __foo__
regexp.replace("(?i)[a-z]+[0-9]+", "__hello123__HI456__", "bye")      # __bye__bye__
regexp.replace("([a-z]+)[0-9]+", "__hello123__bye123__", "$1")        # __hello__bye__
regexp.replace("[a-z]+[0-9]+", "__hello123__", lambda s: str(len(s))) # __8__
```

See the [RE2 docs](https://github.com/google/re2/wiki/Syntax) for more on the regex syntax supported.

Note that you can pass either a string or a lambda function as the third parameter. When given a string, `$` symbols are expanded, so that `$1` expands to the first submatch. When given a lambda function, the match is directly replaced by the result of the function.

### url

```python
load("@ytt:url", "url")

url.path_segment_encode("part part")   # "part%20part"
url.path_segment_decode("part%20part") # "part part"

url.query_param_value_encode("part part") # "part+part"
url.query_param_value_decode("part+part") # "part part"

url.query_params_encode({"x":["1"],"y":["2","3"],"z":[""]}) # "x=1&y=2&y=3&z="
url.query_params_decode("x=1&y=2&y=3;z")                    # {"x":["1"],"y":["2","3"],"z":[""]}
```

### version

`load("@ytt:version", "version")` (see [version module doc](lang-ref-ytt-version.md))

---
## Serialization modules

### base64

```python
load("@ytt:base64", "base64")

base64.encode("regular")      # "cmVndWxhcg=="
base64.decode("cmVndWxhcg==") # "regular"
```

### json

```python
load("@ytt:json", "json")

json.encode({"a": [1,2,3,{"c":456}], "b": "str"})
json.decode('{"a":[1,2,3,{"c":456}],"b":"str"}')
```

### yaml

```python
load("@ytt:yaml", "yaml")

yaml.encode({"a": [1,2,3,{"c":456}], "b": "str"})
yaml.decode('{"a":[1,2,3,{"c":456}],"b":"str"}')
```

---
## Hashing modules

### md5

```python
load("@ytt:md5", "md5")

md5.sum("data") # "8d777f385d3dfec8815d20f7496026dc"
```

### sha256

```python
load("@ytt:sha256", "sha256")

sha256.sum("data") # "3a6eb0790f39ac87c94f3856b2dd2c5d110e6811602261a9a923d3bb23adc8b7"
```

---
## Overlay module

See [Overlay specific docs](lang-ref-ytt-overlay.md).

---
## Library module

See [Library specific docs](lang-ref-ytt-library.md).
