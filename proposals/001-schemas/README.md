### Defining a schema document and what type of documents to apply it to

```yaml
#@schema attach="data/values"
---
#! Schema contents
```

Will create a new schema document. The attach keyword is used to specify what types of documents the schema will apply to, in this case, `data/values` documents.
The attach argument will default to `data/values` if none is provided.

### Schema annotations

- #@schema/non-empty
  This annotation asserts that a key cannot have any empty value of its type i.e. "", [], 0, etc. This is useful for defining keys that are not optional and must
  be provided as data values.
  ```yaml
  #@schema/non-empty
  system_domain: ""
  ```
  Because the schema values are extracted and used as defaults, by setting the value of `sytem_domain` to empty string and requiring it to not be empty after interpolating the data values,
  the user is forced to provide a non-empty `system_domain` through their data values. ytt uses starlark [type truth values](https://github.com/google/starlark-go/blob/master/doc/spec.md#data-types) to determine if a value is empty.

- #@schema/type
  This annotation asserts on the type of a keys value. For example,
  ```yaml
    #@schema/type "array"
    app_domains: []
  ```
  will validate that any data value `app_domain` will be of type array. These strings are a predefined set: string, array, int, etc...
  ytt will also infer the type based on the key's value given in a schema document if the annotation is not provided. For example,
  ```yaml
    app_domains: []
  ```
  will also result in the `app_domains` key requiring value type array

- #@schema/validate
  This annotation can be used in order to run a validation function on the value of the key it is applied to. For example,
  ```yaml
    #@schema/validate  number_is_even, max=10
    foo: 5
  ```
  will use the `max` predefined validator as well as the user provided `number_is_even` function to validate the value of `foo`. The funtion signature should match what gets passed to overlay/assert annotations.

- #@schema/example
  Examples will take one arguments which consist of the example
  ```yaml
    #@schema/example "my_domain.example.com"
    system_domain: ""
  ```
  In this example, the example string "my_domain.example.com" will be attached to the `system_domain` key.

- #@schema/examples
  Examples will take one or more tuple arguments which consist of {Title, Example value}
  ```yaml
    #@schema/examples ("Title 1", value1), ("Title 2", title2_example())
    foo: bar
  ```
  In this example, the Title 1 and Title 2 examples and their values are attached with the key `foo`.


- #@schema/title -> title for node (can maybe infer from the key name ie app_domain -> App domain)
  This annotation provides a way to add a short title to the schema applied to a key.
  ```yaml
    #@schema/title "User Password"
    user_password: ""
  ```
  If the annotation is not present, the title will be infered from the key name by replacing special characters with spaces and capitalizing the first letter similar to rails humanize functionality.

- #@schema/doc
  This annotation is a way to add a longer description of the schema applied to a key, similar to the JSON Schema description field.
  ```yaml
    #@schema/doc "The user password used to log in to the system"
    user_password: ""
  ```

- #@schema/any-of
Requires the key to satisy _at least one_ of the provided schemas
```yaml
  #@schema/any-of schema1, schema2()
  foo: ""
```
Note, any values passed to the `any-of` call will be interpreted as schema documents.

- #@schema/all-of
Requires the key to satisy _every_ provided schema
```yaml
  #@schema/all-of schema1, schema2()
  foo: ""
```
Note, any values passed to the `all-of` call will be interpreted as schema documents.

- #@schema/one-of
Requires the key to satisy _exactly one_ of provided schema
```yaml
  #@schema/one-of schema1, schema2()
  foo: ""
```
Note, any values passed to the `one-of` call will be interpreted as schema documents.

- #@schema/any_key -> Applies to items of a map. Allows users to assert on structure while maintaining freedom to have any key
Allows a schema to assert on the structure of map items without making any assertions on the value of the key
```yaml
  foo:
    #@schema/any-key
    _: 
    - bar
    - baz
```
This example will allow the map to contain any key names that have values structure [bar, baz]

- #@schema/key-may-be-present
  This annotation asserts a key is allowed by the schema but is not guaranteed. This allows schemas to validate contents of a structure in cases where
  the contents are not referenced directly. For example,
  ```yaml
  connection_options:
    #@schema/key-may-be-present
    pooled: true
  ```
  will assert that the key `pooled` is allowed under the schema but not guaranteed. This would be useful when accessing something like #@ data.values.connection_options in a template
  instead of #@ data.values.connection_options.pooled. See more advanced examples below for more.


### Sequence of events
1. Extract defaults from the provided schemas
2. 

assert on higher level structure with optional lower key
```yaml
schema
---
foo: ""data/values
---
foo: "dmitriy"config.yml:
---
foo: #@ data.values.foo   no member 'foo' on data.valuestype Foo struct 
  Data interface{}schema
---
#@schema/key-may-be-there
foo:data/values
---
foo: "dmitriy"config.yml:
---
foo: #@ yaml.encode(data.values) => {},  {"foo": "dmitriy"}schema
---
#@schema/key-may-be-there
foo:
  max_connections: 100
  username: ""data/values
---
foo:
  username: dmitriyconfig.yml:
---
foo: #@ yaml.encode(data.values) => {},  {"foo": {"username": "dmitriy", "max_connections":100}}
```
