# Seldon Core Release 0.4.1

A summary of the main contributions to the [Seldon Core release 0.4.1](https://github.com/SeldonIO/seldon-core/releases/tag/v0.4.1).

## Black Box Model Explanations

By utlizing Seldon'sopen source Model Explanation library [Alibi](https://github.com/SeldonIO/alibi) we provide the ability to launch a model and an associated explainer for that model. At present we support the [Anchors](https://homes.cs.washington.edu/~marcotcr/aaai18.pdf) explanation technique for tabular text and image examples.

An example SeldonDeployment for an image model with associated explainer is shown below:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: image
spec:
  annotations:
    seldon.io/rest-read-timeout: "10000000"
    seldon.io/grpc-read-timeout: "10000000"
    seldon.io/grpc-max-message-size: "1000000000"
  name: image
  predictors:
  - graph:
      children: []
      implementation: TENSORFLOW_SERVER
      modelUri: gs://seldon-models/tfserving/imagenet/model
      name: classifier
      endpoint:
        type: GRPC
      parameters:
        - name: model_name
          type: STRING
          value: classifier
        - name: model_input
          type: STRING
          value: input_image
        - name: model_output
          type: STRING
          value: predictions/Softmax:0
    engineResources:
      requests:
        memory: 1Gi
    explainer:
      type: anchor_images
      modelUri: gs://seldon-models/tfserving/imagenet/explainer
    name: default
    replicas: 1

```

The Tensorflow model has a `anchor_images` explainer associated with it. An example input showing a persian cat along with an example explanation for that image showing the segment of the image the model focused on for providing the classifcation result can be seen below.

![cat](../analytics/cat.png)

![cat-explanation](../analytics/cat_explanation.png)

We provide [an example notebook with tabular, text and image model examples](../examples/explainer_examples.html).

## Misc. Updates

 * [Chainer Example](../examples/chainer_mnist.html)
 * We now use [Jenkins X](https://jenkins.io/projects/jenkins-x/) as our project CICD platform.
 * We are available on the [RedHat Container Catalog](https://access.redhat.com/containers/?tab=overview#/registry.connect.redhat.com/seldonio/seldon-operator-0-4-0). Update to 0.4.1 soon.

[Join our slack community to discuss](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU).







