# Annotation Based Configuration

You can configure aspects of Seldon Core via annotations in the SeldonDeployment resource and also the optional API OAuth Gateway. Please create an issue if you would like some configuration added.

## SeldonDeployment Annotations

### gRPC API Control

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size
   * Locations : SeldonDeployment.spec.annotations
   * [gRPC message size example](model_rest_grpc_settings.md)
 * ```seldon.io/grpc-read-timeout``` : gRPC read timeout
   * Locations : SeldonDeployment.spec.annotations
   * [gRPC read timeout example](model_rest_grpc_settings.md)


### REST API Control

 * ```seldon.io/rest-read-timeout``` : REST read timeout
   * Locations : SeldonDeployment.spec.annotations
   * [REST read timeout example](model_rest_grpc_settings.md)
 * ```seldon.io/rest-connection-timeout``` : REST connection timeout
   * Locations : SeldonDeployment.spec.annotations
   * [REST read connection timeout example](model_rest_grpc_settings.md)

### Service Orchestrator

  * ```seldon.io/engine-separate-pod``` : Use a separate pod for the service orchestrator
    * Locations : SeldonDeployment.spec.annotations
    * [Separate svc-orc pod example](model_svcorch_sep.md)
  * ```seldon.io/headless-svc``` : Run main endpoint as headless kubernetes service. This is required for gRPC load balancing via Ambassador.
    * Locations : SeldonDeployment.spec.annotations
    * [gRPC headless example](grpc_load_balancing_ambassador.md)

Otherwise any annotations starting with `seldon.io/engine-` will be interpreted as specifying environment variables for the engine container. These include:

  * ```seldon.io/engine-java-opts``` : Java Opts for Service Orchestrator
    * Locations : SeldonDeployment.spec.predictors.annotations
    * [Java Opts example](model_engine_java_opts.md)
    * Translates to the environment variable JAVA_OPTS
  * ```seldon.io/engine-log-requests``` : Whether to log raw requests from engine
    * Locations : SeldonDeployment.spec.predictors.annotations
    * Translates to the environment variable LOG_REQUESTS
  * ```seldon.io/engine-log-responses``` : Whether to log raw responses from engine
    * Locations : SeldonDeployment.spec.predictors.annotations
    * Translates to the environment variable LOG_RESPONSES
  * ```seldon.io/engine-log-messages-externally``` : Option to turn on logging of requests via a logging service
    * Locations : SeldonDeployment.spec.predictors.annotations
    * Translates to the environment variable LOG_MESSAGES_EXTERNALLY
  * ```seldon.io/engine-log-message-type``` : Option to override type set on messages when sending to logging service. Used to determine which logger impl
    * Locations : SeldonDeployment.spec.predictors.annotations
    * Translates to the environment variable LOG_MESSAGE_TYPE
  * ```seldon.io/engine-message-logging-service``` : Option to override url to broker that sends to logging service
    * Locations : SeldonDeployment.spec.predictors.annotations
    * Translates to the environment variable MESSAGE_LOGGING_SERVICE

More details on logging-related variables can be seen in the [request-logging example](https://github.com/SeldonIO/seldon-core/tree/master/examples/centralised-logging/README.md).

Environment variables for the engine can also be set in the `svcOrchSpec` section of the SeldonDeployment, alongside engine resources. For examples see the helm charts or the [distributed tracing example](./distributed-tracing.md).

If both annotations and `svcOrchSpec` environment variables are used to set an environment variable for the engine container then `svcOrchSpec` environment variables take priority.

The above are the key engine env vars. For a full listing of engine env vars see the application.properties file of the engine source code.

## API OAuth Gateway Annotations
The API OAuth Gateway, if used, can also have the following annotations:

### gRPC API Control

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size
 * ```seldon.io/grpc-read-timeout``` : gRPC read timeout


### REST API Control

 * ```seldon.io/rest-read-timeout``` : REST read timeout
 * ```seldon.io/rest-connection-timeout``` : REST connection timeout


### Control via Helm
The API OAuth Gateway annotations can be set via Helm via the seldon-core values file, for example:

```yaml
apife:
  annotations:
      seldon.io/grpc-max-message-size: "10485760"
```
