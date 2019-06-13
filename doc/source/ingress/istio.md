# Istio

Seldon core can be used in conjuction with [istio](https://istio.io/). Istio provides an [ingress gateway](https://istio.io/docs/tasks/traffic-management/ingress/) which Seldon Core can automatically wire up new deployments to. The steps to using istio are described below.

## Install Seldon Core Operator

Ensure when you install the seldon-core operator via Helm that you enabled istio. For example:

```bash 
helm install seldon-core-operator --name seldon-core --set istio.enabled=true --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

You need an istio gateway installed. By default we assume one called seldon-gateway. For example you can create this with the following yaml:

```
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: seldon-gateway
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
```

If you have your own gateway you will use then you can provide the name when installing the seldon operator. For example if your gateway is called `mygateway` you can install the operator with:

```bash 
helm install seldon-core-operator --name seldon-core --set istio.enabled=true --set istio.gateway=mygateway --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

You can also provide the gateway on a per Seldon Deployment resource basis by providing it with the annotation `seldon.io/istio-gateway`.

## Traffic Routing

Istio has the capability for fine grained traffic routing to your deployments. This allows:

 * canary updates
 * green-blue deployments
 * A/B testing

An example showing canary updates can be found [here](../examples/istio_canary.html)



