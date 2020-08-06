# Prepackaged Model Servers

Seldon provides several prepacked servers you can use to deploy trained models:

- [SKLearn Server](./sklearn.html)
- [XGBoost Server](xgboost.html)
- [Tensorflow Serving](tensorflow.html)
- [MLflow Server](mlflow.html)
- [Custom Servers](custom.md)

For these servers you only need the location of the saved model in a local filestore, Google bucket, S3 bucket, azure or minio. An example manifest with an sklearn server is shown below:

```yaml
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

- Google Cloud Storage (using `gs://`)
- S3-compatible (using `s3://`)
- Minio-compatible (using `s3://`)
- Azure Blob storage (using `https://(.+?).blob.core.windows.net/(.+)`)

The download is handled by an initContainer that runs before your predictor loads. This initContainer image uses our [Storage.py library](https://github.com/SeldonIO/seldon-core/blob/master/python/seldon_core/storage.py) to download the files. However it is also possible for you to override the initContainer with your own custom container to download any files from custom resources.

## Further Customisation for Pre-packaged Model Servers

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

A Kubernetes PersistentVolume [can be used](../examples/pvc-tfjob.html) instead of a bucket using `pvc://`.

Next steps:

- [Worked notebook](../examples/server_examples.html)
- [SKLearn Server](./sklearn.html)
- [XGBoost Server](xgboost.html)
- [Tensorflow Serving](tensorflow.html)
- [MLflow Server](mlflow.html)
- [SKLearn Server with MinIO](../examples/minio-sklearn.html)

You can also build and add your own [custom inference servers](./custom.md),
which can then be used in a similar way as the pre-packaged ones.

If your use case does not fit for a reusable standard server then you can create your own component using our wrappers.

## Handling Credentials

In order to handle credentials you must make available a secret with the environment variables that will be added into the initContainer. For this you need to perform the following actions:

1. Understand which environment variables you need to set
2. Create a secret containing the environment variables
3. Provide the Seldon Core Controller or Seldon Deployment with the name of the secret

### 1. Understand which Environment Variables you need to set

In order to understand what are the environment variables required, you can have a look directly into our [Storage.py library](https://github.com/SeldonIO/seldon-core/blob/master/python/seldon_core/storage.py) that we use in our initContainer.

#### AWS Required Variables

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_ENDPOINT_URL
- USE_SSL

#### Minio Required Variables

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_ENDPOINT_URL
- USE_SSL

#### Azure Required Variables

- AZ_TENANT_ID
- AZ_CLIENT_ID
- AZ_CLIENT_SECRET
- AZ_SUBSCRIPTION_ID

#### Google Cloud Required Variables

Currently for Google Cloud it is required to follow a slightly more complex method given that it requires the secret to be mounted as a file. For this please follow the example at the Google Cloud Section.

If application cretentials are not set, the client will use an Anonymous client.

### 2. Create a secret containing the environment variables

You can now create a secret, below we show what the env variables would look like for the AWS credentials.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-init-container-secret
type: Opaque
data:
  AWS_ACCESS_KEY_ID: XXXX
  AWS_SECRET_ACCESS_KEY: XXXX
  AWS_ENDPOINT_URL: XXXX
  USE_SSL: XXXX
```

It is also possible to create a `Secret` object from the command line:

```
kubectl create secret generic seldon-init-container-secret \
    --from-literal=AWS_ENDPOINT_URL='XXXX' \
    --from-literal=AWS_ACCESS_KEY_ID='XXXX' \
    --from-literal=AWS_SECRET_ACCESS_KEY='XXXX' \
    --from-literal=USE_SSL=false
```

You can read the [documentation of Kubernetes](https://kubernetes.io/docs/concepts/configuration/secret/) to learn more about Kubernetes Secrets.

### 3. Ensure your SeldonDeployment has access to the secret

In order for your SeldonDeployment to know what is the name of the secret, we have to specify the name of the secret we created - in the example above we named the secret `seldon-init-container-secret`.

#### Option 1: Default Seldon Core Manager Controller value

You can set a global default when you install Seldon Core through the Helm chart through the `values.yaml` variable `executor.defaultEnvSecretRefName`. You can see all the variables available in the [Advanced Helm Installation Page](../reference/helm.rst).

```yaml
# ... other variables
predictiveUnit:
  defaultEnvSecretRefName: seldon-init-container-secret
# ... other variables
```

#### Option 2: Override through SeldonDeployment config

It is also possible to provide an override value when you deploy your model using the SeldonDeploymen YAML. You can do this through the `envSecretRefName` value:

```yaml
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
      envSecretRefName: seldon-init-container-secret
      name: classifier
    name: default
    replicas: 1
```

### Examples

#### MinIO running inside same Kubernetes cluster
Assuming that you have MinIO instance running on port `9000` avaible at `minio.minio-system.svc.cluster.local` and you want to reference bucket `mymodel` you would set
```
AWS_ENDPOINT_URL=http://minio.minio-system.svc.cluster.local:9000
```
with `modelUri` being set as `s3://mymodel`.

For full example please see this [notebook](../examples/minio-sklearn.html).

## Adding Credentials for Google Cloud

Currently the Google Credentials require a file to be set up so the process required involves creation of a service account as outlined below.

You can also create a `ServiceAccount` and attach a differently formatted `Secret` to it similar to how kfserving does it. See kfserving documentation [on this topic](https://github.com/kubeflow/kfserving/tree/master/docs/samples/s3). Supported annotation prefix includes `serving.kubeflow.org` and `machinelearning.seldon.io`.

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
