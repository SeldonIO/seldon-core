# Install Seldon-Core

## Pre-requisites:
* Kubernetes cluster version equal or higher than 1.12
    * For Openshift it requires version 4.2 or higher
* Installer method
    * Helm version equal or higher than 3.0
    * Kustomize version equal or higher than 0.1.0

#### Running older versions of Seldon Core? 

Make sure you read the ["Upgrading Seldon Core Guide"](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html)

* **Seldon Core will stop supporting versions prior to 1.0 so make sure you upgrade.** 
* If you are running an older version of Seldon Core, and will be upgading it please make sure you read the [Upgrading Seldon Core docs]() to understand breaking changes and best practices for upgrading.
* Please see [Migrating from Helm v2 to Helm v3](https://helm.sh/blog/migrate-from-helm-v2-to-helm-v3/) if you are already running Seldon Core using Helm v2 and wish to upgrade.


## Install Seldon Core with Helm 

First [install Helm 3.x](https://docs.helm.sh). When helm is installed you can deploy the seldon controller to manage your Seldon Deployment graphs.

If you want to provide advanced parameters with your installation you can check the full [Seldon Core Helm Chart Reference](https://docs.seldon.io/projects/seldon-core/en/latest/reference/helm.html).

The namespace `seldon-system` is preferred, so we can create it:

```bash
kubectl create namespace seldon-system
```

Now we can install Seldon Core in the `seldon-system` namespace.

```bash
helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --namespace seldon-system
```

Make sure you install it with the relevant ingress (ambassador.enabled, istio.enabled, etc) so you are able to send requests (instructions below).

### Install with cert-manager

You can follow [the cert manager documentation to install it](https://docs.cert-manager.io/en/latest/getting-started/install/kubernetes.html)

You can then install seldon-core with:

```bash 
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --set certManager.enabled=true
```

### Ingress Support

For particular ingresses that we support, you can inform the controller it should activate processing for them.

 * Ambassador
   * add `--set ambassador.enabled=true` : The controller will add annotations to services it creates so Ambassador can pick them up and wire an endpoint for your deployments.
 * Istio Gateway
   * add `--set istio.enabled=true` : The controller will create virtual services and destination rules to wire up endpoints in your istio ingress gateway.

### Install an Ingress Gateway

We presently support two API Ingress Gateways

 * [Ambassador](https://www.getambassador.io/)
 * [Istio Ingress](https://istio.io/)

#### Install Ambassador

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

#### Install Istio Ingress Gateway

Follow [the istio docs](https://istio.io/) to install istio in your cluster. 

You must make sure you create a gateway - by default seldon-core will expect a gateway called seldon-gateway.

You can find full details on how to install Seldon Core with Istio (As well as how to create the gateway) in the [Istio Ingress Section](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html).


## Seldon Core Kustomize Install 

The [Kustomize](https://github.com/kubernetes-sigs/kustomize) installation can be found in the `/operator/config` folder of the repo. You should copy this template to your own kustomize location for editing.

To use the template directly there is a Makefile which has a set of useful commands:

For kubernetes clusters of version higher than 1.15, make sure you comment the patch_object_selector [here](https://github.com/SeldonIO/seldon-core/blob/master/operator/config/webhook/kustomization.yaml)

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

