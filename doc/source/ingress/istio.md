# Ingress with Istio

Seldon Core can be used in conjunction with [istio](https://istio.io/). Istio provides an [ingress gateway](https://istio.io/docs/tasks/traffic-management/ingress/) which Seldon Core can automatically wire up new deployments to. The steps to using istio are described below.

## Install Seldon Core Operator

Ensure when you install the seldon-core operator via Helm that you enabled istio. For example:

```bash 
helm install seldon-core seldon-core-operator --set istio.enabled=true --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

You need an istio gateway installed in the `istio-system` namespace. By default we assume one called seldon-gateway. For example you can create this with the following yaml:

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

If you want to want to create SSL based gateway, create your signed certificate or actual signed certificate (for example named fullchain.pem), key (privkey.pem) and then run follwing commands to get SSL gateway. Assuming we're not using [cert-manager](https://istio.io/latest/docs/ops/integrations/certmanager/) then create self-signed certificate with


```bash
openssl req -nodes -x509 -newkey rsa:4096 -keyout privkey.pem -out fullchain.pem -days 365 -subj "/C=GB/ST=GreaterLondon/L=London/O=SeldonSerra/OU=MLOps/CN=localhost"
```

Import certificate and key as a secret into istio-system namespace

```bash
kubectl create -n istio-system secret tls seldon-ssl-cert --key=privkey.pem --cert=fullchain.pem
```

and create SSL Istio Gateway using following YAML file

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
  - hosts:
    - '*'
    port:
      name: https
      number: 443
      protocol: HTTPS
    tls:
      credentialName: seldon-ssl-cert
      mode: SIMPLE
```


If you have your own gateway you will use then you can provide the name when installing the seldon operator. For example if your gateway is called `mygateway` you can install the operator with:

```bash 
helm install seldon-core seldon-core-operator --set istio.enabled=true --set istio.gateway=mygateway --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

You can also provide the gateway on a per Seldon Deployment resource basis by providing it with the annotation `seldon.io/istio-gateway`.

## Istio Configuration Annotation Reference

| Annotation | Default |Description |
|------------|---------|------------|
|`seldon.io/istio-gateway:<gateway name>`| istio-system/seldon-gateway | The gateway to use for this deployment. If no namespace prefix is applied it will refer to the namespace of the Seldon Deployment. |
| `seldon.io/istio-retries` | None | The number of istio retries |
| `seldon.io/istio-retries-timeout` | None | The per try timeout if istio retries is set |
| `seldon.io/istio-host` | `*` | The Host for istio Virtual Service |

All annotations should be placed in `spec.annotations` or `metadata.annotations`. `spec.annotations` will take precedence.


## Traffic Routing

Istio has the capability for fine grained traffic routing to your deployments. This allows:

 * canary updates
 * green-blue deployments
 * A/B testing
 * shadow deployments

More information can be found in our [examples](../examples/istio_examples.html), including [canary updates](../examples/istio_canary.html).

## Configuring Authentication/Authorization
To force clients to authenticate/authorize themselves in order to access the seldon model deployments, you can leverage Istio's 
`RequestAuthentication` and `AuthorizationPolicy`. This will deny or accept requests to the model depending on specified conditions that you designated in the policies. 
More information can be found [here](https://istio.io/latest/docs/reference/config/security/authorization-policy/).

You can set the policies to target all the models belonging to a specific namespace, but you must be using istio sidecar proxy, 
and ensure your seldon operator configuration has the following:
```
istio:
  enabled: true
  tlsMode: STRICT
```

When you've set up an `AuthorizationPolicy`, this will disrupt Prometheus from scraping metrics. Two proposed options to 
resolve this issue are: 
- You can specify that you want to allow GET requests to the prometheus endpoint in the `AuthorizationPolicy`

Example:
```
  - to:
    - operation:
        methods: ["GET"]
        paths: ["/prometheus"]
        ports: ["6000", "8000", "6001"]
```

- You can also exclude ports in your Istio Operator configuration
```
  proxy:
        autoInject: enabled
        clusterDomain: cluster.local
        componentLogLevel: misc:error
        enableCoreDump: false
        excludeInboundPorts: ""
        excludeOutboundPorts: "15021"
```

## Troubleshoot
If you saw errors like `Failed to generate bootstrap config: mkdir ./etc/istio/proxy: permission denied`, it's probably because you are running istio version <= 1.6.
Istio proxy sidecar by default needs to run as root (This changed in version >= 1.7, non-root is the default)
You can fix this by changing `defaultUserID=0` in your helm chart, or add the following `securityContext` to your istio proxy sidecar.

```
securityContext:
  runAsUser: 0
```
