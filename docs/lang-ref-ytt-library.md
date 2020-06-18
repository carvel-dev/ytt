### ytt Library: Library annotations

Available in v0.28.0+

- `#@library/ref`: Attaches a yaml document to the specified library to be used during evalutaion via the library module (only supported for [data value documents](./ytt-data-values.md#library-setting-via-files))

```yaml
#@library/ref "@app"
#@data/values
---
name: "app1"
```

Note: data values may also be attached to libraries via [command line flags](ytt-data-values.md#library-setting-via-cmd)

### ytt Library: Library module

Library module `@ytt:library` provides a way to programmatically get result of templates included in a library. Libraries are found within `_ytt_lib` subdirectory.

- `load("@ytt:library", "library")`
```python
# build library instance
app1 = library.get("app")

# build new copy of library with data values (does not mutate app1)
app1_with_vals = app1.with_data_values({"name": "app1"})

# return results of all YAML templates
app1_with_vals.eval()

# return url function defined within app library
url_func = app1_with_vals.export("url")
url_func() # result of url function
```

- `library.get(name)` (`name`: string, returned: library): returns library object that is backed by content under `_ytt_lib/<name>` (found in the same directory as the file containing this call). `name` could contain '/' slashes for directories (e.g. `github.com/k14s/k8s-lib/app`).

- `x.with_data_values(vals)` (`x`: library, `vals`: dict or YAML fragment, returned: library): returns a new library copy with added data values. Given data values are overlayed on top of data values found within library.

```yaml
#@ def app_vals():
name: app1
env_vars:
  #@overlay/match missing_ok=True
  custom_key: val
#@ end

#! with_data_values 
#@ app1_with_vals = app1.with_data_values(app_vals())
```

- `x.eval()` (`x`: library, returned: YAML document set): returns computed YAML document set based on library configuration and data values.

- `x.export(name, [path=])` (`x`: library, `name`: string, `path`: string, returned: any value including function): returns value of a symbol found within a library. Typically used to export a function but could also be used for variables. `path` keyword argument can specify location for the symbol if name is not unique within a library (`path` should not be used unless there are multiple symbols with the same name).

```python
url_func = app1_with_vals.export("url", path="config.lib.yml")
```

Available in v0.28.0+

- `x.data_values()` (`x`: library, returned: data values): returns the data values of the library instance.

```python
app_values = app1_with_vals.data_values()
```

#### Examples

See [ytt-library-module example](../examples/playground/example-ytt-library-module).
