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
