---
---

# Kubernetes Server with PVC

{% hint style="info" %}
**Note**:  The Seldon CLI allows you to view information about underlying Seldon resources and make changes to them through the scheduler in non-Kubernetes environments. However, it cannot modify underlying manifests within a Kubernetes cluster. Therefore, using the Seldon CLI for control plane operations in a Kubernetes environment is not recommended. For more details, see [Seldon CLI](../getting-started/cli.md).
{% endhint %}

```
import os
```

```python
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

### Kind cluster setup

To run this example in Kind we need to start Kind with access to a local folder where are models are location. In this example it is a folder in `/tmp` and associate that with a path in the container.

```bash
cat kind-config.yaml
```

```yaml
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane
  extraMounts:
    - hostPath: /tmp/models
      containerPath: /models
```

To start a Kind cluster see, [Learning environment](../installation/learning-environment/README.md).

Create the local folder formodels and copy an example iris sklearn model to it.

```bash
mkdir -p /tmp/models
gsutil cp -r gs://seldon-models/mlserver/iris /tmp/models
```

### Create Server with PVC

Create a storage class and associated persistent colume referencing the `/models` folder where models are stored.

```bash
cat pvc.yaml
```

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-path-immediate
provisioner: rancher.io/local-path
reclaimPolicy: Delete
mountOptions:
  - debug
volumeBindingMode: Immediate
---
kind: PersistentVolume
apiVersion: v1
metadata:
  name: ml-models-pv
  namespace: seldon-mesh
  labels:
    type: local
spec:
  storageClassName: local-path-immediate
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/models"
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: ml-models-pvc
  namespace: seldon-mesh
spec:
  storageClassName: local-path-immediate
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  selector:
    matchLabels:
      type: local
```

Now create a new Server based on the provided MLServer configuration but extend it with our PVC by adding this to the rclone container which will allow rclone to move models from this PVC onto the server.

We also add a new capability `pvc` to allow us to schedule models to this server that has the PVC.

```bash
cat server.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-pvc
spec:
  serverConfig: mlserver
  extraCapabilities:
  - "pvc"
  podSpec:
    volumes:
    - name: models-pvc
      persistentVolumeClaim:
        claimName: ml-models-pvc
    containers:
    - name: rclone
      volumeMounts:
      - name: models-pvc
        mountPath: /var/models
```

### SKLearn Model

Use a simple sklearn iris classification model with the added `pvc` requirement so that MLServer with the PVC is targeted during scheduling.

```bash
cat ./iris.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "/var/models/iris"
  requirements:
  - sklearn
  - pvc
```

```bash
kubectl create -f iris.yaml -n ${NAMESPACE}
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

```
{
  "conditions": [
    {
      "lastTransitionTime": "2022-12-24T11:04:37Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2022-12-24T11:04:37Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}
```

{% tabs %}

{% tab title="curl" %}
```bash
curl -k http://${MESH_IP}:80/v2/models/iris/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Seldon-Model: iris" \
  -H "Content-Type: application/json" \
  -d '{
    "inputs": [
      {
        "name": "predict",
        "datatype": "FP32",
        "shape": [1,4],
        "data": [[1,2,3,4]]
      }
    ]
  }' | jq -M .
``` 
{% endtab %}

{% tab title="seldon-cli" %}
```bash
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```  
{% endtab %}


{% endtabs %}

```
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "dc032bcc-3f4e-4395-a2e4-7c1e3ef56e9e",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": null,
			"data": [
				2
			]
		}
	]
}
```

Do a gRPC inference call

{% tabs %}

{% tab title="curl" %}
```bash
curl -k http://${MESH_IP}:80/v2/models/iris/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Seldon-Model: iris" \
  -H "Content-Type: application/json" \
  -d '{
    "model_name": "iris",
    "inputs": [
      {
        "name": "input",
        "datatype": "FP32",
        "shape": [1,4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .
``` 
{% endtab %}

{% tab title="seldon-cli" %}
```bash
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP}:80 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```  
{% endtab %}


{% endtabs %}


```outputs
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
      "contents": {
        "int64Contents": [
          "2"
        ]
      }
    }
  ]
}
```

```python
kubectl delete -f ./iris.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted
```
