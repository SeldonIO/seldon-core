## Tabular Income Classifier Kubernetes Test

To run this notebook you need the inference data. This can be acquired in two ways:

  * Run train.ipynb
  * `gsutil cp -R gs://seldon-models/scv2/examples/income/infer-data .`

```python
import os
os.environ["NAMESPACE"] = "seldon-mesh"
```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```

```
'172.19.255.1'

```

```python
import numpy as np
import json
import requests
```

```python
with open('./infer-data/test.npy', 'rb') as f:
    x_ref = np.load(f)
    x_h1 = np.load(f)
    y_ref = np.load(f)
    x_outlier = np.load(f)
```

```python
reqJson = json.loads('{"inputs":[{"name":"input_1","data":[],"datatype":"FP32","shape":[]}]}')
url = "http://"+MESH_IP+"/v2/models/model/infer"
url
```

```
'http://172.19.255.1/v2/models/model/infer'

```

```python
def infer(resourceName: str, batchSz: int, requestType: str):
    if requestType == "outlier":
        rows = x_outlier[0:0+batchSz]
    elif requestType == "drift":
        rows = x_h1[0:0+batchSz]
    else:
        rows = x_ref[0:0+batchSz]
    reqJson["inputs"][0]["data"] = rows.flatten().tolist()
    reqJson["inputs"][0]["shape"] = [batchSz, rows.shape[1]]
    headers = {"Content-Type": "application/json", "seldon-model":resourceName}
    response_raw = requests.post(url, json=reqJson, headers=headers)
    print(response_raw)
    print(response_raw.json())
```

### Pipeline with model, drift detector and outlier detector

```bash
kubectl create -f ../../models/income.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income condition met

```

```bash
cat ../../pipelines/income-v1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: income-prod
spec:
  steps:
    - name: income
  output:
    steps:
    - income

```

```bash
kubectl apply -f ../../pipelines/income-v1.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod configured

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod condition met

```

```python
batchSz=1
print(y_ref[0:batchSz])
infer("income-prod.pipeline",batchSz,"normal")
```

```
[0]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0], 'name': 'predict', 'shape': [1], 'datatype': 'INT64'}]}

```

```bash
kubectl create -f ../../models/income-drift.yaml -n ${NAMESPACE}
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income condition met
model.mlops.seldon.io/income-drift condition met

```

```bash
cat ../../pipelines/income-v2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: income-prod
spec:
  steps:
    - name: income
    - name: income-drift
      batch:
        size: 20
  output:
    steps:
    - income

```

```bash
kubectl apply -f ../../pipelines/income-v2.yaml -n ${NAMESPACE}
```

```yaml
Warning: resource pipelines/income-prod is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
pipeline.mlops.seldon.io/income-prod configured

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod condition met

```

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-prod.pipeline",batchSz,"normal")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}]}

```

```bash
kubectl create -f ../../models/income-preprocess.yaml -n ${NAMESPACE}
kubectl create -f ../../models/income-outlier.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income-preprocess created
model.mlops.seldon.io/income-outlier created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income condition met
model.mlops.seldon.io/income-drift condition met
model.mlops.seldon.io/income-outlier condition met
model.mlops.seldon.io/income-preprocess condition met

```

```bash
kubectl apply -f ../../pipelines/income-v3.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod configured

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod condition met

```

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-prod.pipeline",batchSz,"outlier")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 1, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-prod.pipeline",batchSz,"normal")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

### Cleanup

```bash
kubectl delete -f ../../pipelines/income-v3.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io "income-prod" deleted

```

```bash
kubectl delete -f ../../models/income.yaml -n ${NAMESPACE}
kubectl delete -f ../../models/income-drift.yaml -n ${NAMESPACE}
kubectl delete -f ../../models/income-preprocess.yaml -n ${NAMESPACE}
kubectl delete -f ../../models/income-outlier.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "income" deleted
model.mlops.seldon.io "income-drift" deleted
model.mlops.seldon.io "income-preprocess" deleted
model.mlops.seldon.io "income-outlier" deleted

```

```python

```
