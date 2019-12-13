# Install Seldon Core on EKS via AWS MarketPlace

 * Subscribe to Seldon Core on [AWS MarketPlace](https://aws.amazon.com/marketplace/seller-profile?id=cec67450-7a7e-43d5-8e5f-61e94e7c9e03&ref=dtl_B07KCNBCHV) and retrieve the log in command to authenticate your Docker client.

 (Note: below is for the 0.5.0 release and will differ for other releases).
  ```
  $(aws ecr get-login --no-include-email --region us-east-1 --registry-ids 403495124976)
  ```

 * [Create your EKS cluster and authenticate kubectl](https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html).
   Configure local Kubectl
  ```
  aws eks --region <CLUSTER_REGION> update-kubeconfig --name <CLUSTER-NAME>
  ```

 * Install [helm](https://docs.helm.sh/) on your cluster if it is not there already. **Note: Helm v3 is used for the following set up.**  

  Create a namespace for the Seldon system.

  ```
  kubectl create namespace seldon-system
  ```

 * Install Seldon Core for the release you subscribed to on Amazon MarketPlace:

For **Seldon 0.5.0**

 ```
  helm install seldon-core seldon-core-aws --repo https://storage.googleapis.com/seldon-aws-charts --version 0.5.0 --set usageMetrics.enabled=true --namespace seldon-system
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

If you are using istio then the controller will create virtual services for an istio gateway. By default it will assume the gateway `seldon-gateway` as the name of the gateway. To change the default gateway add `--set istio.gateway=XYZ` when installing the seldon-core-operator.


## Next Steps

For next steps on using Seldon Core and deploying your first ML models visit the [Seldon Core project page](https://github.com/SeldonIO/seldon-core).

