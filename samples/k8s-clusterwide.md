## Seldon V2 Multi-Namespace Kubernetes Example

```bash
helm upgrade --install seldon-core-v2-crds  ../k8s/helm-charts/seldon-core-v2-crds -n seldon-mesh
```

```
Release "seldon-core-v2-crds" does not exist. Installing it now.
NAME: seldon-core-v2-crds
LAST DEPLOYED: Fri Jun 23 14:41:22 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm upgrade --install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh --set controller.clusterwide=true
```

```
Release "seldon-v2" does not exist. Installing it now.
NAME: seldon-v2
LAST DEPLOYED: Fri Jun 23 14:41:26 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
kubectl create namespace ns1
kubectl create namespace ns2
```

```
namespace/ns1 created
namespace/ns2 created

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns1 --wait
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Fri Jun 23 14:10:23 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-servers ../k8s/helm-charts/seldon-core-v2-servers  -n ns1 --wait
```

```yaml
NAME: seldon-v2-servers
LAST DEPLOYED: Fri Jun 23 14:10:38 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns2 --wait
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Fri Jun 23 14:10:42 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-servers ../k8s/helm-charts/seldon-core-v2-servers  -n ns2 --wait
```

```yaml
NAME: seldon-v2-servers
LAST DEPLOYED: Fri Jun 23 14:10:44 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns1 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS1=MESH_IP[0]
import os
os.environ['MESH_IP_NS1'] = MESH_IP_NS1
MESH_IP_NS1
```

```
'172.21.255.2'

```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns2 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS2=MESH_IP[0]
import os
os.environ['MESH_IP_NS2'] = MESH_IP_NS2
MESH_IP_NS2
```

```
'172.21.255.4'

```

### Run Models in Different Namespaces

```bash
cat ./models/sklearn-iris-gs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.0/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki

```

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ns1
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns1
```

```
model.mlops.seldon.io/iris condition met

```

```bash
seldon model infer iris --inference-host ${MESH_IP_NS1}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "276de7e7-9f2f-4329-9179-08d4d54bb0b5",
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
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP_NS1}:80 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```

```json
{
  "modelName": "iris_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "predict",
      "datatype": "INT64",
      "shape": [
        "1",
        "1"
      ],
      "parameters": {
        "content_type": {
          "stringParam": "np"
        }
      },
      "contents": {
        "int64Contents": [
          "2"
        ]
      }
    }
  ]
}

```

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ns2
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns2
```

```
model.mlops.seldon.io/iris condition met

```

```bash
seldon model infer iris --inference-host ${MESH_IP_NS2}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "7c0bb004-be2c-4483-97a9-6fd2e0a834ad",
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
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP_NS2}:80 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```

```json
{
  "modelName": "iris_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "predict",
      "datatype": "INT64",
      "shape": [
        "1",
        "1"
      ],
      "parameters": {
        "content_type": {
          "stringParam": "np"
        }
      },
      "contents": {
        "int64Contents": [
          "2"
        ]
      }
    }
  ]
}

```

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns1
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns2
```

```
model.mlops.seldon.io "iris" deleted
model.mlops.seldon.io "iris" deleted

```

## TearDown

```bash
helm delete seldon-v2-servers -n ns1 --wait
helm delete seldon-v2-servers -n ns2 --wait
```

```
release "seldon-v2-servers" uninstalled
release "seldon-v2-servers" uninstalled

```

```bash
helm delete seldon-v2-runtime -n ns1 --wait
helm delete seldon-v2-runtime -n ns2 --wait
```

```
release "seldon-v2-runtime" uninstalled
Error: uninstall: Release not loaded: seldon-v2-runtime: release: not found

```

```bash
helm delete seldon-v2 -n seldon-mesh --wait
```

```
release "seldon-v2" uninstalled

```

```bash
helm delete seldon-core-v2-crds -n seldon-mesh
```

```
release "seldon-core-v2-crds" uninstalled

```

```bash
kubectl delete namespace ns1
kubectl delete namespace ns2
```

```
namespace "ns1" deleted
namespace "ns2" deleted

```

```python

```
