# Prepackaged Model Servers

Seldon provides several prepacked servers you can use to deploy trained models:

- [SKLearn Server](./sklearn.html)
- [XGBoost Server](./xgboost.html)
- [Tensorflow Serving](./tensorflow.html)
- [MLflow Server](./mlflow.html)
- [Custom Servers](./custom.html)

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


## Init Containers

Seldon Core uses [Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) to download model binaries for the prepackaged model servers. We use [kfserving's storage.py library](https://github.com/kubeflow/kfserving/blob/master/python/kfserving/kfserving/storage.py
) for our `Init Containers` by defining
```yaml
storageInitializer:
  image: gcr.io/kfserving/storage-initializer:v0.4.0
```
in our default [helm values](../charts/seldon-core-operator.html#values).
See the [Dockerfile](https://github.com/kubeflow/kfserving/blob/master/python/storage-initializer.Dockerfile
) and its [entrypoint](https://github.com/kubeflow/kfserving/blob/master/python/storage-initializer/scripts/initializer-entrypoint
) for a detailed reference.


### Customizing Init Containers

One can specify a custom `Init Container` globally by overwriting the `storageInitializer.image` helm value as metnioned above.
The `entrypoint` of the `container` must take two arguments:
- first representing the models URI
- second the desired path where binary should be downloaded to

To illustrate how `initContainers` are used by the prepackaged model servers, following effective `SeldonDeployment` with defualt `initContainer` details updated by the mutating webhook is provided:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example
spec:
  name: iris
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: classifier
          volumeMounts:
          - mountPath: /mnt/models
            name: classifier-provision-location
            readOnly: true
        initContainers:
        - name: classifier-model-initializer
          image: gcr.io/kfserving/storage-initializer:v0.4.0
          imagePullPolicy: IfNotPresent
          args:
          - s3://sklearn/iris
          - /mnt/models
          envFrom:
          - secretRef:
              name: seldon-init-container-secret
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /mnt/models
            name: classifier-provision-location
        volumes:
        - emptyDir: {}
          name: classifier-provision-location
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      name: classifier
    name: defaul
```

Key observations:
- our prepackaged model will expect model binaries to be saved into `/mnt/models` path
- default `initContainers` name is constructed from `{predictiveUnitName}-model-initializer`

Currently image used for `initContainers` can only be specified globally via helm values.
The per deployment customisation is explored in this [GitHub issue #2611](https://github.com/SeldonIO/seldon-core/issues/2611).

## Further Customisation for Prepackaged Model Servers

If you want to customize the resources for the server you can add a skeleton `Container` with the same name to your podSpecs, e.g.

```yaml
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
- [XGBoost Server](./xgboost.html)
- [Tensorflow Serving](./tensorflow.html)
- [MLflow Server](./mlflow.html)
- [SKLearn Server with MinIO](../examples/minio-sklearn.html)

You can also build and add your own [custom inference servers](./custom.md),
which can then be used in a similar way as the prepackaged ones.

If your use case does not fit for a reusable standard server then you can create your own component using our wrappers.

## Handling Credentials

In order to handle credentials you must make available a secret with the environment variables that will be added into the `Init Containers`. For this you need to perform the following actions:

1. Understand which environment variables you need to set
2. Create a secret containing the environment variables
3. Provide the Seldon Core Controller or Seldon Deployment with the name of the secret

### 1. Understand which Environment Variables you need to set

In order to understand what are the environment variables required, you can have a look directly into our [Storage.py library](https://github.com/SeldonIO/seldon-core/blob/master/python/seldon_core/storage.py) that we use in our `Init Containers`.

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

```bash
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
```bash
AWS_ENDPOINT_URL=http://minio.minio-system.svc.cluster.local:9000
```
with `modelUri` being set as `s3://mymodel`.

For full example please see this [notebook](../examples/minio-sklearn.html).

## Adding Credentials for Google Cloud

Currently the Google Credentials require a file to be set up so the process required involves creation of a service account as outlined below.

You can also create a `ServiceAccount` and attach a differently formatted `Secret` to it similar to how kfserving does it. See kfserving documentation [on this topic](https://github.com/kubeflow/kfserving/tree/master/docs/samples/s3). Supported annotation prefix includes `serving.kubeflow.org` and `machinelearning.seldon.io`.

For GCP/GKE, go to gcloud console and create a key as json and export as a file. Then create a secret from the file using:

```bash
kubectl create secret generic user-gcp-sa --from-file=gcloud-application-credentials.json=<LOCALFILE>
```

The file in the secret needs to be called `gcloud-application-credentials.json` (the name can be configured in the seldon configmap, visible in `kubectl get cm -n seldon-system seldon-config -o yaml`).

Then create a service account to reference the secret:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: user-gcp-sa
secrets:
  - name: user-gcp-sa
```

This can then be referenced in the SeldonDeployment manifest by setting `serviceAccountName: user-gcp-sa` at the same level as `m̀odelUri` e.g.

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
      serviceAccountName: user-gcp-sa
      name: classifier
    name: default
    replicas: 1
```
