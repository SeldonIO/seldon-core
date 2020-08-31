# Model Explainers

![cat](cat.png)
![explanation](cat_explanation.png)

Seldon provides model explanations using its [Alibi](https://github.com/SeldonIO/alibi)  Open Source library.

We provide [an example notebook](../examples/explainer_examples.html) showing how to deploy an explainer for Tabular, Text and Image models.

## Creating your explainer

For Alibi explainers that need to be trained you should

 1. Use python 3.6 as the Seldon python wrapper also runs in python 3.6 when it loads your explainer.
 1. Follow the [Alibi docs](https://docs.seldon.io/projects/alibi/en/latest/index.html) for your particular desired explainer. The Seldon Wrapper presently supports: Anchors (Tabular, Text and Image), KernelShap and Integrated Gradients.
 1. Save your explainer to a file called `explainer.dill` using the [dill](https://pypi.org/project/dill/) python package and store on a bucket store or PVC in your cluster. We support gcs, s3 (including Minio) or Azure blob.


## Explain API

For the Seldon Protocol an endpoint path will be exposed for:

```
http://<ingress-gateway>/seldon/<namespace>/<deployment name>/<predictor name>/api/v1.0/explain
```

So for example if you deployed:

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: income
  namespace: seldon
spec:
  name: income
  annotations:
    seldon.io/rest-timeout: "100000"
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/income/model-0.23.2
      name: classifier
    explainer:
      type: AnchorTabular
      modelUri: gs://seldon-models/sklearn/income/explainer-py36-0.5.2
    name: default
    replicas: 1
```

And were port forwarding to Ambassador on localhost:8003 then the API call would be:

```
http://localhost:8003/seldon/seldon/income-explainer/default/api/v1.0/explain
```