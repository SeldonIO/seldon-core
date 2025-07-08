# Upgrade to rclone-based Storage Initializer - secret format intuition

In this documentation page we provide an example upgrade path from kfserving-based to rclone-based storage initializer. This is required due to the fact that secret format expected by these two storage initializers is different. 

Storage initializers are used by Seldon's pre-packaged model servers to download models binaries. 
As it is explained in the [SC 1.8 upgrading notes](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html#upgrading-to-1-8) the [seldonio/rclone-storage-initializer](https://github.com/SeldonIO/seldon-core/tree/master/components/rclone-storage-initializer) became default storage initializer in v1.8.0.

In this tutorial we will show how to upgrade your configuration to new Storage Initializer with focus on getting the new format of a required secret right.

Read more:
- [Prepackaged Model Servers documentation page](https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html)
- [SC 1.8 upgrading notes](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html#upgrading-to-1-8)
- [Example upgrade path to use rclone-based storage initializer globally](https://docs.seldon.io/projects/seldon-core/en/latest/examples/global-rclone-upgrade.html)

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * mc client
 * curl

## Steps in this tutorial

 * Copy iris model from GCS into in-cluster minio and configure old-style storage initializer secret
 * Deploy SKlearn Pre-Packaged server using kfserving storage initializer
 * Discuss upgrading procedure and tips how to test new secret format
 * Deploy Pre-packaged model server using rclone storage initializer
 
## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). 

## Setup MinIO

Use the provided [notebook](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html) to install Minio in your cluster and configure `mc` CLI tool. 

## Copy iris model into local MinIO


```bash
%%bash
mc config host add gcs https://storage.googleapis.com "" "" 

mc mb minio-seldon/sklearn/iris/ -p
mc cp gcs/seldon-models/sklearn/iris/model.joblib minio-seldon/sklearn/iris/
mc cp gcs/seldon-models/sklearn/iris/metadata.yaml minio-seldon/sklearn/iris/
```


```bash
%%bash
mc ls minio-seldon/sklearn/iris/
```

## Deploy SKLearn Server with kfserving-storage-initializer

First we deploy the model using kfserving-storage-initializer. This is using the default Storage Initializer for pre Seldon Core v1.8.0.


```python
%%writefile sklearn-iris-kfserving.yaml

apiVersion: v1
kind: Secret
metadata:
  name: seldon-kfserving-secret
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: minioadmin
  AWS_SECRET_ACCESS_KEY: minioadmin
  AWS_ENDPOINT_URL: http://minio.minio-system.svc.cluster.local:9000
  USE_SSL: "false"
    
---
    
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-iris-kfserving
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      envSecretRefName: seldon-kfserving-secret
      storageInitializerImage: kfserving/storage-initializer:v0.6.1
```

    Overwriting sklearn-iris-kfserving.yaml



```python
!kubectl apply -f sklearn-iris-kfserving.yaml
```

    secret/seldon-kfserving-secret configured
    seldondeployment.machinelearning.seldon.io/sklearn-iris-kfserving configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-kfserving -o jsonpath='{.items[0].metadata.name}')
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-kfserving/api/v1.0/predictions  | jq .
```

## Preparing rclone-compatible secret

The [rclone](https://rclone.org/)-based storage initializer expects one to define a new secret. General documentation credentials hadling can be found [here](https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html#handling-credentials) with constantly updated examples of tested configurations.

If we do not have yet an example for Cloud Storage solution that you are using, please, consult the relevant page on [RClone documentation](https://rclone.org/#providers).

### Preparing seldon-rclone-secret

Knowing format of required format of the secret we can create it now


```python
%%writefile seldon-rclone-secret.yaml
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


```python
!kubectl apply -f seldon-rclone-secret.yaml
```

### Testing seldon-rclone-secret

Before deploying SKLearn server one can test directly using the rclone-storage-initializer image


```python
%%writefile rclone-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: rclone-pod
spec:
  containers:
  - name: rclone
    image: seldonio/rclone-storage-initializer:1.19.0-dev
    command: [ "/bin/sh", "-c", "--", "sleep 3600"]
    envFrom:
    - secretRef:
        name: seldon-rclone-secret
```


```python
!kubectl apply -f rclone-pod.yaml
```


```python
! kubectl exec -it rclone-pod -- rclone ls s3:sklearn
```


```python
! kubectl exec -it rclone-pod -- rclone copy s3:sklearn .
```


```python
! kubectl exec -it rclone-pod -- sh -c "ls iris/"
```

Once we tested that secret format is correct we can delete the pod


```python
!kubectl delete -f rclone-pod.yaml
```

## Deploy SKLearn Server with rclone-storage-initializer


```python
%%writefile sklearn-iris-rclone.yaml

apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-iris-rclone
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      envSecretRefName: seldon-rclone-secret
      storageInitializerImage: seldonio/rclone-storage-initializer:1.19.0-dev
```


```python
!kubectl apply -f sklearn-iris-rclone.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-rclone -o jsonpath='{.items[0].metadata.name}')
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-rclone/api/v1.0/predictions  | jq .
```

## Cleanup


```bash
%%bash
kubectl delete -f sklearn-iris-rclone.yaml
kubectl delete -f sklearn-iris-kfserving.yaml
```
