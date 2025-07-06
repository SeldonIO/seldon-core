# Basic Examples for SKlearn Prepackaged Server with MinIO


## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

## Setup MinIO

Use the provided [notebook](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html) to install Minio in your cluster and configure `mc` CLI tool. 
Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html).

## Copy iris model into local MinIO


```bash
%%bash
mc config host add gcs https://storage.googleapis.com "" "" 

mc mb minio-seldon/iris -p
mc cp gcs/seldon-models/sklearn/iris/model.joblib minio-seldon/iris/
mc cp gcs/seldon-models/sklearn/iris/metadata.yaml minio-seldon/iris/
```

    Added `gcs` successfully.
    Bucket created successfully `minio-seldon/iris`.
    `gcs/seldon-models/sklearn/iris/model.joblib` -> `minio-seldon/iris/model.joblib`
    Total: 0 B, Transferred: 1.06 KiB, Speed: 2.16 KiB/s
    `gcs/seldon-models/sklearn/iris/metadata.yaml` -> `minio-seldon/iris/metadata.yaml`
    Total: 0 B, Transferred: 162 B, Speed: 335 B/s


## Modify model metadata (optional)


```bash
%%bash
mc cat minio-seldon/iris/metadata.yaml
```

    
    name: iris
    versions: [iris/v1]
    platform: sklearn
    inputs:
    - datatype: BYTES
      name: input
      shape: [ 4 ]
    outputs:
    - datatype: BYTES
      name: output
      shape: [ 3 ]



```python
%%writefile metadata.yaml

name: iris
versions: [iris/v1-updated]
platform: sklearn
inputs:
- datatype: BYTES
  name: input
  shape: [ 1, 4 ]
outputs:
- datatype: BYTES
  name: output
  shape: [ 3 ]
```

    Overwriting metadata.yaml



```bash
%%bash
mc cp metadata.yaml minio-seldon/iris/
```

    `metadata.yaml` -> `minio-seldon/iris/metadata.yaml`
    Total: 0 B, Transferred: 173 B, Speed: 25.02 KiB/s


## Deploy sklearn server


```python
%%writefile secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: seldon-init-container-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: minio
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: minioadmin
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: minioadmin
  RCLONE_CONFIG_S3_ENDPOINT: http://minio.minio-system.svc.cluster.local:9000
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
```

    Overwriting secret.yaml



```python
!kubectl apply -f secret.yaml
```

    secret/seldon-init-container-secret created



```python
%%writefile deploy.yaml

apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: minio-sklearn
spec:
  name: iris
  predictors:
  - componentSpecs:
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: s3://iris
      envSecretRefName: seldon-init-container-secret
      name: classifier
    name: default
    replicas: 1
```

    Overwriting deploy.yaml



```python
!kubectl apply -f deploy.yaml
```

    seldondeployment.machinelearning.seldon.io/minio-sklearn created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=minio-sklearn -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "minio-sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "minio-sklearn-default-0-classifier" successfully rolled out


## Test deployment

### Test prediction


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/minio-sklearn/api/v1.0/predictions  | jq .
```

    {
      "data": {
        "names": [
          "t:0",
          "t:1",
          "t:2"
        ],
        "ndarray": [
          [
            0.9548873249364169,
            0.04505474761561406,
            5.7927447968952436e-05
          ]
        ]
      },
      "meta": {}
    }


### Test model metadata (optional)


```bash
%%bash
curl -s http://localhost:8003/seldon/seldon/minio-sklearn/api/v1.0/metadata/classifier | jq .
```

    {
      "inputs": [
        {
          "datatype": "BYTES",
          "name": "input",
          "shape": [
            1,
            4
          ]
        }
      ],
      "name": "iris",
      "outputs": [
        {
          "datatype": "BYTES",
          "name": "output",
          "shape": [
            3
          ]
        }
      ],
      "platform": "sklearn",
      "versions": [
        "iris/v1-updated"
      ]
    }


### Test for CI


```python
import json

data = !curl -s -X POST -H 'Content-Type: application/json' -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' http://localhost:8003/seldon/seldon/minio-sklearn/api/v1.0/predictions
data = json.loads(data[0])

assert data == {
    "data": {
        "names": ["t:0", "t:1", "t:2"],
        "ndarray": [[0.9548873249364169, 0.04505474761561406, 5.7927447968952436e-05]],
    },
    "meta": {},
}
```


```python
import json

meta = !curl -s http://localhost:8003/seldon/seldon/minio-sklearn/api/v1.0/metadata/classifier
meta = json.loads(meta[0])

assert data == {
    "data": {
        "names": ["t:0", "t:1", "t:2"],
        "ndarray": [[0.9548873249364169, 0.04505474761561406, 5.7927447968952436e-05]],
    },
    "meta": {},
}
```

## Cleanup


```python
!kubectl delete -f deploy.yaml
```

    seldondeployment.machinelearning.seldon.io "minio-sklearn" deleted

