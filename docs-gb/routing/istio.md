# Ingress with Istio

Seldon Core can be used in conjunction with [istio](https://istio.io/). Istio provides an [ingress gateway](https://istio.io/docs/tasks/traffic-management/ingress/) which Seldon Core can automatically wire up new deployments to. The steps to using istio are described below.

## Install Seldon Core Operator

Ensure when you install the seldon-core operator via Helm that you enabled istio. For example:

```bash 
helm install seldon-core seldon-core-operator --set istio.enabled=true --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

You need an istio gateway installed in the `istio-system` namespace. By default we assume one called seldon-gateway. For example you can create this with the following yaml:

```yaml
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

```yaml
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
```yaml
istio:
  enabled: true
  tlsMode: STRICT
```

When you've set up an `AuthorizationPolicy`, this will disrupt Prometheus from scraping metrics. Two proposed options to 
resolve this issue are: 
- You can specify that you want to allow GET requests to the prometheus endpoint in the `AuthorizationPolicy`

Example:
```yaml
  - to:
    - operation:
        methods: ["GET"]
        paths: ["/prometheus"]
        ports: ["6000", "8000", "6001"]
```

- You can also exclude ports in your Istio Operator configuration
```yaml
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

```yaml
securityContext:
  runAsUser: 0
```


# Using the Istio Service Mesh
Istio can also be used to direct traffic internal to the cluster, rather than using it as an ingress (traffic from outside the cluster). 

To do this, the Virutal Services Seldon will create need to be attached to the "special" Gateway named `mesh`. This applies the routing rules to traffic inside the mesh without needing to route through a Gateway.

Due to limitations in Istio (as of v1.11.3), virtual services in the local mesh can only apply to one Host. (see their docs [here](https://istio.io/latest/docs/ops/best-practices/traffic-management/#split-virtual-services)). Therefor, a unique service is required for each Graph, which can be achieved by setting the `seldon.io/svc-name` annotation in the main predictor. 

Here's an example `SeldonDeployment` that will utilize the internal mesh networking to split traffic between two predictors, 75% to the first, 25% to the second: 
``` yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: canary-example-1
  namespace: my-ns
spec:
  annotations:
    seldon.io/istio-gateway: mesh # NOTE
    seldon.io/istio-host: canary-example-1 # NOTE
  name: canary-example-1
  predictors:
  - annotations:
      seldon.io/svc-name: canary-example-1   # NOTE
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.11.0
          imagePullPolicy: IfNotPresent
          name: classifier
          securityContext:
            readOnlyRootFilesystem: false
        terminationGracePeriodSeconds: 1
    graph:
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      sidecar.istio.io/inject: "true"
    name: main
    replicas: 1
    traffic: 75
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.11.0
          imagePullPolicy: IfNotPresent
          name: classifier
        terminationGracePeriodSeconds: 1
    graph:
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      sidecar.istio.io/inject: "true"
    name: canary
    replicas: 1
    traffic: 25
```

A few key things to point out:  
1. A unique service is created for the main (first) predictor named `canary-example-1`. This service cannot collide with any other services in the namespace. This service could be a service _not_ created via the SeldonDeployment, but also must match the necessary Istio routing rules. 
2. The above service is referenced in the annotations in `spec` by specify ing the host as follows:  `seldon.io/istio-host: canary-example-1`. This will set the host in the Istio Virutal Service to be the newly created service. 
3. The gateway is specified as `seldon.io/istio-gateway: mesh` to utilize this routing in the Istio Mesh. NOTE: In order to call this service, and have the appropriate routing take place, the Client _must_ also be _inside_ the mesh. This is accomplished by injecting the Istio Sidecar into the pod of the client. 

From within the cluster, and inside a pod that is inside the mesh, a call like the following will work, as well as split traffic between the two predictors:
``` shell
curl -X POST -H 'Content-Type: application/json' \
  -d '{"data": { "names": ["a", "b"], "ndarray": [[1,2]]}}' \
  http://mysvcname:8000/seldon/batest/canary-example-1/api/v1.0/predictions
```