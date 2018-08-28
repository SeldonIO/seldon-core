# Annotation Based Configuration

You can configure aspects of Seldon Core via annotations in the SeldonDeployment resource. Please create an issue if you would like some configuration added.

# Available Annotations

 * ```seldon.io/grpc-max-message-size``` : Maximum gRPC message size
   * Location : SeldonDeployment.spec.annotations
   * [Example](../notebooks/resources/model_grpc_size.json)