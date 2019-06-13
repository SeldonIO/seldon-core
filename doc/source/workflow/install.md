# Install Seldon-Core

To install seldon-core on a Kubernetes cluster you have several choices:

 * If you have a Google Cloud Platform account you can install via the [GCP Marketplace](https://console.cloud.google.com/marketplace/details/seldon-portal/seldon-core).

For CLI installs:

We presently support Helm installs.

## Install Seldon Core

First [install Helm](https://docs.helm.sh). When helm is installed you can deploy the seldon controller to manage your Seldon Deployment graphs.

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

Notes

 * You can use ```--namespace``` to install the seldon-core controller to a particular namespace
 * For full configuration options see [here](../reference/helm.md)

For particular ingresses we support you can inform the controller it should activate processing for them.

 * Ambassador
   * add `--set ambassador.enabled=true` : The controller will add annotations to services it creates so Ambassador can pick them up and wire an endpoint for your deployments.
 * Istio Gateway
   * add `--set istio.enabled=true` : The controller will create virtual services and destination rules to wire up endpoints in your istio ingress gateway.

## Install an Ingress Gateway

We presently support two API Ingress Gateways

 * [Ambassador](https://www.getambassador.io/)
 * [Istio Ingress](https://istio.io/)
 * Seldon Core OAuth Gateway

### Install Ambassador

We suggest you install [the official helm chart](https://github.com/helm/charts/tree/master/stable/ambassador). At present we recommend 0.40.2 version due to issues with grpc in the latest.

```
helm install stable/ambassador --name ambassador --set crds.keep=false
```

### Install Istio Ingress Gateway

If you are using istio then the controller will create virtual services for an istio gateway. By default it will assume the gateway `seldon-gateway` as the name of the gateway. To change the default gateway add `--set istio.gateway=XYZ` when installing the seldon-core-operator.

### Install Seldon OAuth Gateway

This provides a basic OAuth Gateway.

```bash
helm install seldon-core-oauth-gateway --name seldon-gateway --repo https://storage.googleapis.com/seldon-charts
```

## Other Options

### Install with Kubeflow

  * [Install Seldon as part of Kubeflow.](https://www.kubeflow.org/docs/guides/components/seldon/#seldon-serving)


## Upgrading from Previous Versions

See our [upgrading notes](../reference/upgrading.md)

