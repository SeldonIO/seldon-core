# Model

A Model is the core atomic building block. It specifies a machine learning artifact that will be loaded onto one of the running Servers. A model could be a standard machine learning inference component such as

 * a Tensorflow model, PyTorch model or SKLearn model.
 * an inference transformation component such as a SKLearn pipeline or a piece of custom python logic.
 a monitoring component such as an outlier detector or drift detector.
 * An alibi-explain model explainer

An example is shown below for a SKLearn model for iris classification:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/samples/models/sklearn-iris-gs.yaml" %}

```yaml
# samples/models/sklearn-iris-gs.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.5.0/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki
```

Its Kubernetes `spec` has two core requirements

 * A `storageUri` specifying the location of the artifact. This can be any rclone URI specification.
 * A `requirements` list which provides tags that need to be matched by the Server that can run this artifact type. By default when you install Seldon we provide a set of Servers that cover a range of artifact types.


