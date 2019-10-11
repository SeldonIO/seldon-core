# Prepackaged Model Servers

Seldon provides several prepacked servers you can use to deploy trained models:

 * [SKLearn Server](./sklearn.html)
 * [XGBoost Server](xgboost.html)
 * [Tensorflow Serving](tensorflow.html)
 * [MLFlow Server](mlflow.html)

For these servers you only need the location of the saved model in a local filestore, Google bucket or S3 bucket. An example manifest with an sklearn server is shown below:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
```

The `modelUri` specifies the bucket containing the saved model, in this case `gs://seldon-models/sklearn/iris`.

`modeluri` supports the following three object storage providers:

  * Google Cloud Storage (using `gs://`)
  * S3-compatible (using `s3://`)
  * Azure Blob storage (using `https://(.+?).blob.core.windows.net/(.+)`)
  
When using S3-compatible object storage provider, you can provide access credential and custom endpoint by creating a `Secret` object:

```
apiVersion: v1
kind: Secret
metadata:
  name: s3-secret
type: Opaque
data:
  AWS_ACCESS_KEY_ID: XXXX
  AWS_SECRET_ACCESS_KEY: XXXX
  S3_ENDPOINT: XXXX
```

You can create a `Secret` object from command line by

```
kubectl create secret generic s3-secret --from-literal=S3_ENDPOINT='XXXX' --from-literal=AWS_ACCESS_KEY_ID='XXXX' --from-literal=AWS_SECRET_ACCESS_KEY='XXXX'
```

and you can [read more](https://kubernetes.io/docs/concepts/configuration/secret/) about interacting with `Secret` object.

And reference the `Secret` using `envSecretRefName` as shown below.

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: s3://seldon-models/sklearn/iris
      envSecretRefName: s3-secret
      name: classifier
    name: default
    replicas: 1
```

Alternatively, you can also create a `ServiceAccount` and attach a differently formatted `Secret` to it similar to how kfserving does it.  See kfserving documentation [on this topic](https://github.com/kubeflow/kfserving/tree/master/docs/samples/s3).  Supported annotation prefix includes `serving.kubeflow.org` and `machinelearning.seldon.io`.

For GCP/GKE, go to gcloud console and create a key as json and export as a file. Then create a secret from the file using:
```
kubectl create secret generic user-gcp-sa --from-file=gcloud-application-credentials.json=<LOCALFILE>
```
The file in the secret needs to be called `gcloud-application-credentials.json` (the name can be configured in the seldon configmap, visible in `kubectl get cm -n seldon-system seldon-config -o yaml`).

Then create a service account to reference the secret:
```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: user-gcp-sa
secrets:
  - name: user-gcp-sa
```

This can then be referenced in the SeldonDeployment manifest by setting `serviceAccountName: user-gcp-sa` at the same level as `m̀odelUri` e.g.


```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      serviceAccountName: user-gcp-sa
      name: classifier
    name: default
    replicas: 1
```

If you want to customize the resources for the server you can add a skeleton `Container` with the same name to your podSpecs, e.g.

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: classifier
          resources:
            requests:
              memory: 50Mi
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1

```

The image name and other details will be added when this is deployed automatically.

Next steps:

   * [Worked notebook](../examples/server_examples.html)
   * [SKLearn Server](./sklearn.html)
   * [XGBoost Server](xgboost.html)
   * [Tensorflow Serving](tensorflow.html)
   * [MLflow Server](mlflow.html)

If your use case does not fall into the above standard servers then you can create your own component using our wrappers.

