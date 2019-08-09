# Serving Predictions

Depending on whether you deployed Seldon Core with Ambassador or the API Gateway you can access your models as discussed below:

## Ambassador

### Ambassador REST

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```:

 * A REST endpoint will be exposed at : ```http://<ambassadorEndpoint>/seldon/<deploymentName>/api/v0.1/predictions```


### Ambassador gRPC

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```:

  * A gRPC endpoint will be exposed at ```<ambassadorEndpoint>``` and you should send metadata in your request with key ```seldon``` and value ```<deploymentName>```.


## Client Implementations

### Curl Examples

#### Ambassador REST

Assuming a SeldonDeployment ```mymodel``` with Ambassador exposed on 0.0.0.0:8003:

```bash
curl -v 0.0.0.0:8003/seldon/mymodel/api/v0.1/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

### OpenAPI REST

Use Swagger to generate a client for you from the [OpenAPI specifications](../reference/apis/openapi.html).

### gRPC

Use [gRPC](https://grpc.io/) tools in your desired language from the [proto buffer specifications](../reference/apis/prediction.md).

#### Reference Python Client

Use our [reference python client](../python/python_module.md) which is part of the `seldon-core` module.

