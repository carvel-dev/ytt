This example shows how to use `ytt` overlaying in regards to accomplish the modification of X Kubernetes Ingress object.

Specifically the `apiVersion` is bumped to **networking.k8s.io/v1** from **networking.k8s.io/v1beta1**. As a consequence this requires us to
`ytt overlay` on **backend** and **paths** on the **Kubernetes Ingress Object**. In order for these to match the `apiVersion` that we bump to.

Specifically the method is used on Falco Helm chart having issues with the version tag that K3s uses. Resulting in that chart using the **v1beta1** of the **networking** API.

The cmdline used:

```text
        ytt -f "./config.yaml" -f "./schema.yml" --data-value cluster="my-cluster" \
        | helm upgrade --atomic --install "HELM_INSTALL_NAME" "HELM_CHART_NAME" --version "HELM_VERSION" --create-namespace --namespace "KUBERNETES_NAMESPACE" --values - \
        --post-renderer "./ytt-helm-postrender/ytt-overlay-on-helm-post-renderer.sh"
```

> N.B. the Falco chart can, if enabled, request two **Kubernetes Ingress Objects** to be created. This is done in this example and therefore we need to overlay on both. Specificall notice the difference on the `overlay.subset()` call, filtering on `{"metadata": {"name": ....` (together with the `overlay.subset({"kind": "Ingress"})` matcher).
