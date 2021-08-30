# Tensorflow Serving

If you have a trained Tensorflow model you can deploy this directly via REST or gRPC servers. 

## MNIST Example

### REST MNIST Example

For REST you need to specify parameters for:

 * signature_name
 * model_name

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tfserving
spec:
  name: mnist
  predictors:
  - graph:
      children: []
      implementation: TENSORFLOW_SERVER
      modelUri: gs://seldon-models/tfserving/mnist-model
      name: mnist-model
      parameters:
        - name: signature_name
          type: STRING
          value: predict_images
        - name: model_name
          type: STRING
          value: mnist-model
    name: default
    replicas: 1

```

### gRPC MNIST Example

For gRPC you need to specify the following parameters:

 * signature_name
 * model_name
 * model_input
 * model_output

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tfserving
spec:
  name: mnist
  predictors:
  - graph:
      children: []
      implementation: TENSORFLOW_SERVER
      modelUri: gs://seldon-models/tfserving/mnist-model
      name: mnist-model
      endpoint:
        type: GRPC
      parameters:
        - name: signature_name
          type: STRING
          value: predict_images
        - name: model_name
          type: STRING
          value: mnist-model
        - name: model_input
          type: STRING
          value: images
        - name: model_output
          type: STRING
          value: scores          
    name: default
    replicas: 1

```


Try out a [worked notebook](../examples/server_examples.html)


## Multi-Model Serving

You can utilize Tensorflow Serving's functionality to load multiple models from one model repository as shown in this [example notebook](../examples/protocol_examples.html). You should follow the configuration details as disucussed in the [Tensorflow Serving documentation on advanced configuration](https://www.tensorflow.org/tfx/serving/serving_config).

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-tfserving
spec:
  protocol: tensorflow
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - args: 
          - --port=8500
          - --rest_api_port=8501
          - --model_config_file=/mnt/models/models.config
          image: tensorflow/serving
          name: multi
          ports:
          - containerPort: 8501
            name: http
            protocol: TCP
          - containerPort: 8500
            name: grpc
            protocol: TCP
    graph:
      name: multi
      type: MODEL
      implementation: TENSORFLOW_SERVER
      modelUri: gs://seldon-models/tfserving/multi-model
      endpoint:
        httpPort: 8501
        grpcPort: 8500
    name: model
    replicas: 1
```
