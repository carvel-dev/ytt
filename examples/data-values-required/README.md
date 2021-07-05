This example was created to show various patterns to validate data values (https://kubernetes.slack.com/archives/CH8KCCKA5/p1571366872072800).

In following examples, we validate that `version` is not null, and port is > 0.

User is forced to provide `version` value via cmd line or additional data values file because default `schema.yml` specifies it as `@schema/nullable` and configuration validates it to be non-null.

Inline approach shows how to have short validations in places where data.values... is used:

```bash
ytt -f inline/ -v version=123
```

Function approach shows how to split out data values and their validations into their own functions:

```bash
ytt -f function/ -v version=123
```

Bulk approach shows how to validate data values in bulk, and then expose them for later use:

```bash
ytt -f bulk/ -v version=123
```
