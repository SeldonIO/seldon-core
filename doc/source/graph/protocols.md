# Protocols

Tensorflow protocol is only available in version >=1.1.

Seldon Core supports the following data planes:

 * [REST and gRPC Seldon protocol](#rest-and-grpc-seldon-protocol)
 * [REST and gRPC Tensorflow Serving Protocol](#rest-and-grpc-tensorflow-protocol)
 * [REST and gRPC V2 KFServing Protocol](#v2-kfserving-protocol)

## REST and gRPC Seldon Protocol

 * [REST Seldon Protocol](../reference/apis/index.html)

Seldon is the default protocol for SeldonDeployment resources. You can specify the gRPC protocol by setting `transport: grpc` in your SeldonDeployment resource or ensuring all components in the graph have endpoint.tranport set ot grpc.

See [example notebook](../examples/protocol_examples.html). 

## REST and gRPC Tensorflow Protocol

   * [REST Tensorflow Protocol definition](https://github.com/tensorflow/serving/blob/master/tensorflow_serving/g3doc/api_rest.md).
   * [gRPC Tensorflow Procotol definition](https://github.com/tensorflow/serving/blob/master/tensorflow_serving/apis/prediction_service.proto).

Activate this protocol by speicfying `protocol: tensorflow` and `transport: rest` or `transport: grpc` in your Seldon Deployment. See [example notebook](../examples/protocol_examples.html). 

For Seldon graphs the protocol will work as expected for single model graphs for Tensorflow Serving servers running as the single model in the graph. For more complex graphs you can chain models:

 * Sending the response from the first as a request to the second. This will be done automatically when you defined a chain of models as a Seldon graph. It is up to the user to ensure the response of each changed model can be fed a request to the next in the chain.
 * Only Predict calls can be handled in multiple model chaining.


General considerations:

  * Seldon components marked as MODELS, INPUT_TRANSFORMER and OUTPUT_TRANSFORMERS will allow a PredictionService Predict method to be called.
  * GetModelStatus for any model in the graph is available.
  * GetModelMetadata for any model in the graph is available.
  * Combining and Routing with the Tensorflow protocol is not presently supported.
  * `status` and `metadata` calls can be asked for any model in the graph
  * a non-standard Seldon extension is available to call predict on the graph as a whole: `/v1/models/:predict`.
  * The name of the model in the `graph` section of the SeldonDeployment spec must match the name of the model loaded onto the Tensorflow Server.


## V2 KFServing Protocol 

Seldon has collaborated with the [NVIDIA Triton Server Project](https://github.com/triton-inference-server/server) and the [KFServing Project](https://github.com/kubeflow/kfserving) to create a new ML inference protocol.

This protocol can be used by specifing `protocol: kfserving`. At present we only support NVIDIA Triton prepackaged servers but will have support for custom and other server types in the near future.

See [example notebook](../examples/protocol_examples.html). 
