This example was created to show how to change helm chart configuration without having to fork chart itself (https://kubernetes.slack.com/archives/C0NH30761/p1561061403145300).

```bash
helm template ... | ytt -f- -f rbac-fix.yml | kapp deploy -a app1 -f- -y
```
