# Annotation Based Configuration

You can configure aspects of Seldon Core via annotations in the SeldonDeployment resource and also the optional API OAuth Gateway. Please create an issue if you would like some configuration added.

## SeldonDeployment Annotations

### gRPC API Control

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size (bytes)
   * Locations : SeldonDeployment.spec.annotations
   * Default is MaxInt32
   * [gRPC message size example](model_rest_grpc_settings.md)
 * ```seldon.io/grpc-timeout``` : gRPC timeout (msecs)
   * Locations : SeldonDeployment.spec.annotations
   * Default is no timeout
   * [gRPC timeout example](model_rest_grpc_settings.md)


### REST API Control

.. Note:: 
   When using REST APIs, timeouts will only apply to each node and not to the
   full inference graph.
   Therefore, each sub-request for each individual node in the graph will be
   able to take up to ``seldon.io/rest-timeout`` milliseconds.

* ```seldon.io/rest-timeout``` : REST timeout (msecs)
  * Locations : SeldonDeployment.spec.annotations
  * Default is no overall timeout but will use GoLang's default transport settings which include a 30 sec connection timeout.
  * [REST timeout example](model_rest_grpc_settings.md)


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

