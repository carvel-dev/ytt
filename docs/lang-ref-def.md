### Functions

- function definition

Labels returns map with two keys: test1, and test2:

```yaml
#@ def my_labels():
test1: 123
test2: 124
#@ end
```

Above is _almost_ equivalent to:

```yaml
#@ def my_labels():
#@   return {"test1": 123, "test2": 124}
#@ end
```

- common function usage

To set `labels` key to return value of `my_labels()`:

```yaml
labels: #@ my_labels()
```

To merge return value of `my_labels()` into `labels` map:

```yaml
#@ load("@ytt:template", "template")

labels:
  another-label: true
  _: #@ template.replace(my_labels())
```
