# Install Seldon-Core

**You will need a kubernetes cluster with version >=1.12**

To install seldon-core on a Kubernetes cluster you have several choices:

We presently support [Helm](#seldon-core-helm-install) and [Kustomize](#seldon-core-kustomize-install).

## Seldon Core Helm Install

First [install Helm](https://docs.helm.sh). When helm is installed you can deploy the seldon controller to manage your Seldon Deployment graphs.

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system
```

**For the unreleased 0.5.0 version you would need to install 0.5.0-SNAPSHOT to test**:

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --version 0.5.0-SNAPSHOT
```

Notes

 * You can use ```--namespace``` to install the seldon-core controller to a particular namespace but we recommend seldon-system.
 * For full configuration options see [here](../reference/helm.md)


## Install with cert-manager

You can follow [the cert manager documentation to install it](https://docs.cert-manager.io/en/latest/getting-started/install/kubernetes.html)

You can then install seldon-core with:

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --version 0.5.0-SNAPSHOT --set certManager.enabled=true
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

We suggest you install [the official helm chart](https://github.com/helm/charts/tree/master/stable/ambassador). At present we recommend 0.40.2 version due to issues with grpc in the latest.

```
helm install stable/ambassador --name ambassador --set crds.keep=false
```

### Install Istio Ingress Gateway

If you are using istio then the controller will create virtual services for an istio gateway. By default it will assume the gateway `seldon-gateway` as the name of the gateway. To change the default gateway add `--set istio.gateway=XYZ` when installing the seldon-core-operator.


## Seldon Core Kustomize Install 

The [Kustomize](https://github.com/kubernetes-sigs/kustomize) installation can be found in the `/operator/config` folder of the repo. You should copy this template to your own kustomize location for editing.

To use the template directly there is a Makefile which has a set of useful commands:


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

## Upgrading from Previous Versions

See our [upgrading notes](../reference/upgrading.md)
