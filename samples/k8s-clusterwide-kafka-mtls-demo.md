## Kubernetes Clusterwide Kafka mTLS Demo

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
LAST DEPLOYED: Sun May 14 14:01:58 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
kubectl create namespace ns1
```

```
namespace/ns1 created

```

```bash
kubectl create namespace ns2
```

```
namespace/ns2 created

```

```bash
cat ../k8s/samples/values-strimzi-kafka-mtls.yaml
```

```
---
kafka:
  bootstrap: seldon-kafka-bootstrap.seldon-mesh.svc.cluster.local:9093

security:
  kafka:
    protocol: SSL
    ssl:
      client:
        secret: seldon
        brokerValidationSecret: seldon-cluster-ca-cert
        keyPath: /tmp/certs/kafka/client/user.key
        crtPath: /tmp/certs/kafka/client/user.crt
        caPath: /tmp/certs/kafka/client/ca.crt
        brokerCaPath: /tmp/certs/kafka/broker/ca.crt

```

```bash
helm install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh --set controller.clusterwide=true --values ../k8s/samples/values-strimzi-kafka-mtls.yaml
```

```yaml
NAME: seldon-v2
LAST DEPLOYED: Sun May 14 14:02:33 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

## Copy Strimzi secrets to namespaces

Strimzi doesn't allow a single Kafka cluster to be used in TLS easily from multiple namespaces without copying the user Secrets created by the KafkaUser to those namespaces.

```bash
kubectl create -f ../k8s/samples/strimzi-example-tls-user.yaml -n ns1
```

```
kafkauser.kafka.strimzi.io/seldon created

```

```bash
kubectl get secret seldon -n seldon-mesh -o json | jq 'del(.metadata.ownerReferences) | del(.metadata.namespace)' | kubectl create -f - -n ns1
```

```bash
kubectl get secret seldon-cluster-ca-cert -n seldon-mesh -o json | jq 'del(.metadata.ownerReferences) | del(.metadata.namespace)' | kubectl create -f - -n ns1
```

```
secret/seldon-cluster-ca-cert created

```

```bash
kubectl get secret seldon -n seldon-mesh -o json | jq 'del(.metadata.ownerReferences) | del(.metadata.namespace)' | kubectl create -f - -n ns2
```

```
secret/seldon created

```

```bash
kubectl get secret seldon-cluster-ca-cert -n seldon-mesh -o json | jq 'del(.metadata.ownerReferences) | del(.metadata.namespace)' | kubectl create -f - -n ns2
```

```
secret/seldon-cluster-ca-cert created

```

## Create SeldonRuntimes

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns1
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Sun May 14 14:03:14 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns2
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Sun May 14 14:04:39 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

### Get Inference Endpoints

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

### Launch Pipeline in Namespace ns1

```bash
kubectl create -f ./models/tfsimple1.yaml -n ns1
kubectl create -f ./models/tfsimple2.yaml -n ns1
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns1
```

```
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met

```

```bash
kubectl create -f ./pipelines/tfsimples.yaml -n ns1
```

```
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ns1
```

```
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

### Launch Pipeline in Namespace ns2

```bash
kubectl create -f ./models/tfsimple1.yaml -n ns2
kubectl create -f ./models/tfsimple2.yaml -n ns2
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns2
```

```
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met

```

```bash
kubectl create -f ./pipelines/tfsimples.yaml -n ns2
```

```
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ns2
```

```
pipeline.mlops.seldon.io/tfsimples condition met

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

## TearDown

```bash
kubectl delete -f ./pipelines/tfsimples.yaml -n ns1
kubectl delete -f ./pipelines/tfsimples.yaml -n ns2
kubectl delete -f ./models/tfsimple1.yaml -n ns1
kubectl delete -f ./models/tfsimple2.yaml -n ns1
kubectl delete -f ./models/tfsimple1.yaml -n ns2
kubectl delete -f ./models/tfsimple2.yaml -n ns2
```

```
pipeline.mlops.seldon.io "tfsimples" deleted
pipeline.mlops.seldon.io "tfsimples" deleted
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted

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
release "seldon-v2-runtime" uninstalled

```

```bash
helm delete seldon-v2  -n seldon-mesh
```

```
release "seldon-v2" uninstalled

```

```python

```
