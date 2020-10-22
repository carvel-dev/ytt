### Structs

Structs are well-defined data objects, comprised of key/value pairs known as "attributes". They are a way to store and refer to data of a known structure.

The most commonly used `struct` is `data.values`, supplied by the [`@ytt:data`](ytt-data-values.md) module.

For example, a data values defined by:

```yaml
#@data/values
---
db_conn:
  host: acme.example.com
```

is automatically processed into a `struct` (named `values`): the keys in the `@data/values` file are defined one-for-one as attributes on the `struct`.
 
Those attribues can be referenced by name:

```yaml
#@ load("@ytt:data", "data")
---
persistence:
  db_url: #@ data.values.db_conn.host
```

_([ytt @data/values](ytt-data-values.md) describes this process in more detail.)_

__

Described below:
- [Attributes](#attributes) — how to refer to a datum within a `struct`
- [Build-in Functions](#built-in-functions) — functions that support using a `struct`
- [Creating `struct`s](#creating-structs) — support for creating your own `struct`s

---

#### Attributes

Attributes are a key/value pair, where the key is a `string` and the value can be of any type.

Attributes can be referenced:
- by field (using "[dot notation](https://github.com/google/starlark-go/blob/master/doc/spec.md#dot-expressions)")
    ```yaml
    db_url: #@ data.values.db_conn.host
    ```
- by string (using "[index notation](https://github.com/google/starlark-go/blob/master/doc/spec.md#index-expressions)") _(as of v0.31.0)_:
    ```yaml
    db_url: #@ data.values["db_conn"]["host"]
    ```
  useful when the name of the attribute is not known statically.

Referencing an attribute that is not defined on the `struct` results in an error:
```yaml
db_url: #@ data.values.bd_conn              # <== struct has no .bd_conn field or method
db_host: #@ data.values["db_conn"]["hast"]  # <== key "hast" not in struct
```

__

#### Built-in Functions

The following built-ins can be useful with `struct` values:

- [`dir()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#dir) enumerates all its attributes.
    ```
    load("@ytt:struct", "struct")
    
    foo = struct.encode({"a": 1, "b": 2, "c": 3})
    keys = dir(foo)  # <== ["a", "b", "c"]
    ```
  
- [`getattr()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#getattr) can be used to select an attribute.
    ```yaml
    #@ for vpc_config in getattr(getattr(data.values.accounts, foundation_name), vpc_name):
       ...
    #@ end
    ```
  - as of v0.31.0, `struct`s support [index notation](https://github.com/google/starlark-go/blob/master/doc/spec.md#index-expressions) behaving identically, more succinctly/readably:
    ```yaml
    #@ for vpc_config in data.values.accounts[foundation_name][vpc_name]:
       ...
    #@ end
    ```

- [`hasattr()`](https://github.com/google/starlark-go/blob/master/doc/spec.md#hasattr) reports whether a value has a specific attribute
    ```python
    # `value` is a struct that _might_ have a field, "additional_ports"
    
    def ports(value):
      if hasattr(value, "additional_ports"):
      ... 
    end
    ```

__

#### Creating `struct`s

`struct` instances can be made using the `@ytt:struct` module.

See [ytt Library: struct module](lang-ref-ytt-struct.md) for details.

