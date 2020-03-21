Helm 3.1 introduced a way to make modifications to templated resources before installing them via a pluggable interface. Helm will call a binary with rendered template contents over stdin and expect results over stdout. It's trivial to "plug" ytt into this model to apply additional changes. This directory showcases such setup.

Directory contents:

- `ytt-post-renderer`: bash script that calls ytt to apply files in `config/` directory
- `config/`
  - `add-label.yml`: includes overlay that adds new label to all resources (including ones specified in `additional-resources.yml`)
  - `additional-resources.yml`: includes two ConfigMaps to show how to add resources

From within this directory:

```bash
helm repo add stable https://kubernetes-charts.storage.googleapis.com
helm pull --untar stable/postgresql
helm template postgresql postgresql/ --post-renderer ./ytt-post-renderer
                                     # ^ same flag goes for install/upgrade
```
