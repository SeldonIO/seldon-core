# Ingress with Ambassador

Seldon Core works well with [Ambassador](https://www.getambassador.io/), allowing a single ingress to be used to expose Ambassador and [running machine learning deployments that can then be dynamically exposed](https://kubernetes.io/blog/2018/06/07/dynamic-ingress-in-kubernetes/) through Seldon-created Ambassador configurations. Ambassador is a Kubernetes-native API Gateway built on Envoy Proxy. Managed entirely via Kubernetes Custom Resource Definitions, Ambassador provides powerful capabilities for traffic management, authentication, and observability. Ambassador has native integrations for popular service meshes, including Consul, Istio, and Linkerd.  In this doc we will discuss how your Seldon Deployments are exposed via Ambassador and how you can use both to do various production rollout strategies.

## Installing Ambassador

You have [two options](https://www.getambassador.io/editions/) when installing Ambassador:

### Option 1: Ambassador Edge Stack

The [Ambassador Edge Stack](https://www.getambassador.io/products/edge-stack/) is the easiest way to get started with Ambassador, providing easy, [Automatic TLS configuration](https://www.getambassador.io/docs/edge-stack/latest/topics/running/host-crd/#acme-support) in addition to other features, like [Authentication](https://www.getambassador.io/docs/edge-stack/latest/topics/using/filters/), [Rate Limiting](https://www.getambassador.io/docs/edge-stack/latest/topics/using/rate-limits/), and advanced routing behaviors, like [custom prefixes](https://www.getambassador.io/docs/edge-stack/latest/topics/using/intro-mappings/), [virtual hosting](https://www.getambassador.io/docs/edge-stack/latest/topics/using/headers/host/), [method](https://www.getambassador.io/docs/edge-stack/latest/topics/using/method/)/[query-parameter](https://www.getambassador.io/docs/edge-stack/latest/topics/using/query_parameters/) based routing, and many others.
Follow the [AES installation instructions](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/) to install it on your Kubernetes cluster.

Using `helm` the steps can be summarised as
```bash
kubectl create namespace ambassador || echo "namespace ambassador exists"

helm repo add datawire https://www.getambassador.io
helm install ambassador datawire/ambassador \
  --set image.repository=docker.io/datawire/ambassador \
  --set crds.keep=false \
  --namespace ambassador
```

### Option 2: Ambassador API Gateway (now Emissary Ingress)

The [Ambassador API Gateway](https://www.getambassador.io/products/api-gateway/) (now Emissary Ingress) is the fully open source version of Ambassador Edge Stack and provides all the functionality of a traditional ingress controller.
Follow the [Ambassador OSS instructions](https://www.getambassador.io/docs/latest/topics/install/install-ambassador-oss/) to install it on your kubernetes cluster.

Using `helm` the steps can be summarised as
```bash
kubectl create namespace ambassador || echo "namespace ambassador exists"

helm repo add datawire https://www.getambassador.io
helm install ambassador datawire/ambassador \
  --set image.repository=docker.io/datawire/ambassador \
  --set enableAES=false \
  --set crds.keep=false \
  --namespace ambassador
```

### TLS

It is highly recommended to utilize TLS to encrypt traffic that is coming into Ambassador.  The default Ambassador Edge Stack installation comes with a self-signed certificate that will be used absent any other TLS configuration.  Follow the [instructions](https://www.getambassador.io/docs/edge-stack/latest/howtos/tls-termination/) for setting up TLS on your cluster.

## Ambassador REST

Assuming Ambassador is exposed at `<ambassadorEndpoint>` and with a Seldon deployment name `<deploymentName>` running in a namespace `namespace`:

Note, if you chose to install the Ambassador Edge Stack then you will need to use https.  You can either [set up TLS](https://www.getambassador.io/docs/edge-stack/latest/howtos/tls-termination/) or pass the `-k` flag in `curl` to allow the self-signed certificate.

For Seldon Core restricted to a namespace, `singleNamespace=true`, the endpoints exposed are:

 * `http(s)://<ambassadorEndpoint>/seldon/<deploymentName>/api/v1.0/predictions`
 * `http(s)://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions`

For Seldon Core running cluster wide, `singleNamespace=false`, the endpoints exposed are all namespaced:

 * `http(s)://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions`

## Example Curl

### Ambassador REST

If you installed the OSS Ambassador API Gateway, and assuming a Seldon Deployment `mymodel` with Ambassador exposed on `0.0.0.0:8003` you can send a curl request as follows:

```bash
curl -v 0.0.0.0:8003/seldon/mymodel/api/v1.0/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

Alternatively, if you installed the Ambassador Edge Stack with TLS configured, and assuming a Seldon Deployment `mymodel` with the Ambassador hostname `example-hostname.com`:

```bash
curl -v https://example-hostname.com/seldon/mymodel/api/v1.0/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

If you did not, you can use the exposed IP address in place of `example-hostname` and pass the -k flag for insecure TLS.

```bash
curl -vk https://0.0.0.0/seldon/mymodel/api/v1.0/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

## Ambassador Configuration Annotations Reference

| Annotation | Description |
|------------|-------------|
|`seldon.io/ambassador-config:<configuration>`| Custom Ambassador Configuration |
|`seldon.io/ambassador-header:<header>`| The header to add to Ambassador configuration |
|`seldon.io/ambassador-id:<instance id>`| The instance id to be added to Ambassador `ambassador_id` configuration |
|`seldon.io/ambassador-regex-header:<regex>`| The regular expression header to use for routing via headers|
|`seldon.io/ambassador-retries:<number of retries>` | The number of times ambassador will retry request on connect-failure. Default 0. Use custom configuration if more control needed.|
|`seldon.io/ambassador-service-name:<existing_deployment_name>`| The name of the existing Seldon Deployment for shadow or header based routing |
|`seldon.io/grpc-timeout: <gRPC read timeout (msecs)>` | gRPC read timeout |
|`seldon.io/rest-timeout:<REST read timeout (msecs)>` | REST read timeout |
|`seldon.io/ambassador-circuit-breakers-max-connections:<maximum number of connections>` | The maximum number of connections will make to the Seldon Deployment |
|`seldon.io/ambassador-circuit-breakers-max-pending-requests:<maximum number of queued requests>` | The maximum number of requests that will be queued while waiting for a connection |
|`seldon.io/ambassador-circuit-breakers-max-requests:<maximum number of parallel outstanding requests>` | The maximum number of parallel outstanding requests to the Seldon Deployment |
|`seldon.io/ambassador-circuit-breakers-max-retries:<maximum number of parallel retries>` | The maximum number of parallel retries allowed to the Seldon Deployment |

All annotations should be placed in `spec.annotations`.

See below for details.

### Canary Deployments

Canary rollouts are available where you wish to push a certain percentage of traffic to a new model to test whether it works ok in production. To add a canary to your SeldonDeployment simply add a new predictor section and set the traffic levels for the main and canary to desired levels. For example:

```YAML
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: example
spec:
  name: canary-example
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.2.1
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: main
    replicas: 1
    traffic: 75
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.2.2
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: canary
    replicas: 1
    traffic: 25

```

The above example has a "main" predictor with 75% of traffic and a "canary" with 25%.

A worked example for [canary deployments](../examples/ambassador_canary.html) is provided.

### Shadow Deployments

Shadow deployments allow you to send duplicate requests to a parallel deployment but throw away the response. This allows you to test machine learning models under load and compare the results to the live deployment.

Simply set the `shadow` boolean in your shadow predictor.

A worked example for [shadow deployments](../examples/ambassador_shadow.html) is provided.

To understand more about the Ambassador configuration for this see [their docs on shadow deployments](https://www.getambassador.io/reference/shadowing/).

### Header based Routing

Header based routing allows you to route requests to particular Seldon Deployments based on headers in the incoming requests.

You simply need to add some annotations to your Seldon Deployment resource.

  * `seldon.io/ambassador-header:<header>` : The header to add to Ambassador configuration
     * Example:  `"seldon.io/ambassador-header":"location: london"	    `
  * `seldon.io/ambassador-regex-header:<header>` : The regular expression header to add to Ambassador configuration
     * Example:  `"seldon.io/ambassador-header":"location: lond.*"	    `
  * `seldon.io/ambassador-service-name:<existing_deployment_name>` : The name of the existing Seldon Deployment you want to attach to as an alternative mapping for requests.
     * Example: `"seldon.io/ambassador-service-name":"example"`

A worked example for [header based routing](../examples/ambassador_headers.html) is provided.

To understand more about the Ambassador configuration for this see [their docs on header based routing](https://www.getambassador.io/reference/headers).

### Circuit Breakers

By preventing additional connections or requests to an overloaded Seldon Deployment, circuit breakers help improve resilience of your system.

You simply need to add some annotations to your Seldon Deployment resource.

  * `seldon.io/ambassador-circuit-breakers-max-connections:<maximum number of connections>` : The maximum number of connections will make to the Seldon Deployment
     * Example:  `"seldon.io/ambassador-circuit-breakers-max-connections":"200"`
  * `seldon.io/ambassador-circuit-breakers-max-pending-requests:<maximum number of queued requests>` : The maximum number of requests that will be queued while waiting for a connection
     * Example:  `"seldon.io/ambassador-circuit-breakers-max-pending-requests":"100"`
  * `seldon.io/ambassador-circuit-breakers-max-requests:<maximum number of parallel outstanding requests>` : The maximum number of parallel outstanding requests to the Seldon Deployment
     * Example: `"seldon.io/ambassador-circuit-breakers-max-requests":"200"`
  * `seldon.io/ambassador-circuit-breakers-max-retries:<maximum number of parallel retries>` : The maximum number of parallel retries allowed to the Seldon Deployment
     * Example: `"seldon.io/ambassador-circuit-breakers-max-retries":"3"`

A worked example for [circuit breakers](../examples/ambassador_circuit_breakers.html) is provided.

To understand more about the Ambassador configuration for this see [their docs on circuit breakers](https://www.getambassador.io/docs/latest/topics/using/circuit-breakers/).

## Multiple Ambassadors in the same cluster

To avoid conflicts in a cluster with multiple ambassadors running, you can add the following annotation to your Seldon Deployment resource.

  * `seldon.io/ambassador-id:<instance id>`: The instance id to be added to Ambassador `ambassador_id` configuration

For example,

```YAML
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: multi-ambassadors
spec:
  annotations:
    seldon.io/ambassador-id: my_instance
  name: ambassadors-example
```

Note that your Ambassador instance must be configured with matching `ambassador_id`.

See [AMBASSADOR_ID](https://www.getambassador.io/docs/latest/topics/running/running/#ambassador_id) for details

### Custom Amabassador configuration

The above discussed configurations should cover most cases, however custom configuration is necessary to leverage the full feature-set of Ambassador Edge Stack as a traffic management and authentication tool.  There are two options for configuring Ambassador, based on how you want to manage the configuration: Custom Resource Definitions (CRD's) and Annotations.

Ambassador primarily utilizes [Custom Resource Definitions](https://www.getambassador.io/docs/edge-stack/latest/topics/concepts/gitops-continuous-delivery/#policies-declarative-configuration-and-custom-resource-definitions) for managing configuration.  These are custom `kind` resources that the Kubernetes API can read, and are used to update Ambassador's config in a way that is observable to the cluster as a whole (e.g. you can `kubectl get` these resources).  These CRD's can be managed independent of the Seldon Deploy itself.

Ambassador also supports annotation based configuration, which can be applied to a Seldon Deployment using the `seldon.io/ambassador-config` annotation key.  The overall formatting of the annotation-based config is the same as the CRD based config, except without the `metadata` and `spec` fields.  The following snippets demonstrate the difference between CRD and Annotation based configurations, and are functionally identical.

```yaml
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: seldon_example_rest_mapping
spec:
  prefix: /mycompany/ml/
  service: production-model-example.seldon:8000
  timeout_ms: 3000
```

 * `seldon.io/ambassador-config:<configuration>` : The custom ambassador configuration
    * Example: `"seldon.io/ambassador-config":"apiVersion: ambassador/v2\nkind: Mapping\nname: seldon_example_rest_mapping\nprefix: /mycompany/ml/\nservice: production-model-example.seldon:8000\ntimeout_ms: 3000"`

A worked example for [custom Ambassador config](../examples/ambassador_custom.html) is provided.
