### ytt Library: struct module

The `@ytt:struct` module provides functions for constructing and deconstructing [`struct`](lang-ref-structs.md) values.

To use these functions, include the `@ytt:struct` module:
 
```python
load("@ytt:struct", "struct")
```

The functions exported by this module are:
- [struct.decode()](#structdecode)
- [struct.encode()](#structencode)
- [struct.make()](#structmake)
- [struct.make_and_bind()](#structmake_and_bind)

__

#### struct.decode()
Deconstructs a given value into plain/Starlark values, recursively.

```python
struct.decode(struct_val)
```

- `struct_val` ([`struct`](lang-ref-structs.md)) — the value to decompose.
  - `struct` values are converted into `dict` values where each attribute in the `struct` becomes a 
    key on the `dict`.
  - if the value of an attribute is a `struct`, it is likewise converted to a `dict`.
  - all other values are copied, as is.


**Examples:** 

_Example 1: _

```python
load("@ytt:struct", "struct")

foo = struct.encode({"a": [1,2,3,{"c":456}], "b": "str"})
bar = struct.decode(foo)

bar["a"]  # <== [1, 2, 3, {"c": 456}]
```


__

#### struct.encode()

Makes a `struct` of a given value, recursively. 

```python
struct.encode(value)
```

- `value` ([`dict`](lang-ref-dict.md) | [`list`](lang-ref-list.md) | scalar) — the value to encode.
  - `dict` values are converted into `struct`s where each key in the `dict` becomes an attribute on the `struct`.
    Keys of the items in `dict` values must be strings.
  - if a `dict` or `list` contains a value that is a `dict`, it is likewise converted to a `struct`.

**Examples:** 

_Example 1: Data structure from a dictionary_

```python
load("@ytt:struct", "struct")

d = struct.encode({"a": [1,2,3,{"c":456}], "b": "str"})
d.a        # <== [1, 2, 3, c: (struct...)]
d.a[3].c   # <== 456
bar["b"]   # <== "str"
```

__

_Example 2: Collection of functions_

"Export" a set of functions from a library file.

`urls.star`
```python
load("@ytt:struct", "struct")

def _valid_port(port):
...
end

def _url_encode(url):
...
end

urls = struct.encode({"valid_port": _valid_port, "encode": _url_encode, ...}) 

```

```yaml
#@ load("urls.star", "urls")

#@ if/end urls.valid_port(...):

```
__

#### struct.make()

Instantiates a `struct` based on the key/value pairs provided.

```python
struct.make(key1=value1, key2=value2, ...)
```

- `keyN` (keyword argument name) — becomes the name of the Nth attribute in the constructed
  `struct`. 
- `=valueN`(`any`) — becomes the value of the Nth attribute in the constructed `struct`.

**Notes:**
- `make()` does not modify `values` in any way (e.g. if `valueN` is a dictionary, it is
  _not_ converted into a `struct`). To recursively build a hierarchy of `struct`s, see [`struct.encode()`](#structencode).
  

**Examples:** 

_Example 1: Scalar values_

For visually pleasing collections of fields
```python
load("@ytt:struct", "struct")

consts = struct.make(version="0.39.0", service_name="prometheus")

consts.version      # <== "0.39.0"
consts.service_name # <== "prometheus"
```

__

_Example 2: Data structures_

Dictionaries values remain instances of `dict`.
```python
load("@ytt:struct", "struct")

consts = struct.make(service={"version": "0.39.0", "name": "prometheus"})

consts.service["version"]  # <== "0.39.0"
consts.service["name"]     # <== "prometheus"
# const.service.version    # Error! "dict has no .version field or method"
```
__

_Example 3: Nested `struct`s_

Nested invocations of `make()` to retain dot expression access.
```python
load("@ytt:struct", "struct")

consts = struct.make(service=struct.make(version="0.39.0", name="prometheus"})

consts.service.version  # <== "0.39.0"
consts.service.name     # <== "prometheus"
```
See also: [`struct.encode()`](#structencode) to convert all `dict` values to `struct`s, recursively.


__

#### struct.make_and_bind()

Binds one or more function(s) to a `struct`, making them method(s) on that struct.
This allows `struct`s to carry both data and behavior related to that data.

```python
struct.make_and_bind(receiver, method_name1=function1, ...)
```

- `receiver` ([`struct`](lang-ref-structs.md)) — "object" to attach the function(s) to.
- `method_nameN` (keyword argument name) — the name that callers will specify to invoke the method.
- `functionN` ([`function`](lang-ref-def.md)) — the function value (either the name of a `function` or a `lambda` expression)
  that will be bound to `receiver` by the name `method_nameN`.
  - the first parameter of `functionN` is `receiver`, implicitly.
  - the remaining parameters of `functionN` become the parameters of `receiver.method_nameN()`

**Notes:**
- Binding is useful for cases where a commonly desired value is a calculation of two or more
  values on `receiver`.
  
**Examples:** 

_Example 1: Binding a function value_

```python
load("@ytt:struct", "struct")

conn = struct.make(hostname="svc.example.com", default_port=1022, protocol="https")

def _url(self, port=None):
  port = port or self.default_port
  return "{}://{}:{}".format(self.protocol, self.hostname, port)
end

conn = struct.make_and_bind(conn, url=_url)

conn.url()      # ==> https://svc.example.com:1022
conn.url(8080)  # ==> https://svc.example.com:8080
```

__

_Example 2: Binding a lambda expression_

```python
load("@ytt:struct", "struct")

_conn_data = struct.make(hostname="svc.example.com", default_port=1022, protocol="https")

conn = struct.make_and_bind(_conn_data, url=lambda self, port=None: "{}://{}:{}".format(self.protocol, self.hostname, port or self.default_port))

conn.url()      # ==> https://svc.example.com:1022
conn.url(8080)  # ==> https://svc.example.com:8080
```

