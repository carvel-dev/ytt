This example demonstrates how ytt handles specifying a directory for Data Values
inputs. (originally https://kubernetes.slack.com/archives/CH8KCCKA5/p1651167583289939)

With this:
```
├── config
│   ├── config.yml
│   └── values-schema.yml
└── values                   <-- any YAML under here is assumed to be plain Data Values
    ├── appdev-overrides     <-- sorted by full pathname, alphabetically
    │   └── values.yaml
    └── operator-overrides
        ├── 50-operations-overrides.yml
        ├── 99-opssec-overrides.yml
        └── approvals.toml   <-- non YAML files are ignored.
```

Executing this:

```console
$ ytt -f examples/data-values-directory/config.yml \
      --data-values-file examples/data-values-directory/values/
```