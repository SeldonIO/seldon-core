# Openshift

## Working with RedHat Openshift Service Mesh

If you run with Openshift RedHat Service Mesh you can work with Seldon by following these steps.

### Create Gateway

Ensure you create a Gateway in istio-system. For 

```
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: seldon-gateway
  namespace: istio-system
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

### Activate Istio

1. Update the Seldon Core CSV to activate istio. Add:

```
  config:
    env:
    - name: ISTIO_ENABLED
      value: 'true'
```


### Namespace Seldon Core Install

If you install Seldon Core in a particular namespace you will need to:

 1. Add a NetworkPolicy to allow the webhooks to run. For the namespace yoy are running the operator create:

```
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: seldon-webhook
  namespace: <namespace>
spec:
  ingress:
  - ports:
    - port: 8443
      protocol: TCP
  podSelector:
    matchLabels:
      control-plane: seldon-controller-manager
  policyTypes:
  - Ingress
```


## Deleting Seldon Core Operator

At present webhook configuration is not cleaned up on delete of a Seldon Core Operator. You will need to delete the `MutatingWebhookConfiguration` and `ValidatingWebhookConfiguration`.

For namespace installs of Seldon Core these will be called:

 * `seldon-mutating-webhook-configuration-<namespace>`
 * `seldon-validating-webhook-configuration-<namespace>`