# Schemas

- Originating Issue: https://github.com/k14s/ytt/issues/103
- Status: **Being written** | Being implemented | Included in release | Rejected

## Problem Statement

Software authors want the ability to clearly communicate what data is necessary in order to run their software. Schemas provide the ability for authors to _document_ which data is expected and to _validate_ a consumer's input. This functionality is inspired by [JSON Schema](https://json-schema.org/) but leverages native yaml structure to provide simplicity, brevity, and readability.

## Schema Spec

### Examples

- Using [cf-for-k8's data values](cf-for-k8s/values.yml) as an example, this is what the schema would look like in various formats:
  - [JSON schema (as json)](cf-for-k8s/json-schema.json)
  - [JSON schema (as yaml)](cf-for-k8s/json-schema.yml)
  - [ytt native schema](cf-for-k8s/ytt-schema.yml)

- [Dex](example-partial-dex.yaml)

### Defining a schema document

```yaml
#@schema/match data_values=True
---
#! Schema contents
```

The `#@schema/match` annotation will create a new schema document. The `data_values=True` will apply to all `@data/values`.

### The Type System

Schemas will have a built-in type system similar to Golang's. This decision was made to ensure that when a key is in the schema, its value will be safe to use. In most cases, the type will be inferred from the value itself without the need for explicit annotation. However, when more control is needed, the `@schema/type` annotation is available.

Following types are available:

- `string`
- `float`
- `int`
- `bool`
- `null`
- `map-typed <unique>` (unique indicates that each map defintion is a unique, even if they are same)
- `array-typed <unique>`
- `any` (TBD: vs `#@schema/type any=True`)

`map-typed`, `array-typed`, and `any` are "complex values".

Tentative: use `any` when wanting to say "all types"

TBD: use starlark or yaml terminology (list vs array, map vs dict)?


### Schema annotations

#### Valule type annotations

- `@schema/type string, [...string,] [or_inferred=True]`

  This annotation is used to specify the data type of a value (by listing allowed types) when type inference is not sufficient. For example, when multiple types are allowed:

    ```yaml
    #@schema/type "int", "float"
    percentage: 0
    ```

  will validate that `percentage` is of type float or integer.

  Since `map-typed <unique>` and `array-typed <unique>` types can _only_ be described inline (as a value), `or_inferred=True` must be specified.
  
  `or_inferred=True` can be used iff the value is `map-typed <unique>` or `array-typed <unique>`.

- `@schema/default value`

  This annotation is used to specify default value. It's not necessary in most cases since value provided is used as a default; however, for arrays or multi-type values it's typically necessary to avoid ambiguity.

  ```yaml
  #@schema/type None, or_inferred=True
  #@schema/default None
  aws:
    access_key: ""
    secret_key: ""
  ```

  In this example, since value of aws could be null or aws typed map, and aws typed map definition takes up the value, one has to specify `@schema/default None` to set the default value to null.

  ```yaml
  #@schema/default 3 # <-- error; redundant
  aws: 4
  ```

  `@schema/default` is _disallowed_ when the effective type is:
  - any scalar type or compound type made up only of scalars
  - "None"
  - "Any" 
      
  `@schema/default` is _required_ for `array-typed <unique>` values to indicate explicitly to readers of schema what is the default value.

- `@schema/nullable`

  This annotation is used to specify that the type will be inferred from the value (similar to default behaviour for all declarations) given in the schema, _and_ it is allowed to be null.

    ```yaml
    #@schema/nullable
    aws:
      username: ""
      password: ""
    ```

  It is equivalent for the following `@schema/type` and `@schema/default` declaration:

    ```yaml
    #@schema/type None, or_inferred=True
    #@schema/default None
    aws:
      username: ""
      password: ""
    ```

  In following example, `@schema/nullable` allows cf_db to be set to null; however, by default it is enabled, hence no a null. (TBD better way?)

    ```yaml
    #@schema/nullable
    #@schema/default {}
    cf_db:
      admin_password: ""
    ```

#### Value validation annotations

Beyond specifying a type for a value, one can specify more dynamic constraints on the value via validations.

- `@schema/validate [...user-provided-funcs,] [...builtin-kwargs]`

  This annotation specifies how to validate value. `@schema/validate` will provide a set of keyword arguments which map to built-in validations, such as `max`, `max_len`, etc., but will also accept user-defined validation functions. Less common validators will also be provided via a `validations` library. For example,

  ```yaml
  #@ def number_is_even(num): return (num % 2 == 0, "Number is not even")

  #@schema/validate number_is_even, min=2
  replicas: 6
  ```

  will use the `min` predefined validator as well as the user-provided `number_is_even` function for validations. Function arguments should have a signature which matches the signature of custom functions passed to [`@overlay/assert`](https://github.com/k14s/ytt/blob/develop/docs/lang-ref-ytt-overlay.md).

  Full list of builtin validations exposed via kwargs:

  - min=int
  - max=int
  - min_len=int
  - max_len=int
  - enum=[string|int|...]: any set of values
  - regexp=string (TBD or pattern?) -> golang regexp
  - format=string (mostly copied from JSON Schema)
    - date-time: Date and time together, for example, 2018-11-13T20:20:39+00:00.
    - time: New in draft 7 Time, for example, 20:20:39+00:00
    - date: New in draft 7 Date, for example, 2018-11-13.
    - email: Internet email address, see RFC 5322, section 3.4.1.
    - hostname: Internet host name, see RFC 1034, section 3.1.
    - ip: IPv4 or IPv6
    - ipv4: IPv4 address, according to dotted-quad ABNF syntax as defined in RFC 2673, section 3.2.
    - ipv6: IPv6 address, as defined in RFC 2373, section 2.2.
    - uri: A universal resource identifier (URI), according to RFC3986.
    - port: >= 0 && <= 65535
    - percent: k8s's IsValidPercent e.g. "10%"
    - duration: Golang's duration e.g. "10h"
  - not_null=bool to verify value is not null
  - unique=bool to verify value contains unique elements (TBD?)
  - prefix=string to verify string has prefix (TBD?)
  - TBD

  Full list of builtin validations included in a library:

  - ...kwargs validations as functions...
  - base64.decodable()
  - json.decodable()
  - yaml.decodable()

  String values, by default, have `@schema/validate min_len=1` validation added.

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

  If the user provides a value, a warning that the field has been deprecated is outputted (stderr) to the user.

- `@schema/removed`

  If the user provides a value, an error is returned, and evaluation stops.

#### Map key presence annotations

- `@schema/any-key` (TBD: any other name for this?)

  Applies to multiple KVs of a map that were not matched by explicitly specified keys, allowing users to assert on item structure while maintaining freedom to have any key name.
  ```yaml
  connection_options:
    #@schema/any-key
    _: 
    #@schema/type "any"
    - null
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

### Sequence of events

1. Extract defaults from the provided schemas
1. Apply data values
1. Overlay with type checks
1. Perform validations on the final document (if there's a validation failure, record the error, short circuit to sibling/parent, and continue validation)

### Complex Examples

This snippet will result in an array with a single element, type `bucket`, with `versioning` defaulted to `Enabled`.

```yaml
#@ def default_bucket():
versioning: "Enabled"
#@ end

#@schema/default [default_bucket()]
bucket:
- name: ""
  versioning: ""
  access: ""
```

Partial minio config (based on https://github.com/helm/charts/blob/21e2e1b1f2656c785ece0ec741b047b21539d7b1/stable/minio/values.yaml#L244), shows how to have one-of available options.

```yaml
#@ load("@ytt:json", "json")

#@schema/match data_values=True
---
#@schema/validate lambda g: (g.s3 or g.azure or g.gcs or g.oss, "Must specify one gateway: s3, azure, gcs, oss")
gateway:
  #@schema/nullable
  s3:
    #@schema/validate min=1
    replicas: 4
    #@schema/validate prefix="https://"
    serviceEndpoint: ""
    accessKey: ""
    secretKey: ""

  #@schema/doc "Use minio as an azure blob gateway, you should disable data persistence so no volume claim are created. https://docs.minio.io/docs/minio-gateway-for-azure"
  #@schema/nullable
  azure:
    #@schema/doc "Number of parallel instances"
    #@schema/validate min=1
    replicas: 4

  #@schema/doc "Use minio as GCS (Google Cloud Storage) gateway, you should disable data persistence so no volume claim are created. https://docs.minio.io/docs/minio-gateway-for-gcs"
  #@schema/nullable
  gcs:
    #@schema/doc "Number of parallel instances"
    #@schema/validate min=1
    replicas: 4
    #@schema/doc "credential json file of service account key"
    #@schema/validate lambda val: (json.decodable(val), "key must be valid JSON")
    gcsKeyJson: ""
    #@schema/doc "Google cloud project-id"
    projectId: ""

  #@schema/nullable
  oss:
    #@schema/doc "Number of parallel instances"
    #@schema/validate min=1
    replicas: 4
    #@schema/validate prefix="https://"
    endpointURL: ""
```

Array examples:

```yaml
service:
  #@schema/doc "List of IP addresses at which the Prometheus server service is available. Ref: https://kubernetes.io/docs/user-guide/services/#external-ips"
  #@schema/default []
  externalIPs:
  #@schema/validate format="ip"
  - ""

#@schema/default []
imagePullSecrets:
- name: ""

ingress:
  #@schema/default ["chart-example.local"]
  hosts:
  - ""
  #@schema/default []
  tls:
  - secretName: ""
    #@schema/default []
    #@schema/validate min=1
    hosts:
    - ""

#@schema/default []
tolerations:
#@schema/type "any"
- null

#@schema/doc "Create multiple buckets after minio install. Enabling `defaultBucket` will take priority over this list"
#@schema/example [{"name": "bucket1"}]
buckets:
- name: ""
  #@schema/validate enum=["none"]
  policy: none
  purge: false
```

map[string]string examples:

```yaml
#@ def validate_map_str_to_str(vals):
#@   type(vals) == "dict" or assert.fail("Expected key to be map")
#@   for key in vals:
#@     type(key) == "string" or assert.fail("Expected key to be string")
#@     type(vals[key]) == "string" or assert.fail("Expected val to be string")
#@   end
#@ end

#@schema/type "any"
#@schema/validate validate_map_str_to_str
#@schema/example {"prometheus.io/scrape": "true", "prometheus.io/path": "/minio/prometheus/metrics", "prometheus.io/port": "9000"}
annotations: {}

#! ...

#@schema/example {"prometheus.io/scrape": "true", "prometheus.io/path": "/minio/prometheus/metrics", "prometheus.io/port": "9000"}
annotations:
  #@schema/any-key
  _: ""
```

Type any with default (TBD, should we force schema/default?):

```yaml
#@schema/type "any"
annotations:
  foo: bar
  blah: 123
```

Example of additional free-form properties:

```yaml
annotations:
  #@schema/any-key
  _: ""

db_opts:
  username: ""
  password: ""
  #@schema/any-key
  _:
    #@schema/type "any"
    value:
```

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
