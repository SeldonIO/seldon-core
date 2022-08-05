# KFserving Storage Initializer (Deprecated)

Prior to Seldon Core 1.8 seldon core was using by default `kfserving/storage-initializer` for its pre-packaged model servers. This can be still used by configuring a following helm value:


```yaml
storageInitializer:
  image: kfserving/storage-initializer:v0.6.1
```

> :warning: **NOTE:** Current default storage initializer is `seldonio/rclone-storage-initializer:1.14.1` is described [here](./overview.md).


When `kfserving/storage-initializer` is used `modeluri` supports the following four object storage providers:

- Google Cloud Storage (using `gs://`)
- S3-compatible (using `s3://`)
- Minio-compatible (using `s3://`)
- Azure Blob storage (using `https://(.+?).blob.core.windows.net/(.+)`)

A Kubernetes PersistentVolume [can be used](../examples/pvc-tfjob.html) instead of a bucket using `pvc://`.


## Handling Credentials

In order to handle credentials you must make available a secret with the environment variables that will be added into the `Init Containers`. For this you need to perform the following actions:

1. Understand which environment variables you need to set
2. Create a secret containing the environment variables
3. Provide the Seldon Core Controller or Seldon Deployment with the name of the secret

### 1. Understand which Environment Variables you need to set

In order to understand what are the environment variables required, you can have a look directly into our [Storage.py library](https://github.com/SeldonIO/seldon-core/blob/master/python/seldon_core/storage.py) that we use in our `Init Containers`.

#### AWS Required Variables

  RCLONE_CONFIG_S3_PROVIDER: aws
- RCLONE_CONFIG_S3_ACCESS_KEY_ID
- RCLONE_CONFIG_S3_SECRET_ACCESS_KEY
- RCLONE_CONFIG_S3_ENDPOINT

#### Minio Required Variables

  RCLONE_CONFIG_S3_PROVIDER: minio
- RCLONE_CONFIG_S3_ACCESS_KEY_ID
- RCLONE_CONFIG_S3_SECRET_ACCESS_KEY
- RCLONE_CONFIG_S3_ENDPOINT
- RCLONE_CONFIG_S3_ENV_AUTH

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
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: aws
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: "<your AWS_ACCESS_KEY_ID here>"
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: "<your AWS_SECRET_ACCESS_KEY here>"
  RCLONE_CONFIG_S3_ENDPOINT: "<your S3 endpoint here>"
```

It is also possible to create a `Secret` object from the command line:

```bash
kubectl create secret generic seldon-init-container-secret \
    --from-literal=RCLONE_CONFIG_S3_ENDPOINT='XXXX' \
    --from-literal=RCLONE_CONFIG_S3_ACCESS_KEY_ID='XXXX' \
    --from-literal=RCLONE_CONFIG_S3_SECRET_ACCESS_KEY='XXXX' \
    --from-literal=RCLONE_CONFIG_S3_PROVIDER='aws' \
    --from-literal=RCLONE_CONFIG_S3_TYPE='s3' \
    --from-literal=RCLONE_CONFIG_S3_ENV_AUTH=false
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
RCLONE_CONFIG_S3_ENDPOINT=http://minio.minio-system.svc.cluster.local:9000
```
with `modelUri` being set as `s3://mymodel`.

For full example please see this [notebook](../examples/minio-sklearn.html).

## Adding Credentials for Google Cloud

Currently the Google Credentials require a file to be set up so the process required involves creation of a service account as outlined below.

You can also create a `ServiceAccount` and attach a differently formatted `Secret` to it similar to how kfserving does it. See kfserving documentation [on this topic](https://github.com/kubeflow/kfserving/blob/master/docs/samples/storage/s3/README.md). Supported annotation prefix includes `serving.kubeflow.org` and `machinelearning.seldon.io`.

For GCP/GKE, you will need create a service-account key and have it as local `json` file.
First make sure that you have `[SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com` service account created in the gcloud console that have sufficient permissions to access the bucket with your models (i.e. `Storage Object Admin`).

Now, generate `keys` locally using the `gcloud` tool
```bash
gcloud iam service-accounts keys create gcloud-application-credentials.json --iam-account [SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com
```

Once you have `gcloud-application-credentials.json` file locally create the k8s `secret` with:
```bash
kubectl create secret generic user-gcp-sa --from-file=gcloud-application-credentials.json=<LOCALFILE JSON FILE>
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
      modelUri: gs://seldon-models/v1.14.1/sklearn/iris
      serviceAccountName: user-gcp-sa
      name: classifier
    name: default
    replicas: 1
```
