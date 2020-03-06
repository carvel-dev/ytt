### ytt Library: Version module

Available in v0.26.0+

Version module provides a way to assert on minimum ytt binary version in your configuration. It could be placed into a conditional or just at the top level within `*.star`, `*.yml` or any other template file.

Example configuration directory may look like this:

- `config/`
  - `0-min-version.star`: contents below
  - `deployment.yml`
  - `other.yml`

```python
# filename starts with '0-' to make sure this file gets
# processed first, consequently forcing version check run first
load("@ytt:version", "version")

version.require_at_least("0.26.0")
```

Note that ytt sorts files alphanumerically and executes templates in order. Most of the time it's best to check version first before processing other templates, hence, in the above example we've named file `0-min-version.star` so that it's first alphanumerically.
