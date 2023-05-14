## Kubernetes Clusterwide SSL Demo

### Setup

Create a Kind cluster by using an ansible playbook from the project `ansible` folder.

```bash
cd ../ansible && ansible-playbook playbooks/kind-cluster.yaml
```

```bash
cd ../ansible && ansible-playbook playbooks/setup-ecosystem.yaml
```

```bash
helm upgrade --install seldon-core-v2-crds  ../k8s/helm-charts/seldon-core-v2-crds -n seldon-mesh
```

```
Release "seldon-core-v2-crds" does not exist. Installing it now.
NAME: seldon-core-v2-crds
LAST DEPLOYED: Sat May 13 10:40:02 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
kubectl create namespace ns1
```

```bash
kubectl create namespace ns2
```

```bash
cat ../k8s/samples/values-tls-dataplane-controlplane-example.yaml
```

```bash
helm install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh --set controller.clusterwide=true --values ../k8s/samples/values-tls-dataplane-controlplane-example.yaml
```

```yaml
NAME: seldon-v2
LAST DEPLOYED: Sat May 13 10:48:56 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns1
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Sat May 13 11:58:39 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns2
```

### Setup TLS Config

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

```bash
seldon config add tls ${PWD}/config-dataplane-tls.json
```

```bash
seldon config activate tls
```

### Launch model in namespace ns1

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
	"id": "b0629eed-5438-4dfc-8314-db461cdb08a0",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"data": [
				2
			]
		}
	]
}

```

### Launch model in namespace ns2

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ns2
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns2
```

```bash
seldon model infer iris --inference-host ${MESH_IP_NS2}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

## TearDown

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns1
```

```
Error from server (NotFound): error when deleting "./models/sklearn-iris-gs.yaml": models.mlops.seldon.io "iris" not found

```

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns2
```

```bash
helm delete seldon-v2-runtime -n ns1
```

```
release "seldon-v2-runtime" uninstalled

```

```bash
helm delete seldon-v2-runtime -n ns2
```

```
Error: uninstall: Release not loaded: seldon-v2-runtime: release: not found

```

```bash
helm delete seldon-v2  -n seldon-mesh
```

```
release "seldon-v2" uninstalled

```

```python

```
