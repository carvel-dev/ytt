This example showcases how to wrap an upstream set of templates to expose a simplified set of data values.

In this example:

- `_ytt_lib/app1`: contains app1 templates
  - `schema.yml`: declares data values that configure image (e.g. image.url)
- `schema.yml`: declares simplified data value to configure part of image (e.g. image.username)
- `config.yml`: includes app1 configuration with augmented image data value

```bash
$ ytt -f .

vals:
  image:
    url: registry.com//repo
  other: config
```

```bash
$ ytt -f . -v image.username=bob

vals:
  image:
    url: registry.com/bob/repo
  other: config
```

As of ytt v0.28.0+, you can directly configure library data values as well:

```bash
$ ytt -f . -v image.username=bob -v @app1:other=new-config

vals:
  image:
    url: registry.com/bob/repo
  other: new-config
```

or with this additional data values file:

```bash
$ cat /tmp/custom.yml

#@data/values
---
image:
  username: bob

#@data/values
#@library/ref "@app1"
---
other: new-config
```

```bash
$ ytt -f . -f /tmp/custom.yml

vals:
  image:
    url: registry.com/bob/repo
  other: new-config
```
