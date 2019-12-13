### Functions

Refer to [Starlark function specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#functions) for details about various types of function arguments. Note that ytt's Starlark use requires functions to be closed with an `end`.

- function definition within YAML

Labels returns map with two keys: test1, and test2:

```yaml
#@ def my_labels():
test1: 123
test2: 124
#@ end
```

Above is _almost_ equivalent to (differnce is that return type in one case is a YAML fragment and in another it's a dict):

```yaml
#@ def my_labels():
#@   return {"test1": 123, "test2": 124}
#@ end
```

- function definition within Starlark (.star files)

```python
def my_labels():
  return {"test1": 123, "test2": 124}
end
```

- function arguments (positional and keyword arguments)

```yaml
#@ def my_deployment(name, replicas=1, labels={}):
kind: Deployment
metadata:
  name: #@ name
  labels: #@ labels
spec:
  replicas: #@ replicas
#@ end

---
kind: List
items:
- #@ my_deployment("dep1", replicas=3)
```

- common function usage

To set `labels` key to return value of `my_labels()`:

```yaml
labels: #@ my_labels()
labels_as_array:
- #@ my_labels()
```

To merge return value of `my_labels()` into `labels` map:

```yaml
#@ load("@ytt:template", "template")

labels:
  another-label: true
  _: #@ template.replace(my_labels())
```

Note that in most cases `template.replace` is not necessary since it's only helps replacing one item (array item, map item or document) with multiple items of that type.
