# Schemas

- Originating Issue: https://github.com/k14s/ytt/issues/103
- Status: **Being written** | Being implemented | Included in release | Rejected

## Problem Statement

Software authors want the ability to clearly communicate what data is necessary in order to run their software. Schemas provide the ability for authors to _document_ which data is expected and to _validate_ a consumer's input. This functionality is inspired by [JSON Schema](https://json-schema.org/) but leverages native yaml structure to provide simplicity, brevity, and readability.

## Schema Spec

### Examples

Using [cf-for-k8's data values](cf-for-k8s/values.yml) as an example, this is what the schema would look like in various formats:
- [JSON schema (as json)](cf-for-k8s/json-schema.json)
- [JSON schema (as yaml)](cf-for-k8s/json-schema.yml)
- [ytt native schema](cf-for-k8s/ytt-schema.yml)

### Defining a schema document and what type of documents to apply it to

```yaml
#@schema attach="data/values"
---
#! Schema contents
```

The `#@schema attach="<document-type>"` annotation will create a new schema document. The `attach` argument is used to specify the type of documents the schema will apply to. In this example, the schema will apply to all `data/values` documents.
The `attach` argument will default to `data/values` if the argument is not specified.

### The Type System

Schemas will have a built-in type system similar to Golang's. This decision was made to ensure that when a key is in the schema, its value will be safe to use. In most cases, the type will be inferred from the schema itself without the need for annotations. However, when more control is needed, the `#@schema/type` annotation is available.

The type system will consist of scalar types, "string", "float", "int", "bool", or "none" as well as compound types, "map" and "array". Each compound type will have its key as the name of the type. For example,

```yaml
gcp:
  username: ""
  password: ""
```

will be interpreted as a type named `gcp` defined as a map with keys `username` and `password` of type string.

The default value for arrays are specified using `@schema/default` with type provided in the schema.

```yaml
#@schema/default []
app_endpoints:
- uri: ""
  port: 0
```

The above example will result in `app_endpoints` having a single element of type map with keys `uri` and `port`, types string and integer respectively. The value for `app_endpoints` is defaulted to an empty array. `@schema/default` will accept any value that is of the type specified in the schema.

```yaml
#@ def bucket:
- versioning: "Enabled"
#@ end

#@schema/default bucket()
bucket:
- name: ""
  versioning: ""
  access: ""
```

This snippet will result in an array with a single element, type `bucket`, with `versioning` defaulted to `Enabled`.

### Schema annotations

#### Basic annotations

- `@schema/type`

  This annotation is used to specify the data type of a value when type inference is not sufficient. For example, when multiple types are possible:

    ```yaml
    #@schema/type "int", "float"
    percentage: 0
    ```

  will validate that `percentage` is of type float or integer. `@schema/type` will take one or more string arguments which must come from a predefined set of scalar types: "string", "float", "int", "bool", or "none". The type annotation will also allow a single function argument to specify a complex type to reduce repetition. For example,

  ```yaml
  #@ def iaas:
  #@ 	username:
  #@	password:
  #@ end

  #@schema/type None, func_type=iaas()
  gcp:

  #@schema/type None, func_type=iaas()
  aws:
  ```
  will assert that `gcp` and `aws` keys have either type None, or map with keys `username` and `password`, both strings. 

  Every key will receive the default empty value for its specified type. If None is a specified type, it will be preferred to enable workflows that test for presence. If multiple arguments are provided, the value must be _one of_ the types.

- `@schema/nullable`

  This annotation is used to specify that the type will be inferred from the value given in the schema, or it will be None. 

  Alternatively, use the type keyword argument called `or_inferred` which, when true, will infer the type from the value given in the schema. For example,

    ```yaml
    #@schema/type None, or_inferred=True
    aws:
      password:
      username:
    ```

  will assert that the `aws` key is either type `None` or type `aws` (a map with keys `password` and `username` of type string).

- `@schema/validate`

  This annotation can be used to validate a key's value. `@schema/validate` will provide a set of keyword arguments which map to built-in validations, such as `max`, `max_len`, etc., but will also accept user-defined validation functions. Less common validators will also be provided via a `validations` library. For example,

  ```yaml
  #@schema/validate number_is_even, min=2
  replicas: 6
  ```

  will use the `min` predefined validator as well as the user-provided `number_is_even` function for validations. Function arguments should have a signature which matches the signature of custom functions passed to [`@overlay/assert`](https://github.com/k14s/ytt/blob/master/docs/lang-ref-ytt-overlay.md).

#### Describing annotations

- `@schema/title`
  This annotation provides a way to add a short title to the section of a schema associated with a key, similar to the [JSON schema title field](https://json-schema.org/draft/2019-09/json-schema-validation.html#rfc.section.9.1).

  ```yaml
  #@schema/title "User Password"
  user_password: ""
  ```

  If the annotation is not present, the title will be inferred from the key name by replacing special characters with spaces and capitalizing the first letter similar to [rails humanize functionality](https://apidock.com/rails/String/humanize).

- `@schema/doc`
  This annotation is a way to add a longer description to the section of a schema associated with a key, similar to the JSON Schema description field.

  ```yaml
  #@schema/doc "The user password used to log in to the system"
  user_password: ""
  ```

- `@schema/example`
  This annotation will take one argument that is an example of a value that satisfies the schema.

  ```yaml
  #@schema/example "my_domain.example.com"
  system_domain: ""
  ```

  In this example, the example string "my_domain.example.com" will be attached to the `system_domain` key.

- `@schema/examples`
  This annotation will take one or more tuple arguments which consist of (Title, Example value).

  ```yaml
  #@schema/examples ("Title 1", value1), ("Title 2", title2_example())
  foo: bar
  ```

  In this example, the `"Title 1"` and `"Title 2"` examples and their values are attached to the key `foo`.

- `@schema/deprecated`
  If the user provides a value, a warning that the field has been deprecated is outputted.

- `@schema/removed`
  If the user provides a value, an error is returned.

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
1. Apply data values
1. Overlay with type checks
1. Perform validations on the final document (if there's a validation failure, record the error, short circuit to sibling/parent, and continue validation)

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
