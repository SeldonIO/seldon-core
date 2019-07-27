# Seldon Python Client

We provide an example python client for calling the API using REST or gRPC for internal mciroservice testing or for calling the public external API.

Examples of its use can be found in various notebooks:

  * [Helm based deployment examples](../examples/helm_examples.html)
  * [Istio examples](../examples/istio_examples.html)

To use the client simply create an instance with settings for your use case, for example:

```
from seldon_core.seldon_client import SeldonClient
sc = SeldonClient(deployment_name="mymodel",namespace="seldon",gateway_endpoint="localhost:8003",gateway="ambassador")
```

In the above we set our deployment_name to "mymodel" and the namespace to "seldon". For the full set of parameters see [here](api/seldon_core.html#seldon_core.seldon_client.SeldonClient)

To make a REST call with a random payload:

```
r = sc.predict(transport="rest")
```

To make a gRPC call with random payload:

```
r = sc.predict(transport="grpc")
```


## Advanced Examples

 * [SSL and Authentication](../examples/seldon_client.html)
