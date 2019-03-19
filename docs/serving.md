# Serving Predictions

Depending on whether you deployed Seldon Core with Ambassador or the API Gateway you can access your models as discussed below:

## Ambassador

### Ambassador REST

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```:

 * A REST endpoint will be exposed at : ```http://<ambassadorEndpoint>/seldon/<deploymentName>/api/v0.1/predictions```


### Ambassador gRPC

Assuming Ambassador is exposed at ```<ambassadorEndpoint>``` and with a Seldon deployment name ```<deploymentName>```:

  * A gRPC endpoint will be exposed at ```<ambassadorEndpoint>``` and you should send metadata in your request with key ```seldon``` and value ```<deploymentName>```.


## API OAuth Gateway

The HTTP and OAuth endpoints will be on separate ports, default is 8080 (HTTP) and 5000 (gRPC).

### OAuth REST

Assuming the API Gateway is exposed at ```<APIGatewayEndpoint>```

 1. You should get an OAuth token from ```<APIGatewayEndpoint>/oauth/token```
 1. You should make prediction requests to ```<APIGatewayEndpoint>/api/v0.1/predictions``` with the OAuth token in the header as ```Authorization: Bearer <token>```

### OAuth gRPC

Assuming the API gRPC Gateway is exposed at ```<APIGatewayEndpoint>```

 1. You should get an OAuth token from ```<APIGatewayEndpoint>/oauth/token```
 1. Send gRPC requests to ```<APIGatewayEndpoint>``` with the OAuth token in the meta data as ```oauth_token: <token>```

## Client Implementations

### Curl Examples

#### Ambassador REST

Assuming a SeldonDeplotment ```mymodel``` with Ambassador exposed on 0.0.0.0:8003:

```
curl -v 0.0.0.0:8003/seldon/mymodel/api/v0.1/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```


#### API OAuth Gateway REST

Assume server is accessible at 0.0.0.0:8002.

Get a token. Assuming the OAuth key is ```oauth-key``` and OAuth secret is ```oauth-secret``` as specified in the SeldonDeployment graph you created:

```
TOKENJSON=$(curl -XPOST -u oauth-key:oauth-secret 0.0.0.0:8002/oauth/token -d 'grant_type=client_credentials')
TOKEN=$(echo $TOKENJSON | jq ".access_token" -r)
```

Get predictions
```
curl -w "%{http_code}\n" --header "Authorization: Bearer $TOKEN" 0.0.0.0:8002/api/v0.1/predictions -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}' -H "Content-Type: application/json"
```

### OpenAPI REST

Use Swagger to generate a client for you from the [OpenAPI specifications](../openapi/README.md).

### gRPC

Use [gRPC](https://grpc.io/) tools in your desired language from the [proto buffer specifications](../proto/prediction.proto).

#### Example Python

See [example python code](../notebooks/seldon_utils.py).

