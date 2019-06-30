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
```

- `load("@ytt:assert", "assert")`
```python
assert.fail("expected value foo, but was {}".format(value)) # stops execution
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

#### Serialization

- `load("@ytt:base64", "base64")`
```python
base64.encode("regular") # "cmVndWxhcg=="
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
md5.sum("data") # 8d777f385d3dfec8815d20f7496026dc
```

- `load("@ytt:sha256", "sha256")`
```python
sha256.sum("data") # 3a6eb0790f39ac87c94f3856b2dd2c5d110e6811602261a9a923d3bb23adc8b7
```

#### External libraries

- `load("@ytt:lib", "lib")` `lib.eval(LIBNAME, [DATAVALUES])`

    @ytt:lib can be loaded from within starlark and it provides a `lib`
    modules with a `lib.eval()` function that load and evaluates a library.
    It takes:

    -   A string with the location of the module to load. It can start with
        `@` in which case it is assumed that the library comes from
        `_ytt_lib/`
    -   An optional value object to override current data/values

    After evaluation, the returned object has `output()` and `text()`
    methods to fetch the result from the library evaluation. If called
    without argument, it returns the complete evaluation. Else, it takes a
    file name and returns the evaluation results for that file. `text()`
    always returns a string in YAML form. `output()` returns the YAML data
    structure.


    ```python
    libname = lib.eval("@libname", {"data": "value"}) # Loads _ytt_lib/libname as if --data-value data=value was set
    libname.output() # Returns the evaluated data structure
    libname.text() # Returns the evaluated library as a string

    libname2 = lib.eval("_ytt_lib/libname", {"data": "value"}) # equivalent to the above lib.eval(...)
    libname2.output("main.yml") # Returns the evaluated data structure for main.yml
    libname2.text("main.yml") # Returns the evaluated main.yml file as a string
    ```

#### Overlay

See [Overlay specific docs](lang-ref-ytt-overlay.md).
