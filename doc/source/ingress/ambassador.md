# Deployment Options running Seldon Core with Ambassador

Seldon Core works well with [Ambassador](https://www.getambassador.io/), allowing a single ingress to be used to expose ambassador and [running machine learning deployments can then be dynamically exposed](https://kubernetes.io/blog/2018/06/07/dynamic-ingress-in-kubernetes/) through seldon-created ambassador configurations. In this doc we will discuss how your Seldon Deployments are exposed via Ambassador and how you can use both to do various production rollout strategies.

## Ambassador REST

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>``` running in a namespace ```namespace```:

For Seldon Core restricted to a namespace, `singleNamespace=true`, the endpoints exposed are:

 * ```http://<ambassadorEndpoint>/seldon/<deploymentName>/api/v1.0/predictions```
 * ```http://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions```

For Seldon Core running cluster wide, `singleNamespace=false`, the endpoints exposed are all namespaced:

 * ```http://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions```


## Example Curl

### Ambassador REST

Assuming a Seldon Deployment ```mymodel``` with Ambassador exposed on `0.0.0.0:8003`:

```bash
curl -v 0.0.0.0:8003/seldon/mymodel/api/v1.0/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

## Canary Deployments

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
        - image: seldonio/mock_classifier:1.0
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
        - image: seldonio/mock_classifier_rest:1.1
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

## Shadow Deployments

Shadow deployments allow you to send duplicate requests to a parallel deployment but throw away the response. This allows you to test machine learning models under load and compare the results to the live deployment. 

You simply need to add some annotations to your Seldon Deployment resource for your shadow deployment.

  * `seldon.io/ambassador-shadow:true` : Flag to mark this deployment as a Shadow deployment in Ambassador.
  * `seldon.io/ambassador-service-name:<existing_deployment_name>` : The name of the existing Seldon Deployment you want to attach to as a shadow.
     * Example: `"seldon.io/ambassador-service-name":"example"`

A worked example for [shadow deployments](../examples/ambassador_shadow.html) is provided.

To understand more about the Ambassador configuration for this see [their docs on shadow deployments](https://www.getambassador.io/reference/shadowing/).

## Header based Routing

Header based routing allows you to route requests to particular Seldon Deployments based on headers in the incoming requests.

You simply need to add some annotations to your Seldon Deployment resource.

  * `seldon.io/ambassador-header:<header>` : The header to add to Ambassador configuration	    
     * Example:  `"seldon.io/ambassador-header":"location: london"	    `
  * `seldon.io/ambassador-service-name:<existing_deployment_name>` : The name of the existing Seldon you want to attach to as an alternative mapping for requests. 
     * Example: `"seldon.io/ambassador-service-name":"example"`

A worked example for [header based routing](../examples/ambassador_headers.html) is provided.

To understand more about the Ambassador configuration for this see [their docs on header based routing](https://www.getambassador.io/reference/headers).

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
See [AMBASSADOR_ID](https://github.com/datawire/ambassador/blob/master/docs/reference/running.md#ambassador_id) for details

## Custom Amabassador configuration

The above discussed configurations should cover most cases but there maybe a case where you want to have a very particular Ambassador configuration under your control. You can acheieve this by adding your confguration as an annotation to your Seldon Deployment resource.

 * `seldon.io/ambassador-config:<configuration>` : The custom ambassador configuration
    * Example: `"seldon.io/ambassador-config":"apiVersion: ambassador/v1\nkind: Mapping\nname: seldon_example_rest_mapping\nprefix: /mycompany/ml/\nservice: production-model-example.seldon:8000\ntimeout_ms: 3000"`

A worked example for [custom Ambassador config](../examples/ambassador_custom.html) is provided.

