# Data Values vs Overlays

As folks get started with `ytt`, a common question that arises is, “when should
I use data values versus overlays?” While these features do address a similar
problem space, we recommend using one feature versus the other depending on the
use case. We will detail our guidance below.

## Data Values

[Data values](ytt-data-values.md) provide a way to inject input data into a
template. If you think about a ytt template as a function, then data values are
the varying parameters. The configuration author will expose values that are
likely to change often, such as with every new environment, as data values.
Authors can also set data values to reasonable defaults or leave them empty and
require the consumers to input their own.

Common use cases:

1. Set default values (for generic or specific use-cases)
1. Provide visibility to values that could be updated by the consumer
1. Enable simple conditional behavior (such as requiring a value to be set)

Consider this simple example provided by the configuration author:

`config.yml`
```yaml
#@ load("@ytt:data", "data")
#@ load("@ytt:assert", "assert")
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: #@ data.values.deployment_name or assert.fail("missing deployment_name")
  labels:
    app: nginx
spec:
  replicas: #@ data.values.replicas
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: #@ data.values.nginx_image
        ports:
        - containerPort: 80
```

`values.yml`
```yaml
#@data/values
---
deployment_name: ""
replicas: 1
nginx_image: nginx:1.14.2
```

The configuration consumer can provide their own data values files that override the ones
provided.

This example portrays how to:

1. Set default values. The configuration author has set appropriate default
   values for `replicas` and `nginx_image`.
1. Provide visibility to values that could be updated by the consumer.
   `deployment_name`, `replicas`, and `nginx_image` are all configurable by the
   consumer.
1. Enable simple conditional behavior. `deployment_name` must be set by the
   consumer, otherwise an error will be returned.

Pros:

- Simple (easy to use, easy to understand)

Cons:

- Limited in capability (by design)

---
## Overlays

When consumers would like to configure fields beyond what the original
author has exposed as data values, they should turn to
[Overlays](lang-ref-ytt-overlay.md). These documents provide a way to specify
locations within configuration and either add to, remove from, or replace within
that existing configuration.  With basic usage, overlays can act as an extension
of data values, but as situations inevitably become more complex, overlays
provide many more capabilities.

Common use cases:
1. To replace, delete, or append configuration
1. Situation-specific values

Consider this extension of our previous example:

`config.yml`
```yaml
#@ load("@ytt:data", "data")
#@ load("@ytt:assert", "assert")
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: #@ data.values.deployment_name or assert.fail("missing deployment_name")
  labels:
    app: nginx
spec:
  replicas: #@ data.values.replicas
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: #@ data.values.nginx_image
        ports:
        - containerPort: 80
      - name: to-be-removed
        image: image:1.2.3
        ports:
        - containerPort: 80
```

`values.yml`
```yaml
#@data/values
---
deployment_name: ""
replicas: 1
nginx_image: nginx:1.14.2
```

`add-namespace.yml`
```yaml
#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.all, expects="1+"
---
metadata:
  #@overlay/match missing_ok=True
  namespace: my-namespace
```

`remove-container.yml`
```yaml
#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.all, expects="1+"
---
spec:
  template:
    spec:
      containers:
        #@overlay/match by="name"
        #@overlay/remove
        - name: to-be-removed
```

`add-container.yml`
```yaml
#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.all, expects="1+"
---
spec:
  template:
    spec:
      containers:
        #@overlay/append
        - name: appended-container
          image: image:1.2.3
          ports:
          - containerPort: #@ data.values.appended_container_port
```

`prod-values.yml`
```yaml
#@data/values
---
deployment_name: 3
replicas: 3
#@overlay/match missing_ok=True
appended_container_port: 8080
```

This example demonstrates a number of overlay capabilities:

1. Appending configuration via `add-namespace.yml` and `add-container.yml`
1. Removing configuration via `remove-container.yml`
1. Extending and updating data values via `prod-values.yml`

Use the [ytt playground to play with this
example](https://get-ytt.io/#gist:https://gist.github.com/aaronshurley/b6868b76e25fcb24aedde42f522734af).

Pros:
- Has more functionality
- Doesn’t change the source
- Programmatic capabilities (this wasn't demonstrated here, see
  [example](lang-ref-ytt-overlay.md#programmatic-access))

Cons:
- Added complexity

# If you want to learn more...

Check out the [Getting Started
tutorial](https://get-ytt.io/#example:example-hello-world) on the ytt website
for a detailed introduction to ytt.
