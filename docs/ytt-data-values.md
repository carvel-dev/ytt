## ytt @data/values

One way to inject input data into templates is to include a YAML document annotated with `@data/values`. Example:

```yaml
#@data/values
---
key1: val1
key2:
  nested: val2
key3:
key4:
```

Subsequently these values can be accessed via `@ytt:data` library:

```yaml
#@ load("@ytt:data", "data")

first: #@ data.values.key1
second: #@ data.values.key2.nested
third: #@ data.values.key3
fourth: #@ data.values.key4
```

Additionally ytt CLI allows to override input data via several CLI flags:

- `--data-value` (format: `key=yaml-encoded-value`) can be used to set a specific key to a YAML encoded value
  - dotted keys (e.g. `key2.nested=val`) are interpreted as nested maps
  - examples: `key=123`, `key="string"`, `key=true`
- `--data-value-file` (format: `key=/file-path`) can be used to set a specific key to a string value of given file contents
  - dotted keys (e.g. `key2.nested=val`) are interpreted as nested maps
  - this flag can be very useful when loading values from secret files such as private and public key files, certificates
- `--data-values-env` (format: `DVAL`) can be used to pull out multiple keys from environment variables based on a prefix
  - given two environment variables `DVAL_key1=val1-env` and `DVAL_key2__nested=val2-env`, ytt will pull out `key1=val1-env` and `key2.nested=val2-env` variables

These flags can be repeated multiple times and used together.

```bash
export VAL_key4=true
ytt template -f . \
  --data-value key1="val1-arg" \
  --data-value key2.nested=123 \
  --data-value-file key3=/path \
  --data-values-env VAL
```
