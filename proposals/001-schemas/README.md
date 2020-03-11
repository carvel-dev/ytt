## Schemas

- Issue: https://github.com/k14s/ytt/issues/103
- Status: **Being written** | Being implemented | Included in release | Rejected

### Examples

- Schemas for [cf-for-k8s's values](cf-for-k8s/values.yml)
  - [JSON schema (as json)](cf-for-k8s/json-schema.json)
  - [JSON schema (as yaml)](cf-for-k8s/json-schema.yml)
  - [ytt native schema](cf-for-k8s/ytt-schema.yml)

### Defining a schema document and what type of documents to apply it to

```yaml
#@schema attach="data/values"
---
#! Schema contents
```

Will create a new schema document. The attach keyword is used to specify the type of documents the schema will apply to, in this example, the schema will apply to all `data/values` documents.
The attach argument will default to `data/values` if the argument is not specified.

### Schema annotations

#### Basic annotations

- `@schema/type`
  This annotation is used to specify the data type of a value. For example,
  ```yaml
  #@schema/type "array"
  app_domains: []
  ```
  will validate that `app_domain` has a value of type array. The arguments to the type assertion will be strings from a predefined set i.e. "string", "array", "int", etc...

  If a type annotation is not given, ytt will infer the type based on the key's value in the schema document. For example,
  ```yaml
  app_domains: []
  ```
  will also result in the `app_domains` key requiring a value with type array

- `@schema/validate`
  This annotation can be used to validate a key's value. Validate will provide a set of keyword arguments which map to built-in validations, such as max, max_len, etc..., but will also take user defined validation functions. For example,
  ```yaml
  #@schema/validate number_is_even, min=2
  replicas: 6
  ```
  will use the `min` predefined validator as well as the user provided `number_is_even` function for validations. Function arguments should have a signature which matches the signature of custom functions passed to overlay/assert.

- `@schema/allow-empty`
  ytt defaults to requiring any key specified in the schema to have a non-empty value after including data values. This annotation overrides that default and allows keys to have the empty value of its type i.e. "", [], 0. This is useful for defining keys that are optional and could be avoided as data values. For example,
  ```yaml
  #@schema/allow-empty
  system_domain: ""
  ```
  will make the `system_domain` key optional as a data value because, although its default value is empty string, the annotation overrides the requirement that the value be non-empty.


#### Describing annotations

- `@schema/title -> title for node (can maybe infer from the key name ie app_domain -> App domain)`
  This annotation provides a way to add a short title to the section of a schema associated with a key, similar to the JSON schema title field.
  ```yaml
  #@schema/title "User Password"
  user_password: ""
  ```
  If the annotation is not present, the title will be inferred from the key name by replacing special characters with spaces and capitalizing the first letter similar to rails humanize functionality.

- `@schema/doc`
  This annotation is a way to add a longer description to the section of a schema associated with a key, similar to the JSON Schema description field.
  ```yaml
  #@schema/doc "The user password used to log in to the system"
  user_password: ""
  ```

- `@schema/example`
  Example will take one argument which is an example of a value which satisfies the schema
  ```yaml
  #@schema/example "my_domain.example.com"
  system_domain: ""
  ```
  In this example, the example string "my_domain.example.com" will be attached to the `system_domain` key.

- `@schema/examples`
  Examples will take one or more tuple arguments which consist of (Title, Example value)
  ```yaml
  #@schema/examples ("Title 1", value1), ("Title 2", title2_example())
  foo: bar
  ```
  In this example, the Title 1 and Title 2 examples and their values are attached to the key `foo`.

#### Map key presence annotations

- `@schema/any-key`
  Applies to items of a map, allowing users to assert on item structure while maintaining freedom to have any key name.
  ```yaml
  connection_options:
    #@schema/any-key
    _: 
    - ""
  ```
  This example requires items in the `connection_options` map to have value type array of strings.

- `@schema/key-may-be-present`
  This annotation can be used to allow keys in the schema without providing a guarantee they will be present. This allows schemas to validate contents of a structure in cases where the contents are not referenced directly. For example,
  ```yaml
  connection_options:
    #@schema/key-may-be-present
    pooled: true
  ```
  will allow the key `pooled` but will not guarantee its presence. This is useful when referencing the containing structure, such as #@ data.values.connection_options, in a template instead of the key directly: #@ data.values.connection_options.pooled. See more advanced examples below for more.

#### Complex schema annotations

- `@schema/any-of`
  Requires the key to satisfy _at least one_ of the provided schemas
  ```yaml
  #@schema/any-of schema1(), schema2()
  foo: ""
  ```
  Note, any values passed to the `any-of` call will be interpreted as schema documents.

- `@schema/all-of`
  Requires the key to satisfy _every_ provided schema
  ```yaml
  #@schema/all-of schema1(), schema2()
  foo: ""
  ```
  Note, any values passed to the `all-of` call will be interpreted as schema documents.

- `@schema/one-of`
  Requires the key to satisfy _exactly one_ of provided schema
  ```yaml
  #@schema/one-of schema1(), schema2()
  foo: ""
  ```
  Note, any values passed to the `one-of` call will be interpreted as schema documents.

### Sequence of events
1. Extract defaults from the provided schemas
2. Apply data values, validating against the schema validations after each document
3. Apply validated data values to templates

### Complex Examples

#### Asserting on higher level structure with optional lower key with #@schema/key-may-be-present
```yaml
#@schema
---
foo: ""

#@data/values
---
foo: "val"

config.yml:
---
foo: #@ data.values.foo #! never errors with: no member 'foo' on data.values
```

```yaml
#@schema
---
#@schema/key-may-be-present
foo:

#@data/values
---
foo: "val"

config.yml:
---
foo: #@ yaml.encode(data.values) #! => {}, {"foo": "val"}
```

```yaml
#@schema
---
#@schema/key-may-be-present
foo:
  max_connections: 100
  username: ""

#@data/values
---
foo:
  username: val

config.yml:
---
foo: #@ yaml.encode(data.values) #! => {},  {"foo": {"username": "val", "max_connections":100}}
```
