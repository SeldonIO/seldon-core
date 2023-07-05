## Seldon V2 Kubernetes Examples

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

## Custom Server with Capabilities

The `capabilities` field replaces the capabilities from the ServerConfig.

```bash
cat ./servers/custom-mlserver-capabilities.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-134
spec:
  serverConfig: mlserver
  capabilities:
  - mlserver-1.3.4
  podSpec:
    containers:
    - image: seldonio/mlserver:1.3.4
      name: mlserver

```

```bash
kubectl create -f ./servers/custom-mlserver-capabilities.yaml -n ${NAMESPACE}
```

```
server.mlops.seldon.io/mlserver-134 created

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ${NAMESPACE}
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/mlserver-134 condition met
server.mlops.seldon.io/triton condition met

```

```bash
cat ./models/iris-custom-requirements.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - mlserver-1.3.4

```

```bash
kubectl create -f ./models/iris-custom-requirements.yaml -n ${NAMESPACE}
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
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "057ae95c-e6bc-4f57-babf-0817ff171729",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

```bash
kubectl delete -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted

```

```bash
kubectl delete -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```

```
server.mlops.seldon.io "mlserver-134" deleted

```

## Custom Server with Extra Capabilities

The `extraCapabilities` field extends the existing list from the ServerConfig.

```bash
cat ./servers/custom-mlserver.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-134
spec:
  serverConfig: mlserver
  extraCapabilities:
  - mlserver-1.3.4
  podSpec:
    containers:
    - image: seldonio/mlserver:1.3.4
      name: mlserver

```

```bash
kubectl create -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```

```
server.mlops.seldon.io/mlserver-134 created

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ${NAMESPACE}
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/mlserver-134 condition met
server.mlops.seldon.io/triton condition met

```

```bash
cat ./models/iris-custom-server.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  server: mlserver-134

```

```bash
kubectl create -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
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
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "a3e17c6c-ee3f-4a51-b890-6fb16385a757",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

```bash
kubectl delete -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted

```

```bash
kubectl delete -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```

```
server.mlops.seldon.io "mlserver-134" deleted

```

```python

```
