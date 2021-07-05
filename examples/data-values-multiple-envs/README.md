This example was created to show how one can configure multiple environments that have same set of base values (https://kubernetes.slack.com/archives/CH8KCCKA5/p1571430646113900).

`config/schema.yml` defines set of all data values that are used by configs, with defaults. Each file in `envs/` configures these data values specific to its environment.

```bash
ytt -f config/ -v version=123
ytt -f config/ -f envs/dev.yml
ytt -f config/ -f envs/staging.yml
ytt -f config/ -f envs/prod.yml
```
