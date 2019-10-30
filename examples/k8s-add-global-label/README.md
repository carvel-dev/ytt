This example was created to show how to add a label to all resources (using overlays) with its value injected from outside (https://kubernetes.slack.com/archives/C9A5ALABG/p1572421415180100).

```bash
ytt -f .                  # uses defaults
ytt -f . -v build_num=123 # overrides default
```
