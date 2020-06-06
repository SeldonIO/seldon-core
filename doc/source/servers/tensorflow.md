# Tensorflow Serving

If you have a trained Tensorflow model you can deploy this directly via REST or gRPC servers. 

## MNIST Example

### REST MNIST Example

For REST you need to specify paramaters for:

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
