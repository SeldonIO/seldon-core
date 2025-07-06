# Example model explanations with Seldon  and v2 Protocol - Incubating

In this notebook we will show examples that illustrate how to explain models using [MLServer] (https://github.com/SeldonIO/MLServer).

MLServer is a Python server for your machine learning models through a REST and gRPC interface, fully compliant with KFServing's v2 Dataplane spec. 

## Running this Notebook

 This should install the required package dependencies, if not please also install:
 
- install and configure `mc`, follow the relevant section in this [link](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html)

- run this jupyter notebook in conda environment
```bash
$ conda create --name python3.8-example python=3.8 -y
$ conda activate python3.8-example
$ pip install jupyter
$ jupyter notebook
```

- instal requirements
 - [alibi package](https://pypi.org/project/alibi/)
 - `sklearn`


```python
!pip install sklearn alibi
```

## Setup Seldon Core

Follow the instructions to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core).

 Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`

### Setup MinIO

Use the provided [notebook](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html) to install Minio in your cluster and configure `mc` CLI tool. 
Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html).

## Train `iris` model using `sklearn`


```python
import os
import shutil

from joblib import dump
from sklearn.datasets import load_iris
from sklearn.linear_model import LogisticRegression
```

### Train model


```python
iris_data = load_iris()

clf = LogisticRegression(solver="liblinear", multi_class="ovr")
clf.fit(iris_data.data, iris_data.target)
```

### Save model


```python
modelpath = "/tmp/sklearn_iris"
if os.path.exists(modelpath):
    shutil.rmtree(modelpath)
os.makedirs(modelpath)
modelfile = os.path.join(modelpath, "model.joblib")

dump(clf, modelfile)
```

## Create `AnchorTabular` explainer 

### Create explainer artifact


```python
from alibi.explainers import AnchorTabular

explainer = AnchorTabular(clf.predict, feature_names=iris_data.feature_names)
explainer.fit(iris_data.data, disc_perc=(25, 50, 75))
```

### Save explainer


```python
explainerpath = "/tmp/iris_anchor_tabular_explainer_v2"
if os.path.exists(explainerpath):
    shutil.rmtree(explainerpath)
explainer.save(explainerpath)
```

## Install dependencies to pack the enviornment for deployment


```python
pip install conda-pack mlserver==0.6.0.dev2 mlserver-alibi-explain==0.6.0.dev2
```

## Pack enviornment


```python
import conda_pack

env_file_path = os.path.join(explainerpath, "environment.tar.gz")
conda_pack.pack(
    output=str(env_file_path),
    force=True,
    verbose=True,
    ignore_editable_packages=False,
    ignore_missing_files=True,
)
```

## Copy artifacts to object store (`minio`)

### Configure `mc` to access the minio service in the local kind cluster
note: make sure that minio ip is reflected properly below, run:
- `kubectl get service -n minio-system`
- `mc config host add minio-seldon [ip] minioadmin minioadmin`


```python
target_bucket = "minio-seldon/models"
os.system(f"mc rb --force {target_bucket}")
os.system(f"mc mb {target_bucket}")
os.system(f"mc cp --recursive {modelpath} {target_bucket}")
os.system(f"mc cp --recursive {explainerpath} {target_bucket}")
```

## Deploy to local `kind` cluster

### Create deployment CRD


```python
%%writefile iris-with-explainer-v2.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris
spec:
  protocol: kfserving  # Activate v2 protocol / mlserver usage
  name: iris
  annotations:
    seldon.io/rest-timeout: "100000"
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: s3://models/sklearn_iris
      envSecretRefName: seldon-rclone-secret
      name: classifier
    explainer:
      type: AnchorTabular
      modelUri: s3://models/iris_anchor_tabular_explainer_v2
      envSecretRefName: seldon-rclone-secret
    name: default
    replicas: 1
```

### Deploy


```python
!kubectl apply -f iris-with-explainer-v2.yaml
```

    seldondeployment.machinelearning.seldon.io/iris created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[0].metadata.name}')
```

### Test explainer


```python
!pip install numpy requests
```


```python
import json

import numpy as np
import requests
```


```python
endpoint = "http://localhost:8003/seldon/seldon/iris-explainer/default/v2/models/iris-default-explainer/infer"

test_data = np.array([[5.964, 4.006, 2.081, 1.031]])

inference_request = {
    "parameters": {"content_type": "np"},
    "inputs": [
        {
            "name": "explain",
            "shape": test_data.shape,
            "datatype": "FP32",
            "data": test_data.tolist(),
            "parameters": {"content_type": "np"},
        },
    ],
}
response = requests.post(endpoint, json=inference_request)

explanation = json.loads(response.json()["outputs"][0]["data"])
print("Anchor: %s" % (" AND ".join(explanation["data"]["anchor"])))
print("Precision: %.2f" % explanation["data"]["precision"])
print("Coverage: %.2f" % explanation["data"]["coverage"])
```


```python
!kubectl delete -f iris-with-explainer-v2.yaml
```


```python

```
