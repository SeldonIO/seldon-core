# Install Seldon-Core

**You will need a kubernetes cluster with version >=1.12**

To install seldon-core on a Kubernetes cluster you have several choices:

We presently support [Helm](#seldon-core-helm-install) and [Kustomize](#seldon-core-kustomize-install).

>**Note:** From Seldon Core v1.0.0 onward, the minimum supported version of Helm is v3.0.0. Users still running Helm v2.16.1 or below should upgrade to a supported version before installing Seldon Core.

>Please see [Migrating from Helm v2 to Helm v3](https://helm.sh/blog/migrate-from-helm-v2-to-helm-v3/) if you are already running Seldon Core using Helm v2 and wish to upgrade.

## Seldon Core Helm Install

First [install Helm](https://docs.helm.sh). When helm is installed you can deploy the seldon controller to manage your Seldon Deployment graphs.

```bash
kubectl create namespace seldon-system
```

```bash
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system
```

Notes

 * You can use ```--namespace``` to install the seldon-core controller to a particular namespace but we recommend seldon-system.
 * For full configuration options see [here](../reference/helm.md)


## Install with cert-manager

You can follow [the cert manager documentation to install it](https://docs.cert-manager.io/en/latest/getting-started/install/kubernetes.html)

You can then install seldon-core with:

```bash 
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --version 0.5.0-SNAPSHOT --set certManager.enabled=true
```

## Ingress Support

For particular ingresses that we support, you can inform the controller it should activate processing for them.

 * Ambassador
   * add `--set ambassador.enabled=true` : The controller will add annotations to services it creates so Ambassador can pick them up and wire an endpoint for your deployments.
 * Istio Gateway
   * add `--set istio.enabled=true` : The controller will create virtual services and destination rules to wire up endpoints in your istio ingress gateway.

## Install an Ingress Gateway

We presently support two API Ingress Gateways

 * [Ambassador](https://www.getambassador.io/)
 * [Istio Ingress](https://istio.io/)

### Install Ambassador

We suggest you install [the official helm chart](https://github.com/helm/charts/tree/master/stable/ambassador).


```bash
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
```

```bash
helm repo update
```

```bash
helm install ambassador stable/ambassador --set crds.keep=false
```

### Install Istio Ingress Gateway

Follow [the istio docs](https://istio.io/) to install. 

If you are using istio then the controller will create virtual services for an istio gateway. By default it will assume the gateway `seldon-gateway` as the name of the gateway. To change the default gateway add `--set istio.gateway=XYZ` when installing the seldon-core-operator.


## Seldon Core Kustomize Install 

The [Kustomize](https://github.com/kubernetes-sigs/kustomize) installation can be found in the `/operator/config` folder of the repo. You should copy this template to your own kustomize location for editing.

To use the template directly there is a Makefile which has a set of useful commands:

For kubernetes <1.15 comment the patch_object_selector [here](https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml)

Install cert-manager

```
make install-cert-manager
```

Install Seldon using cert-manager to provide certificates.

```
make deploy
```

Install Seldon with provided certificates in `config/cert/`

```
make deploy-cert
```


## Other Options

### Install with Kubeflow

  * [Install Seldon as part of Kubeflow.](https://www.kubeflow.org/docs/guides/components/seldon/#seldon-serving)

### GCP MarketPlace

If you have a Google Cloud Platform account you can install via the [GCP Marketplace](https://console.cloud.google.com/marketplace/details/seldon-portal/seldon-core).

### AWS MarketPlace

If you have a AWS account you can install via the [AWS Marketplace](https://aws.amazon.com/marketplace/pp/B07KCNBCHV). See our [AWS Install Documentation](../reference/aws-mp-install.md).


## Upgrading from Previous Versions

See our [upgrading notes](../reference/upgrading.md)

## Advanced Usage

### Install Seldon Core in a single namespace (version >=1.0)

**You will need a k8s cluster >= 1.15**

#### Helm

You can install the Seldon Core Operator so it only manages resources in its namespace. An example to install in a namespace `seldon-ns1` is shown below:

```bash
kubectl create namespace seldon-ns1
kubectl label namespace seldon-ns1 seldon.io/controller-id=seldon-ns1
```

We label the namespace with `seldon.io/controller-id=<namespace>` to ensure if there is a clusterwide Seldon Core Operator that it should ignore resources for this namespace.

Install the Operator into the namespace:

```bash
helm install seldon-namespaced seldon-core-operator  --repo https://storage.googleapis.com/seldon-charts  \
    --set singleNamespace=true \
    --set image.pullPolicy=IfNotPresent \
    --set usageMetrics.enabled=false \
    --set crd.create=true \
    --namespace seldon-ns1
```

We set `crd.create=true` to create the CRD. If you are installing a Seldon Core Operator after you have installed a previous Seldon Core Operator on the same cluster you will need to set `crd.create=false`.


#### Kustomize

An example install is provided in the Makefile in the Operator folder:

```
make deploy-namespaced1
```


See the [multiple server example notebook](../examples/multiple_operators.html).

### Label focused Seldon Core Operator (version >=1.0)

**You will need a k8s cluster >= 1.15**

You can install the Seldon Core Operator so it manages only SeldonDeployments with the label `seldon.io/controller-id` where the value of the label matches the controller-id of the running operator. An example for a namespace `seldon-id1` is shown below:


#### Helm

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

For kustomize you will need to uncomment the patch_object_selector [here](https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml)

#### Kustomize

An example install is provided in the Makefile in the Operator folder:

```
make deploy-controllerid
```

See the [multiple server example notebook](../examples/multiple_operators.html).

