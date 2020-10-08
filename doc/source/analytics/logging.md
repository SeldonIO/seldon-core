# Payload Logging

Logging of request and response payloads from your Seldon Deployment can be accomplished by adding a logging section to any part of the Seldon deployment graph. An example is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
      logger:
        url: http://mylogging-endpoint
        mode: all
    name: example
    replicas: 1

```

The logging for the top level requets response is provided by:

```yaml
      logger:
        url: http://mylogging-endpoint
        mode: all
```

In this example both request and response payloads as specified by the `mode` attribute are sent as CloudEvents to the url `http://mylogging-endpoint`.

The specification is:

 * url: Any url. Optional. If not provided then it will default to the default knative borker in the namespace of the Seldon Deployment.
 * mode: Either `request`, `response` or `all`

## Setting Global Default

If you don't want to set up the custom logger every time, you are able to set it with the defaultRequestLoggerEndpointPrefix Helm Chart Variable as outlined in the [helm chart advanced settings section](../reference/helm.rst). 

You just have to provide the prefix, which would then always be suffixed by ".<namespace>" where `namespace` is the namespace where your model is running, and hence there will be a requirement to run a request logger on every namespace.

An example would be setting it to a request logger who's service can be accessible through `custom-request-logger`, and assuming we deploy our request logger in the namespace `deep-learning`, then we should set the helm variable as:

```yaml
#...other variables
executor:
  defaultRequestLoggerEndpointPrefix: 'http://default-broker.'
#...other variables
```

So when the model runs in the `deep-learning` namespace, it will send all the input and output requests to the service `http://default-broker.deep-learning`.

You will still need to make sure the model is deployed with a specification on what requests will be logged, i.e. all, request or response (as outlined above).


### Example Notebook

You can try out an [example notebook with logging](../examples/payload_logging.html)

