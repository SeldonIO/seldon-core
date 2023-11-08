## Seldon V2 Kubernetes Load Balance Examples

Examples showing http and grpc requests load balanced across 3 MLServer replicas.

## Setup

```bash
helm upgrade --install seldon-core-v2-crds  ../k8s/helm-charts/seldon-core-v2-crds -n seldon-mesh
```

```
Release "seldon-core-v2-crds" does not exist. Installing it now.
NAME: seldon-core-v2-crds
LAST DEPLOYED: Fri Jul 21 18:36:15 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm upgrade --install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh
```

```
Release "seldon-v2" does not exist. Installing it now.
NAME: seldon-v2
LAST DEPLOYED: Fri Jul 21 18:36:19 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n seldon-mesh --wait
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Fri Jul 21 18:36:21 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-servers ../k8s/helm-charts/seldon-core-v2-servers  -n seldon-mesh --set mlserver.replicas=3 --wait
```

```yaml
NAME: seldon-v2-servers
LAST DEPLOYED: Fri Jul 21 18:36:53 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

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
'172.18.255.2'

```

### Model

```bash
cat ./models/iris-multi-replica.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki
  replicas: 3

```

```bash
kubectl create -f ./models/iris-multi-replica.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris condition met

```

```bash
kubectl get model iris -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-07-21T17:38:18Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-07-21T17:38:18Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 3
}

```

```bash
seldon model infer iris --inference-host ${MESH_IP}:80 -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::100]

```

```bash
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP}:80 -i 100 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```

```
Success: map[:iris_1::100]

```

```python
import tritonclient.grpc as grpcclient
import numpy as np

grpc_triton_client = grpcclient.InferenceServerClient(
    url=f"{MESH_IP}:80",
    verbose=False,
)
```

```python
model_name = "iris"
headers = {"seldon-model": model_name}

inputs = [
    grpcclient.InferInput("predict", (1, 4), "FP64"),
]
inputs[0].set_data_from_numpy(np.array([[1, 2, 3, 4]]).astype("float64"))

outputs = [grpcclient.InferRequestedOutput("predict")]

for idx in range(0,100):
    result = grpc_triton_client.infer(model_name, inputs, outputs=outputs, headers=headers)
    result.as_numpy("predict")
```

```bash
kubectl logs mlserver-0 -c mlserver -n seldon-mesh | grep inference.GRPCInferenceService/ModelInfer | wc -l
kubectl logs mlserver-1 -c mlserver -n seldon-mesh | grep inference.GRPCInferenceService/ModelInfer | wc -l
kubectl logs mlserver-2 -c mlserver -n seldon-mesh | grep inference.GRPCInferenceService/ModelInfer | wc -l
```

```
69
72
59

```

```bash
kubectl delete -f ./models/iris-multi-replica.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted

```

```python

```
