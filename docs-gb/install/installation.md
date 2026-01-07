# Install On Kubernetes

## Pre-requisites

- Kubernetes cluster >= 1.23
- Installer method
  - Helm version >= 3.0
  - Kustomize version >= 0.1.0
- Ingress
  - Istio: we recommend >= 1.16
  - Ambassador v1 and v2

### Kubernetes Compatibility Matrix

Seldon Core 1.16 bumps minimum Kubernetes version to 1.23. This is because as part of making Seldon Core compatible with Kubernetes 1.25 we moved from autoscaling/v2beta1 apiVersion of HorizontalPodAutoscaler to autoscaling/v2 (see this [PR](https://github.com/SeldonIO/seldon-core/pull/4172) for further details).

Following table provides a summary of Seldon Core / Kubernetes version compatibility for recent version of Seldon Core.

| Core Version \ K8s Version | 1.23 | 1.24 | 1.25 | 1.26 | 1.27 | 1.28 | 1.29 | 1.30 | 1.31 | 1.32 | 1.33 | 1.34 | 1.35 |
|----------------------------|------|------|------|------|------|------|------|------|------|------|------|------|------|
| 1.16                       | ✓    | ✓    | ✓    | ✓    | ✓    |      |      |      |      |      |      |      |      |
| 1.17                       | ✓    | ✓    | ✓    | ✓    | ✓    |      |      |      |      |      |      |      |      |
| 1.18                       | ✓    | ✓    | ✓    | ✓    | ✓    |      |      |      |      |      |      |      |      |
| 1.19                       | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    | ✓    |

It is always recommended to first upgrade Seldon Core to the latest supported version on your Kubernetes cluster and then upgrade the Kubernetes cluster.

### Running older versions of Seldon Core?

Make sure you read the [Upgrading Seldon Core Guide](../upgrading.md) to understand breaking changes and best practices for upgrading.

- **Seldon Core will stop supporting versions prior to 1.0 so make sure you upgrade.**

# Install Seldon Core with Helm

You will first need to add the seldonio Helm repo:
```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

And then install the `seldon-core-operator` chart:
```bash
kubectl create namespace seldon-system
helm install seldon-core-operator seldonio/seldon-core-operator --namespace seldon-system
```

To install a specific version of the chart, you can list the available versions:
```bash
helm search repo seldon/seldon-core-operator --versions
```

And then either install it from scratch:
```bash
helm install seldon-core-operator seldonio/seldon-core-operator \
  --namespace seldon-system \
  --version 1.17.1
```
Or, upgrade your existing installation:
```bash
helm upgrade --install seldon-core-operator seldon/seldon-core-operator \
  --version 1.17.1 \
  --namespace seldon-system
```

To install it from a locally downloaded chart(GitHub Release):
1. Download the chart - https://github.com/SeldonIO/helm-charts/releases/tag/seldon-core-operator-1.18.2
2. Install it the archive from the folder you downloaded it in:
```bash
 helm install seldon-core-operator \
  ./seldon-core-operator-1.18.2.tgz \
  --namespace seldon-system
```

To install it from our GCS bucket:
```bash
helm install seldon-core-operator \
  https://storage.googleapis.com/seldon-charts/seldon-core-operator-1.17.1.tgz \
  --namespace seldon-system
```

If you want to provide advanced parameters with your installation you can check the full [Seldon Core Helm Chart Reference](./advanced-helm-chart-configuration.md).
For example, if you want to install the operator with Istio or Ambassador enabled, and usage metrics enabled you can 
{% tabs %}

{% tab title="Istio" %}

```bash
helm install seldon-core-operator seldonio/seldon-core-operator \
    --set usageMetrics.enabled=true \
    --set istio.enabled=true \
    --namespace seldon-system
```

{% endtab %}

{% tab title="Ambassador" %}

```bash
helm install seldon-core-operator seldonio/seldon-core-operator \
    --set usageMetrics.enabled=true \
    --set ambassador.enabled=true \
    --namespace seldon-system
```

{% endtab %}

{% endtabs %}

You can check that the operator is running:
```bash
kubectl get pods -n seldon-system
```

You should see a `seldon-controller-manager` pod with `STATUS=Running`.


For full instructions on installation with Istio and Ambassador read the following pages:

- [Ingress with Istio](../routing/istio.md)
- [Ingress with Ambassador](../routing/ambassador.md)

### Install a SNAPSHOT version

Whenever a new PR was merged to master, we have set up our CI to build a "SNAPSHOT" version, which would contain the Docker images for that specific development / master-branch code. Whilst the images are pushed under SNAPSHOT, they also create a new "dated" SNAPSHOT version entry, which pushes images with the tag `<next-version>-SNAPSHOT_<timestamp>`. A new branch is also created with the name `v<next-version>-SNAPSHOT_<timestamp>`, which contains the respective helm charts, and allows for the specific version (as outlined by the version in `version.txt`) to be installed.

This means that you can try out a dev version of master if you want to try a specific feature before it's released.

For this you would be able to clone the repository, and then checkout the relevant SNAPSHOT branch.

Once you have done that you can install the `seldon-core-operator`, as described above in the page.

### Install with cert-manager

You can follow [the cert manager documentation to install it](https://cert-manager.io/docs/installation/).

You can then install seldon-core with:

```bash
helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --namespace seldon-system \
    --set certManager.enabled=true
```

# Seldon Core Kustomize Install

The [Kustomize](https://github.com/kubernetes-sigs/kustomize) installation can be found in the `/operator/config` folder of the repo. You should copy this template to your own kustomize location for editing.

To use the template directly, there is a Makefile which has a set of useful commands:

For kubernetes clusters of version higher than 1.15, make sure you [comment the patch_object_selector here](https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml#L8).

Install cert-manager

```bash
make install-cert-manager
```

Install Seldon using cert-manager to provide certificates.

```bash
make deploy
```

Install Seldon with provided certificates in `config/cert/`

```bash
make deploy-cert
```

# Other Options

## Install Production Integrations

Now that you have Seldon Core installed, you can set it up with:

### Install with Kubeflow

- [Install Seldon as part of Kubeflow.](https://www.kubeflow.org/docs/external-add-ons/serving/seldon/)

### GCP MarketPlace

If you have a Google Cloud Platform account you can install via the [GCP Marketplace](https://console.cloud.google.com/marketplace/details/seldon-portal/seldon-core).

### OpenShift

You can install Seldon Core via OperatorHub on the OpenShift console UI.

### OperatorHub

You can install Seldon Core from [Operator Hub](https://operatorhub.io/operator/seldon-operator).

# Upgrading from Previous Versions

See our [upgrading notes](../upgrading.md)

# Advanced Usage

## Install Seldon Core in a single namespace (version >=1.0)

**You will need a k8s cluster >= 1.15**

### Helm

You can install the Seldon Core Operator so it only manages resources in its namespace. An example to install in a namespace `seldon-ns1` is shown below:

```bash
kubectl create namespace seldon-ns1
kubectl label namespace seldon-ns1 seldon.io/controller-id=seldon-ns1
```

We label the namespace with `seldon.io/controller-id=<namespace>` to ensure if there is a clusterwide Seldon Core Operator that it should ignore resources for this namespace.

Install the Operator into the namespace:

```bash
helm install seldon-namespaced seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set singleNamespace=true \
    --set image.pullPolicy=IfNotPresent \
    --set usageMetrics.enabled=false \
    --set crd.create=true \
    --namespace seldon-ns1
```

We set `crd.create=true` to create the CRD. If you are installing a Seldon Core Operator after you have installed a previous Seldon Core Operator on the same cluster you will need to set `crd.create=false`.

### Kustomize

An example install is provided in the Makefile in the Operator folder:

```bash
make deploy-namespaced1
```

See the [multiple server example notebook](../examples/multiple_operators.html).

## Label focused Seldon Core Operator (version >=1.0)

**You will need a k8s cluster >= 1.15**

You can install the Seldon Core Operator so it manages only SeldonDeployments with the label `seldon.io/controller-id` where the value of the label matches the controller-id of the running operator. An example for a namespace `seldon-id1` is shown below:

### Helm

```bash
kubectl create namespace seldon-id1
```

To install the Operator run:

```bash
helm install seldon-controllerid seldon-core-operator  --repo https://storage.googleapis.com/seldon-charts  \
    --set singleNamespace=false \
    --set image.pullPolicy=IfNotPresent \
    --set usageMetrics.enabled=false \
    --set crd.create=true \
    --set controllerId=seldon-id1 \
    --namespace seldon-id1
```

We set `crd.create=true` to create the CRD. If you are installing a Seldon Core Operator after you have installed a previous Seldon Core Operator on the same cluster you will need to set `crd.create=false`.

For kustomize you will need to [uncomment the patch_object_selector here](https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml)

### Kustomize

An example install is provided in the Makefile in the Operator folder:

```bash
make deploy-controllerid
```

See the [multiple server example notebook](../notebooks/multiple_operators.md).

## Install behind a proxy

When your kubernetes cluster is behind a proxy, the `kube-apiserver` typically inherits the system proxy variables. This can block the `kube-apiserver` from reaching the webhooks needed to create Seldon resources.

You could see this error:

```bash
Internal error occurred: failed calling webhook "v1.vseldondeployment.kb.io": Post https://seldon-webhook-service.seldon-system.svc:443/validate-machinelearning-seldon-io-v1-seldondeployment?timeout=30s: Service Unavailable
```

To fix this, ensure the `no_proxy` environment variable for the `kube-apiserver` includes `.svc,.svc.cluster.local`. See [this Github Issue Comment](https://github.com/cert-manager/cert-manager/issues/2640#issuecomment-601872165) for reference. As described there, the error could also occur for the `cert-manager-webhook`.
