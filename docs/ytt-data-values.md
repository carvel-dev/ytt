## ytt @data/values

- [Defining data values](#defining-data-values)
- [Splitting data values into multiple files](#splitting-data-values-into-multiple-files)
- [Overriding data values via command line flags](#overriding-data-values-via-command-line-flags)
- [Library data values](#library-data-values)
  - [Setting via files](#library-setting-via-files)
  - [Setting via command line flags](#library-setting-via-cmd)

### Defining data values

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

Resulting in

```yaml
first: val1
second: val2
third: null
fourth: null
```

We typically recommend to use snake case (e.g. `some_key.nested_one`) for naming data values.

Note that if data value contains a `-` (dash) in its name, you will have to use `getattr` function like so `getattr(data.values.key2, "nested-with-dash")` to access its value. This is to avoid parsing ambiguity with substraction operation.

### Splitting data values into multiple files

Available in v0.13.0.

It's possible to split data values into multiple files (or specify multiple data values in the same file). `@ytt:data` library provides access to the _merged_ result. Merging is controlled via [overlay annotations](lang-ref-ytt-overlay.md) and follows same ordering as [overlays](lang-ref-ytt-overlay.md#overlay-order). Example:

`values-default.yml`:

```yaml
#@data/values
---
key1: val1
key2:
  nested: val2
key3:
key4:
```

`values-production.yml`:

```yaml
#@data/values
---
key3: new-val3
#@overlay/remove
key4:
#@overlay/match missing_ok=True
key5: new-val5
```

Note that `key4` is being removed, and `key5` is marked as `missing_ok=True` because it doesn't exist in `values-default.yml` (this is a safety feature to prevent accidental typos in keys).

`config.yml`:

```yaml
#@ load("@ytt:data", "data")

first: #@ data.values.key1
third: #@ data.values.key3
fifth: #@ data.values.key5
```

Running `ytt -f .` (or `ytt -f config.yml -f values-default.yml -f values-production.yml`) results in:

```yaml
first: val1
third: new-val3
fifth: new-val5
```

See [Multiple data values example](https://get-ytt.io/#example:example-multiple-data-values) in the online playground.

### Overriding data values via command line flags

(As of v0.17.0+ `--data-value` parses value as string by default. Use `--data-value-yaml` to get previous behaviour.)

ytt CLI allows to override input data via several CLI flags:

- `--data-value` (format: `key=val`, `@lib:key=val`) can be used to set a specific key to string value
  - dotted keys (e.g. `key2.nested=val`) are interpreted as nested maps
  - examples: `key=123`, `key=string`, `key=true`, all set to strings
- `--data-value-yaml` (format: `key=yaml-encoded-value`, `@lib:key=yaml-encoded-value`) same as `--data-value` but parses value as YAML
  - examples: `key=123` sets as integer, `key=string` as string, `key=true` as bool
- `--data-value-file` (format: `key=/file-path`, `@lib:key=/file-path`) can be used to set a specific key to a string value of given file contents
  - dotted keys (e.g. `key2.nested=val`) are interpreted as nested maps
  - this flag can be very useful when loading multine line string values from files such as private and public key files, certificates
- `--data-values-env` (format: `DVAL`, `@lib:DVAL`) can be used to pull out multiple keys from environment variables based on a prefix
  - given two environment variables `DVAL_key1=val1-env` and `DVAL_key2__nested=val2-env`, ytt will pull out `key1=val1-env` and `key2.nested=val2-env` variables
  - interprets values as strings
- `--data-values-env-yaml` (format: `DVAL`, `@lib:DVAL`) same as `--data-values-env` but parses values as YAML

These flags can be repeated multiple times and used together. Flag values are merged into data values last.

Note that for override to work data values must be defined in at least one `@data/values` YAML document.

```bash
export STR_VALS_key6=true # will be string 'true'
export YAML_VALS_key6=true # will be boolean true

ytt -f . \
  --data-value key1=val1-arg \
  --data-value-yaml key2.nested=123 \ # will be int 123
  --data-value-yaml 'key3.other={"nested": true}' \
  --data-value-file key4=/path \
  --data-values-env STR_VALS \
  --data-values-env-yaml YAML_VALS
```

---
### Library data values

Available in v0.28.0+

Each library may specify data values which will be evaluated separately from the root level library.

#### <a id='library-setting-via-files'/> Setting via files

To override library data values, add `@library/ref` annotation to data values YAML document, like so:

```yaml
#@library/ref "@lib1"
#@data/values
---
key1: val1
key2: val2

#@library/ref "@lib1"
#@data/values after_libary_module=True
---
key3: val3
```

The `@data/values` annotation also supports a keyword argument `after_libary_module`. If this keyword argument is specified, given data values will take precedence over data values passed to the `.with_data_values(...)` function when evaluating via the [library module](./lang-ref-ytt-library.md).

#### <a id='library-setting-via-cmd'/> Setting via command line flags

Data value flags support attaching values to libraries for use during [library module](./lang-ref-ytt-library.md) evaluation:

```bash
export STR_VALS_key6=true # will be string 'true'
export YAML_VALS_key6=true # will be boolean true

ytt -f . \
  --data-value @lib1:key1=val1-arg \
  --data-value-yaml @lib2:key2.nested=123 \ # will be int 123
  --data-value-yaml '@lib3:key3.other={"nested": true}' \
  --data-value-file @lib4:key4=/path \
  --data-values-env @lib5:STR_VALS \
  --data-values-env-yaml @lib6:YAML_VALS
```
