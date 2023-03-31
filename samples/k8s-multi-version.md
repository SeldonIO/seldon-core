## Seldon V2 Kubernetes Multi Version Artifact Examples

We have a Triton model that has two version folders

Model 1 adds 10 to input, Model 2 multiples by 10 the input. The structure of the artifact repo is shown below:

```
config.pbtxt
1/model.py <add 10>
2/model.py <mul 10>

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
'172.19.255.1'

```

### Model

```bash
cat ./models/multi-version-1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: math
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/multi-version"
  artifactVersion: 1
  requirements:
  - triton
  - python

```

```bash
kubectl apply -f ./models/multi-version-1.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/math created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/math condition met

```

```bash
seldon model infer math --inference-mode grpc --inference-host ${MESH_IP}:80 \
  '{"model_name":"math","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "modelName": "math_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```

```bash
cat ./models/multi-version-2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: math
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/multi-version"
  artifactVersion: 2
  requirements:
  - triton
  - python

```

```bash
kubectl apply -f ./models/multi-version-2.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/math configured

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/math condition met

```

```bash
seldon model infer math --inference-mode grpc --inference-host ${MESH_IP}:80 \
  '{"model_name":"math","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "modelName": "math_2",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}

```

```bash
kubectl delete -f ./models/multi-version-1.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "math" deleted

```

```python

```
