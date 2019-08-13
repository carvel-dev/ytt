## FAQ

### Requiring data values to be provided

> is there a way i can fail in ytt for a missing data value from --data-values-env? for example, i never want to have a MySQL password in a data/values document. but it is always required to be passed as an environment string or the deploy should fail

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565644534231700)

ytt library's `assert` package is useful for such situations. It can be used like so:

```yaml
password: #@ data.values.env.mysql_password if data.values.env.mysql_password else assert.fail("missing env.mysql_password")
```

or even more compactly,

```yaml
password: #@ data.values.env.mysql_password or assert.fail("missing env.mysql_password")
```

Note that empty strings are falsy in Starlark.
