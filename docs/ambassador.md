# Deployment options running Seldon Core with Ambassador

Seldon Core works well with Ambassador (reverse proxy) handling ingress to your running machine learning deployments. In this doc we will discuss how your Seldon Deployments are exposed via Ambassador and how you can use both to do various production rollout strategies.

## Ambassador REST

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>``` running in a namespace ```namespace```:

For Seldon Core restricted to a namespace, `singleNamespace=true`, the endpoints exposed are:

 * ```http://<ambassadorEndpoint>/seldon/<deploymentName>/api/v0.1/predictions```
 * ```http://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v0.1/predictions```

For Seldon Core running cluster wide, `singleNamespace=false`, the endpoints exposed are all namespaced:

 * ```http://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v0.1/predictions```


## Example Curl

### Ambassador REST

Assuming a Seldon Deployment ```mymodel``` with Ambassador exposed on 0.0.0.0:8003:

```
curl -v 0.0.0.0:8003/seldon/mymodel/api/v0.1/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

## Canary Deployments

Canary rollouts are available where you wish to push a certain percentage of traffic to a new model to test whether it works ok in production. You simply need to add some annotations to your Seldon Deployment resource for your canary deployment.

  * `seldon.io/ambassador-weight`:`<weight_value>` : The weight (a value between 0 and 100) to be applied to this deployment.
     * Example: `"seldon.io/ambassador-weight":"25"`
  * `seldon.io/ambassador-service-name`:`<existing_deployment_name>` : The name of the existing Seldon you want to attach to as a canary.
     * Example: "seldon.io/ambassador-service-name":"example"

A worked example notebook can be found [here](https://github.com/SeldonIO/seldon-core/blob/master/examples/ambassador/canary/ambassador_canary.ipynb)

To understand more about the Ambassador configuration for this see [their docs](https://www.getambassador.io/reference/canary/).

## Shadow Deployments

Shadow deployments allow you to send duplicate requests to a parallel deployment but throw away the response. This allows you to test machine learning models under load and compare the results to the live deployment. 

You simply need to add some annotations to your Seldon Deployment resource for your shadow deployment.

  * `seldon.io/ambassador-shadow`:`true` : Flag to mark this deployment as a Shadow deployment in Ambassador.
  * `seldon.io/ambassador-service-name`:`<existing_deployment_name>` : The name of the existing Seldon you want to attach to as a shadow.
     * Example: "seldon.io/ambassador-service-name":"example"

A worked example notebook can be found [here](https://github.com/SeldonIO/seldon-core/blob/master/examples/ambassador/shadow/ambassador_shadow.ipynb)

To understand more about the Ambassador configuration for this see [their docs](https://www.getambassador.io/reference/shadowing/).

## Header based Routing

## Custom Amabassador configuration

