# Annotation Based Configuration

You can configure aspects of Seldon Core via annotations in the SeldonDeployment resource and also the optional API OAuth Gateway. Please create an issue if you would like some configuration added.

## SeldonDeployment Annotations

### gRPC API Control

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size
   * Locations : SeldonDeployment.spec.annotations
   * [Example](model_rest_grpc_settings.md)
 * ```seldon.io/grpc-read-timeout``` : gRPC read timeout
   * Locations : SeldonDeployment.spec.annotations
   * [Example](model_rest_grpc_settings.md)


### REST API Control

 * ```seldon.io/rest-read-timeout``` : REST read timeout
   * Locations : SeldonDeployment.spec.annotations
   * [Example](model_rest_grpc_settings.md)
 * ```seldon.io/rest-connection-timeout``` : REST connection timeout
   * Locations : SeldonDeployment.spec.annotations
   * [Example](model_rest_grpc_settings.md)

### Service Orchestrator

  * ```seldon.io/engine-java-opts``` : Java Opts for Service Orchestrator
    * Locations : SeldonDeployment.spec.predictors.annotations
    * [Example](model_engine_java_opts.md)
  * ```seldon.io/engine-separate-pod``` : Use a separate pod for the service orchestrator
    * Locations : SeldonDeployment.spec.annotations
    * [Example](model_svcorch_sep.md)
  * ```seldon.io/headless-svc``` : Run main endpoint as headless kubernetes service. This is required for gRPC load balancing via Ambassador.
    * Locations : SeldonDeployment.spec.annotations
    * [Example](grpc_load_balancing_ambassador.md)

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

```
apife:
  annotations:
      seldon.io/grpc-max-message-size: "10485760"
```
