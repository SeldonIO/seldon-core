# Multi-Namespace Kubernetes

```bash
helm upgrade --install seldon-core-v2-crds  ../k8s/helm-charts/seldon-core-v2-crds -n seldon-mesh
```

```
Release "seldon-core-v2-crds" does not exist. Installing it now.
NAME: seldon-core-v2-crds
LAST DEPLOYED: Tue Aug 15 11:01:03 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

The below setup also illustrates using kafka specific prefixes for topics and consumerIds for isolation where the kafka cluster is shared with other applications and you want to enforce constraints. You would not strictly need this in this example as we install Kafka just for Seldon here.

```bash
helm upgrade --install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh \
    --set controller.clusterwide=true \
    --set kafka.topicPrefix=myorg \
    --set kafka.consumerGroupIdPrefix=myorg
```

```
Release "seldon-v2" does not exist. Installing it now.
NAME: seldon-v2
LAST DEPLOYED: Tue Aug 15 11:01:07 2023
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
LAST DEPLOYED: Tue Aug 15 11:01:11 2023
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
LAST DEPLOYED: Tue Aug 15 10:47:31 2023
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
LAST DEPLOYED: Tue Aug 15 10:53:12 2023
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
LAST DEPLOYED: Tue Aug 15 10:53:28 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ns1
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/triton condition met

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ns2
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/triton condition met

```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns1 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS1=MESH_IP[0]
import os
os.environ['MESH_IP_NS1'] = MESH_IP_NS1
MESH_IP_NS1
```

```
'172.18.255.2'

```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns2 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS2=MESH_IP[0]
import os
os.environ['MESH_IP_NS2'] = MESH_IP_NS2
MESH_IP_NS2
```

```
'172.18.255.4'

```

### Run Models in Different Namespaces

{% hint style="info" %}
**Note**:  The Seldon CLI allows you to view information about underlying Seldon resources and make changes to them through the scheduler in non-Kubernetes environments. However, it cannot modify underlying manifests within a Kubernetes cluster. Therefore, using the Seldon CLI for control plane operations in a Kubernetes environment is not recommended. For more details, see [Seldon CLI](../cli/).
{% endhint %}

```bash
cat ./models/sklearn-iris-gs.yaml
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
	"id": "3ca1757c-df02-4e57-87c1-38311bcc5943",
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
	"id": "f706a23e-775f-4765-bd18-2e98d83bf7d5",
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

## Pipelines

```bash
kubectl create -f ./models/tfsimple1.yaml -n ns1
kubectl create -f ./models/tfsimple2.yaml -n ns1
kubectl create -f ./models/tfsimple1.yaml -n ns2
kubectl create -f ./models/tfsimple2.yaml -n ns2
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns1
kubectl wait --for condition=ready --timeout=300s model --all -n ns2
```

```
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met

```

```bash
kubectl create -f ./pipelines/tfsimples.yaml -n ns1
kubectl create -f ./pipelines/tfsimples.yaml -n ns2
```

```
pipeline.mlops.seldon.io/tfsimples created
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ns1
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ns2
```

```
pipeline.mlops.seldon.io/tfsimples condition met
pipeline.mlops.seldon.io/tfsimples condition met

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc --inference-host ${MESH_IP_NS1}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    },
    {
      "name": "OUTPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc --inference-host ${MESH_IP_NS2}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    },
    {
      "name": "OUTPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    }
  ]
}

```

If you have installed Kafka via the ansible playbook setup-ecosystem then you can use the following command to see the consumer group ids which are reflecting the settings we created.

```bash
kubectl exec seldon-kafka-0 -n seldon-mesh -- bin/kafka-consumer-groups.sh --list --bootstrap-server localhost:9092
```

```
myorg-ns2-seldon-pipelinegateway-dfd61b49-4bb9-4684-adce-0b7cc215d3af
myorg-ns2-seldon-modelgateway-17
myorg-ns1-seldon-pipelinegateway-d4fc83e6-29cb-442e-90cd-92a389961cfe
myorg-ns2-seldon-modelgateway-60
myorg-ns2-seldon-dataflow-73d465744b7b1b5be20e88d6245e50bd
myorg-ns1-seldon-modelgateway-60
myorg-ns1-seldon-modelgateway-17
myorg-ns1-seldon-dataflow-f563e04e093caa20c03e6eced084331b

```

We can similarly show the topics that have been created.

```bash
kubectl exec seldon-kafka-0 -n seldon-mesh -- bin/kafka-topics.sh --bootstrap-server=localhost:9092 --list
```

```
__consumer_offsets
myorg.ns1.errors.errors
myorg.ns1.model.iris.inputs
myorg.ns1.model.iris.outputs
myorg.ns1.model.tfsimple1.inputs
myorg.ns1.model.tfsimple1.outputs
myorg.ns1.model.tfsimple2.inputs
myorg.ns1.model.tfsimple2.outputs
myorg.ns1.pipeline.tfsimples.inputs
myorg.ns1.pipeline.tfsimples.outputs
myorg.ns2.errors.errors
myorg.ns2.model.iris.inputs
myorg.ns2.model.iris.outputs
myorg.ns2.model.tfsimple1.inputs
myorg.ns2.model.tfsimple1.outputs
myorg.ns2.model.tfsimple2.inputs
myorg.ns2.model.tfsimple2.outputs
myorg.ns2.pipeline.tfsimples.inputs
myorg.ns2.pipeline.tfsimples.outputs

```

```bash
kubectl delete -f ./pipelines/tfsimples.yaml -n ns1
kubectl delete -f ./pipelines/tfsimples.yaml -n ns2
```

```
pipeline.mlops.seldon.io "tfsimples" deleted
pipeline.mlops.seldon.io "tfsimples" deleted

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ns1
kubectl delete -f ./models/tfsimple2.yaml -n ns1
kubectl delete -f ./models/tfsimple1.yaml -n ns2
kubectl delete -f ./models/tfsimple2.yaml -n ns2
```

```
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted

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
release "seldon-v2-runtime" uninstalled

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
