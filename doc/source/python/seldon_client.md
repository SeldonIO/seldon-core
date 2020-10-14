# Seldon Python Client

We provide an example python client for calling the API using REST or gRPC for internal microservice testing or for calling the public external API.

Examples of its use can be found in most of the [Seldon Core examples](../examples/notebooks.html).

To use the client simply create an instance with settings for your use case, for example:

```python
from seldon_core.seldon_client import SeldonClient
sc = SeldonClient(deployment_name="mymodel",namespace="seldon",gateway_endpoint="localhost:8003",gateway="ambassador")
```

In the above we set our deployment_name to "mymodel" and the namespace to "seldon". For the full set of parameters see [here](./api/seldon_core.html#seldon_core.seldon_client.SeldonClient).

To make a REST call with a random payload:

```python
r = sc.predict(transport="rest")
```

To make a gRPC call with random payload:

```python
r = sc.predict(transport="grpc")
```

The default return type is the protobuf of type "SeldonMessage", but you can also choose to return a JSON dictionary which could make it easier for interacting with the internal data. You can do this by passing the kwarg `client_return_type` with the value `dict` (default value is `proto`) when either initialising your Seldon Client or when sending a predict request. For example:

```python
sc = SeldonClient(..., client_return_type="dict")
```

Or alternatively you can pass it when sending the request to override your default parameter:

```python
sc = SeldonClient(..., client_return_type="proto")

sc.predict(..., client_return_type="dict") # Here we override it
```

## Advanced Examples

 * [SSL and Authentication](../examples/seldon_client.html)
