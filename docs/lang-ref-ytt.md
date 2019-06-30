### ytt Library

Following modules are included with `ytt`. Example usage:

```yaml
#@ load("@ytt:regexp", "regexp")

regex: #@ regexp.match("[a-z]+[0-9]+", "__hello123__")
```

#### General

<a id="struct"></a>

- `load("@ytt:struct", "struct")`
```python
st = struct.make(field1=123, field2={"key": "val"})
st.field1 # 123
st.field2 # {"key": "val"}

def _add_method(num, right):
  return num.val + right
end

num123 = struct.make(val=123)

calc = struct.make(add=struct.bind(_add_method, num123))
calc.add(3) # 126

# make_and_bind automatically binds any callable value to given object
calc = struct.make_and_bind(num123, add=_add_method)
calc.add(3) # 126

struct.encode({"a": [1,2,3,{"c":456}], "b": "str"})      # struct with contents
struct.decode(struct.make(a=[1,2,3,{"c":456}], b="str")) # plain values extracted from struct

# plain values extracted from data.values struct
struct.decode(data.values) # {...}
```

- `load("@ytt:assert", "assert")`
```python
assert.fail("expected value foo, but was {}".format(value)) # stops execution
x = data.values.env.mysql_password or assert.fail("missing env.mysql_password")
```

- `load("@ytt:data", "data")` (see [ytt @data/values](ytt-data-values.md) for more details)
```python
data.values                # struct that has input values
data.list()                # ["template.yml", "data/data.txt"]
data.read("data/data.txt") # "data-txt contents"
```

- `load("@ytt:regexp", "regexp")`
```python
regexp.match("[a-z]+[0-9]+", "__hello123__") # True
```

- `load("@ytt:url", "url")`
```python
url.path_segment_encode("part part")   # "part%20part"
url.path_segment_decode("part%20part") # "part part"

url.query_param_value_encode("part part") # "part+part"
url.query_param_value_decode("part+part") # "part part"

url.query_params_encode({"x":["1"],"y":["2","3"],"z":[""]}) # "x=1&y=2&y=3&z="
url.query_params_decode("x=1&y=2&y=3;z")                    # {"x":["1"],"y":["2","3"],"z":[""]}
```

#### Serialization

- `load("@ytt:base64", "base64")`
```python
base64.encode("regular")      # "cmVndWxhcg=="
base64.decode("cmVndWxhcg==") # "regular"
```

- `load("@ytt:json", "json")`
```python
json.encode({"a": [1,2,3,{"c":456}], "b": "str"})
json.decode('{"a":[1,2,3,{"c":456}],"b":"str"}')
```

- `load("@ytt:yaml", "yaml")`
```python
yaml.encode({"a": [1,2,3,{"c":456}], "b": "str"})
yaml.decode('{"a":[1,2,3,{"c":456}],"b":"str"}')
```

#### Hashes

- `load("@ytt:md5", "md5")`
```python
md5.sum("data") # "8d777f385d3dfec8815d20f7496026dc"
```

- `load("@ytt:sha256", "sha256")`
```python
sha256.sum("data") # "3a6eb0790f39ac87c94f3856b2dd2c5d110e6811602261a9a923d3bb23adc8b7"
```

#### Library Loading

 - `load("@ytt:lib", "lib")`
    ```python
    lib.get(LIBNAME)
    ```

    @ytt:lib can be loaded from within starlark and it provides a `lib` module
    with a `lib.get()` function that returns a library object.

    The library name `LIBNAME` must match one of the forms:

    - `@LIBNAME:[PACKAGE/MODULE]`: loads a library name. The library is found in
      the `_ytt_lib` directory at the root of the current library.

    - `LIBPATH:[PACKAGE/MODULE]`: loads a library path relative to current
       package path. The library is loaded relative to the current executing
       file location.

    - `@:[PACKAGE/MODULE]`: loads the current library. The package specified is
      loaded relative to the current library root.

    - `PACKAGE/MODULE`: loads a package in the current library relative to the
       current executing file location.

    If a package and module is specified, only this package and module will be
    evaluated. Package evaluation is always recursive.

    The library is not evaluated until one of its methods requires its
    evaluation. The library object takes the following methods:

    - `.data_values({...})`: Returns a new library object with its data values
      overlayed using the provided argument

    - `.yaml()`: Returns the YAML result of the library evaluation. If an
      argument is specified, it must be a list of file names. Only the result
      for the specified file names will be returned.

    - `.text()`: Returns the text result of the library evaluation. If an
      argument is specified, it must be a list of file names. Only the result
      for the specified file names will be returned.

    - `exports()`: Returns a dict of library exports. Either the complete
      library exports are returned, or if a file name is specified, only the
      exports for this file are returned.

    - `list()`: Returns the list of files evaluated.

    Examples:

    ```python
    libname = lib.get("@libname:").data_values({"data": "value"}) # Loads _ytt_lib/libname as if --data-value data=value was set
    libname.yaml() # Returns the evaluated data structure
    libname.text() # Returns the evaluated library as a string

    libname2 = lib.get("_ytt_lib/libname:").data_values({"data": "value"}) # equivalent to the above lib.get(...)
    libname2.yaml("main.yml") # Returns the evaluated data structure for main.yml
    libname2.text("main.yml") # Returns the evaluated main.yml file as a string
    ```

#### Overlay

See [Overlay specific docs](lang-ref-ytt-overlay.md).
