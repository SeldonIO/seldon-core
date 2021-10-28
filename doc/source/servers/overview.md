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
  name: sklearn-iris
spec:
  predictors:
    - name: default
      replicas: 1
      graph:
        name: classifier
        implementation: SKLEARN_SERVER
        modelUri: gs://seldon-models/v1.12.0-dev/sklearn/iris
```

By default only public models published to Google Cloud Storage will be accessible.
See below notes on how to configure credentials for AWS S3, Minio and other storage solutions.


## Init Containers

Seldon Core uses [Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) to download model binaries for the prepackaged model servers. We use [rclone](https://rclone.org/)-based [storage initailizer](https://github.com/SeldonIO/seldon-core/tree/master/components/rclone-storage-initializer
) for our `Init Containers` by defining

```yaml
storageInitializer:
  image: seldonio/rclone-storage-initializer:1.12.0-dev
```
in our default [helm values](../charts/seldon-core-operator.html#values).
See the [Dockerfile](https://github.com/SeldonIO/seldon-core/blob/master/components/rclone-storage-initializer/Dockerfile
) for a detailed reference.
You can overwrite this value to specify another default `initContainer`. See details on requirements bellow

Secrets are injected into the init containers as environmental variables from kubernetes `secrets`.
The default secret name can be defined by setting following [helm value](../charts/seldon-core-operator.html#values)

```yaml
predictiveUnit:
  defaultEnvSecretRefName: ""
```

Note: prior to Seldon Core 1.8 we were using `kfserving/storage-initializer`, see [these](./kfserving-storage-initializer.md) notes if you wish to keep using it.


### Customizing Init Containers

You can specify a custom `initContainer` image and default `secret` **globally** by overwriting the helm values specified in the previous section.

To illustrate how `initContainers` are used by the prepackaged model servers, consider a following Seldon Deployment with `volumes`, `volumeMounts` and `initContainers` equivalent to ones that would be injected by the `Seldon Core Operator` if this was prepackaged model server:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-iris
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      type: MODEL

    componentSpecs:
    - spec:
        volumes:
        - name: classifier-provision-location
          emptyDir: {}

        initContainers:
        - name: classifier-model-initializer
          image: seldonio/rclone-storage-initializer:1.12.0-dev
          imagePullPolicy: IfNotPresent
          args:
            - "s3://sklearn/iris"
            - "/mnt/models"

          volumeMounts:
          - mountPath: /mnt/models
            name: classifier-provision-location

          envFrom:
          - secretRef:
              name: seldon-init-container-secret

        containers:
        - name: classifier
          image: seldonio/sklearnserver:1.8.0-dev

          volumeMounts:
          - mountPath: /mnt/models
            name: classifier-provision-location
            readOnly: true

          env:
          - name: PREDICTIVE_UNIT_PARAMETERS
            value: '[{"name":"model_uri","value":"/mnt/models","type":"STRING"}]'
```

Key observations:
- Our prepackaged model will expect model binaries to be saved into `/mnt/models` path
- Default `initContainers` name is constructed from `{predictiveUnitName}-model-initializer`
- The `entrypoint` of the `container` must take two arguments:
  - First representing the models URI
  - Second the desired path where binary should be downloaded to
- If user would to provide their own `initContainer` which name matches the above pattern it would be used as provided

This is equivalent to the following `sklearn-iris` Seldon Deployment. As we can see using prepackaged model servers allow one to avoid defining boilerplate and make definition much cleaner:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-sklearn
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      storageInitializerImage: seldonio/rclone-storage-initializer:1.12.0-dev  # Specify custom image here
      envSecretRefName: seldon-init-container-secret                          # Specify custom secret here
```
Note that image and secret used by Storage Initializer can be customised per-deployment.

See our [example](../examples/custom_init_container.html) that explains in details how init containers are used and how to write a custom one using [rclone](https://rclone.org/) for cloud storage operations as an example.

## Further Customisation for Prepackaged Model Servers

If you want to customize the resources for the server you can add a skeleton `Container` with the same name to your podSpecs, e.g.

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn-iris
spec:
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: classifier
          resources:
            requests:
              memory: 50Mi
    name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.12.0-dev/sklearn/iris
```

The image name and other details will be added when this is deployed automatically.

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

### General notes

Rclone remotes can be configured using the environmental variables:
```
RCLONE_CONFIG_<remote name>_<config variable>: <config value>
```

Note: multiple remotes can be configured simultaneously.

Once the remote is configured the modelUri that is compatible with `rclone` takes form
```
modelUri: <remote>:<bucket name>
```
for example `modelUri: s3:sklearn/iris`.

Note: Rclone will remove the leading slashes for buckets so this is equivalent to `s3://sklearn/iris`.

Below you will find a few example configurations. For other cloud solutions, please, consult great [documentation](https://rclone.org/) of the rclone project.


### Example for public GCS configuration

Note: this is configured by default in the `seldonio/rclone-storage-initializer` image.

Reference: [rclone documentation](https://rclone.org/googlecloudstorage/).


```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_GS_TYPE: google cloud storage
  RCLONE_CONFIG_GS_ANONYMOUS: "true"
```

Example deployment

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: rclone-sklearn-gs
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: gs:seldon-models/sklearn/iris
      envSecretRefName: seldon-rclone-secret
```


### Example minio configuration

Reference: [rclone documentation](https://rclone.org/s3/#minio)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: minio
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: minioadmin
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: minioadmin
  RCLONE_CONFIG_S3_ENDPOINT: http://minio.minio-system.svc.cluster.local:9000
```

### Example AWS S3 with access key and secret

Reference: [rclone documentation](https://rclone.org/s3/#amazon-s3)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: aws
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: "<your AWS_ACCESS_KEY_ID here>"
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: "<your AWS_SECRET_ACCESS_KEY here>"
```


### Example AWS S3 with IAM roles configuration

Reference: [rclone documentation](https://rclone.org/s3/#amazon-s3)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: aws
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: ""
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: ""
  RCLONE_CONFIG_S3_ENV_AUTH: "true"
```


### Example for GCP/GKE

Reference: [rclone documentation](https://rclone.org/googlecloudstorage/)

For GCP/GKE, you will need create a service-account key and have it as local `json` file.
First make sure that you have `[SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com` service account created in the gcloud console that have sufficient permissions to access the bucket with your models (i.e. `Storage Object Admin`).

Now, generate `keys` locally using the `gcloud` tool
```bash
gcloud iam service-accounts keys create gcloud-application-credentials.json --iam-account [SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com
```

Now using the content of locally saved `gcloud-application-credentials.json` file create a secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_GCS_TYPE: google cloud storage
  RCLONE_CONFIG_GCS_ANONYMOUS: "false"
  RCLONE_CONFIG_GCS_SERVICE_ACCOUNT_CREDENTIALS: '{"type":"service_account", ... <rest of gcloud-application-credentials.json>}'
```

Note: remote name is `gcs` here so urls would take form similar to `gcs:<your bucket>`.


### Directly from PVC

You are able to make models available directly from PVCs instead of object stores. This may be desirable if you have a lot of very large files and you want to avoid uploading/downloading, for example through NFS drives.

The way in which you are able to specify the PVC is using the `modelUri` with the following format below. One thing to take into consideration is the permissions in the files as the containers will have their respective `runAsUser` parameters.

```
...
    modelUri: pvc://<pvc-name>/<path>
```
