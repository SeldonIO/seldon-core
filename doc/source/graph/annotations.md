# Annotation Based Configuration

You can configure aspects of Seldon Core via annotations in the SeldonDeployment resource and also the optional API OAuth Gateway. Please create an issue if you would like some configuration added.

## SeldonDeployment Annotations

### gRPC API Control

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size (bytes)
   * Locations : SeldonDeployment.spec.annotations
   * [gRPC message size example](model_rest_grpc_settings.md)
 * ```seldon.io/grpc-read-timeout``` : gRPC read timeout (msecs)
   * Locations : SeldonDeployment.spec.annotations
   * [gRPC read timeout example](model_rest_grpc_settings.md)


### REST API Control

 * ```seldon.io/rest-read-timeout``` : REST read timeout (msecs)
   * Locations : SeldonDeployment.spec.annotations
   * [REST read timeout example](model_rest_grpc_settings.md)
 * ```seldon.io/rest-connection-timeout``` : REST connection timeout (msecs)
   * Locations : SeldonDeployment.spec.annotations
   * [REST read connection timeout example](model_rest_grpc_settings.md)

### Service Orchestrator

  * ```seldon.io/engine-separate-pod``` : Use a separate pod for the service orchestrator
    * Locations : SeldonDeployment.spec.annotations
    * [Separate svc-orc pod example](model_svcorch_sep.md)
  * ```seldon.io/headless-svc``` : Run main endpoint as headless kubernetes service. This is required for gRPC load balancing via Ambassador.
    * Locations : SeldonDeployment.spec.annotations
    * [gRPC headless example](grpc_load_balancing_ambassador.md)


### Misc

 * ```seldon.io/svc-name``` : Custom service name for predictor. You will be responsible that it doesn't clash with any existing service name in the namespace of the deployed SeldonDeployment.
   * Locations : SeldonDeployment.spec.predictors[].annotations
   * [custom service name example](custom_svc_name.md)

