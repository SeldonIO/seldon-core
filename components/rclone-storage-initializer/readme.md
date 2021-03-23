# Rclone Based Storage Initializer



## Rclone configuration

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

### Helm values

To configure rclone-based storage initializer with your Seldon Core installation create
the `seldon-rclone-secret` using one of the configurations bellow and use following helm values:


```yaml
storageInitializer:
  image: seldonio/rclone-storage-initializer:1.8.0-dev

predictiveUnit:
  defaultEnvSecretRefName: seldon-rclone-secret
```


### Example for public GCS configuration

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



### Example S3 with IAM roles configuration

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
