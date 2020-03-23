# Testing Your Model Endpoints

In order to test your components you are able to send the requests directly using CURL/grpCURL or a similar utility, as well as by using our Python SeldonClient SDK.

## Testing options

There are several options for testing your model before deploying it.

* Running your model directly with the Python Client
* Running your model as a Docker container
    * This can be used for all Language Wrappers (but not prepackaged inference servers)
* Run your SeldonDeployment in a Kubernetes Dev client such as KIND
    * This can be used for any models

### Running your model directly with the Python Client

* This can be used for Python Language Wrapped Models only

When you create your Python model, such as a file called `Model.py` with the contents:

```
class Model:
    def __init__(self):
        pass
    def predict(*args, **kwargs):
        pass
```

You can actually run this model by simply using the Python module.

Once you install the Python module you will have some 



### Test using the Python Client

If you have containerised a model using our Python Language server, you will be able to test it locally.


## Pre-requisites

First you need to make sure you've deployed your model, and the model is available through one of the supported [Ingress (as outlined in installation docs)](../workflow/install.md) you are able

Depending on whether you deployed Seldon Core with Ambassador or the API Gateway you can access your models as discussed below:

## Ambassador

### Ambassador REST

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```  in namespace ```<namespace>```::

 * A REST endpoint will be exposed at : ```http://<ambassadorEndpoint>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions```


### Ambassador gRPC

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```:

  * A gRPC endpoint will be exposed at ```<ambassadorEndpoint>``` and you should send header metadata in your request with:
    * key ```seldon``` and value ```<deploymentName>```.
    * key ```namespace``` and value ```<namespace>```.

## Istio

### Istio REST

Assuming the istio gateway is at ```<istioGateway>``` and with a Seldon deployment name ```<deploymentName>``` in namespace ```<namespace>```:

 * A REST endpoint will be exposed at : ```http://<istioGateway>/seldon/<namespace>/<deploymentName>/api/v1.0/predictions```


### Istio gRPC

Assuming the istio gateway is at ```<istioGateway>``` and with a Seldon deployment name ```<deploymentName>``` in namespace ```<namespace>```:

  * A gRPC endpoint will be exposed at ```<istioGateway>``` and you should send header metadata in your request with:
    * key ```seldon``` and value ```<deploymentName>```.
    * key ```namespace``` and value ```<namespace>```.


## Client Implementations

### Curl Examples

#### Ambassador REST

Assuming a SeldonDeployment ```mymodel``` with Ambassador exposed on 0.0.0.0:8003:

```bash
curl -v 0.0.0.0:8003/seldon/mymodel/api/v1.0/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

### OpenAPI REST

Use Swagger to generate a client for you from the [OpenAPI specifications](../reference/apis/openapi.html).

### gRPC

Use [gRPC](https://grpc.io/) tools in your desired language from the [proto buffer specifications](../reference/apis/prediction.md).

#### Reference Python Client

Use our [reference python client](../python/python_module.md) which is part of the `seldon-core` module.

