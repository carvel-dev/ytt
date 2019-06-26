## Injecting Secrets

**This document is work in progress.**

Unlike most configuration, many organizations disallow storing of plain secret values next to other code/configuration.

This document:
- presents several common approaches used to make secrets available to your templates
- does _not_ cover injection of secrets directly into an application at runtime (overall may be the best approach)
- does _not_ recommend one approach over another (though it does state pros and cons)
- does _not_ talk about where resulting templates are forwarded

One common question that's asked is why not to extend ytt to allow it to shell out to other programs or why not include builtin library that can fetch secrets from outside (e.g. imagine having builtin `vault` function). We are _not_ planning to support either of these features because we want to maintain strong sandbox boundary between templating language and ytt's external environment. Our expectation is that if ytt user wants to provide specific data to templates, it should be injected into ytt rather than templates (or ytt itself) fetching it. We believe this is a good security posture and as a side effect makes templates more portable (some users may store secrets in Vault and some other in Lastpass).

There might be other ways to inject secrets into ytt that are not listed here.

### Via command line args

```bash
ytt -f . --data-value username=user --data-value password=pass
```

Cons:

- depending on a computing environment secret values are visible in a process tree (as part of full command line) **which typically is considered unwanted**

### Via environment variables

```bash
export CFG_secrets__username=user
export CFG_secrets__password=pass
ytt -f . --data-values-env CFG
```

Cons:

- depending on a computing environment secret values are visible in process metadata (e.g. typically accessible to root user under `/proc/<pid>/environ`)

### Via secret references, before templating

Given following two files:

`config.yml`:

```yaml
#@ load("@ytt:data", "data")

users:
- username: #@ data.values.username
  password: #@ data.values.password
```

`secrets`:

```yaml
username: (( vault "secret/my/credentials/admin:username" ))
password: (( vault "secret/my/credentials/admin:password" ))
```

(Note that `secrets` file does not have `yaml` extension to make ytt exclude it from template result. Alternatively `--file-mark` flag can be used to exclude that file.)

One can compose ytt with external tooling that knows how to resolve referenced values. For example, [spruce](https://starkandwayne.com/blog/simple-secure-credentials-into-yaml-with-vault-and-spruce/) can resolve secret references and then forward that information onto ytt:

```bash
echo -e "#@data/values\n---\n$(spruce merge ./secrets)" | ytt -f . -f -
```

Pros:

- able to commit mapping of data values to secrets in your preferred secret storage (in this example Vault)
- secrets can be inserted in the middle of a string or base64 encoded (as sometimes required by k8s)

### Via secret references, after templating

Given following two files:

`config.yml`:

```yaml
#@ load("@ytt:data", "data")

users:
- username: #@ data.values.username
  password: #@ data.values.password
```

`values.yml`:

```yaml
#@data/values
---
username: (( vault "secret/my/credentials/admin:username" ))
password: (( vault "secret/my/credentials/admin:password" ))
```

One can compose ytt with external tooling that knows how to resolve referenced values. Such tools typically look for a particular patterned map or array items. One such tool is [spruce](https://starkandwayne.com/blog/simple-secure-credentials-into-yaml-with-vault-and-spruce/):

```bash
ytt -f . | spruce merge -
```

Pros:

- provides a way to commit mapping of data values and their associated secret locations (e.g. `username` comes from `vault` under `secret/my/credentials/admin:username`)

Cons:

- secret resolution tools may not be able to find secret references inside strings (e.g. `http://(( ... )):(( ... ))@website`)
- secret resolution tools will not understand if secret reference was encoded (for example with `base64` encoding as wanted sometimes by k8s)
