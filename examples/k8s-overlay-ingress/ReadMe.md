This example shows how to use `ytt` overlaying in regards to accomplish the modification of X Kubernetes Ingress object.

Specifically the `apiVersion` is bumped to **networking.k8s.io/v1** from **networking.k8s.io/v1beta1**. This is done in order to get the Falco installation Helm post-rendered into a state that matches the version of the `networking` API on a specific `Kubernetes` version.
As a consequence of that we're required to `ytt overlay` **backend** and **paths** on the **Kubernetes Ingress Object**. In order for these to match the `apiVersion` that is overlayed to.

All of this stems from the Falco Helm chart having issues with the version tag that K3s uses. It's e.g `v1.22.3+k3s2`. This results in the **Falco** chart using the **v1beta1** of the **networking** API instead of the newest version of the **networking** API that we want.

The cmdline used:

```text
        ytt -f "./config.yaml" -f "./schema.yml" --data-value cluster="my-cluster" \
        | helm upgrade --atomic --install "HELM_INSTALL_NAME" "HELM_CHART_NAME" --version "HELM_VERSION" --create-namespace --namespace "KUBERNETES_NAMESPACE" --values - \
        --post-renderer "./ytt-helm-postrender/ytt-overlay-on-helm-post-renderer.sh"
```

Here's what's going on:

- The cmd uses ytt's data values schema to perform type-checking
- it uses those values as input to a helm chart
- and finally it uses ytt to patch the result of the helm templates

> N.B. the Falco chart can, if enabled, request two **Kubernetes Ingress Objects** to be created. This is done in this example and therefore we need to overlay on both. Specificall notice the difference on the `overlay.subset()` call, filtering on `{"metadata": {"name": ....` (together with the `overlay.subset({"kind": "Ingress"})` matcher).

_The `--post-renderer` script being called is simply an extra call to `ytt` and the `schema.yaml` and `network-api-fix.yaml` files specifically. So that `ytt` renders these two files against the `Helm` generated output ( the input to `ytt` ).


