## Seldon V2 Kubernetes Examples



```python
MESH_IP=kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```
````{collapse} Expand to see output
```bash



    '172.22.255.9'
```
````

### Model


```bash
cat ./models/sklearn-iris-gs.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
```
````

```python
buildkubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

    model.mlops.seldon.io/iris condition met
```
````

```bash
kubectl get model iris -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-05-26T10:09:32Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-05-26T10:09:32Z",
          "status": "True",
          "type": "Ready"
        }
      ]
    }
```
````

```bash
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "5890ac9d-c1ed-4343-b9e6-a460bead5fe8",
    	"parameters": null,
    	"outputs": [
    		{
    			"name": "predict",
    			"shape": [
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
````

```bash
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP}:80 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "iris_1",
      "modelVersion": "1",
      "outputs": [
        {
          "name": "predict",
          "datatype": "INT64",
          "shape": [
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
````

```bash
kubectl get server mlserver -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-05-26T09:58:57Z",
          "status": "True",
          "type": "Ready"
        },
        {
          "lastTransitionTime": "2022-05-26T09:58:57Z",
          "reason": "StatefulSet replicas matches desired replicas",
          "status": "True",
          "type": "StatefulSetReady"
        }
      ],
      "loadedModels": 1
    }
```
````

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "iris" deleted
```
````
### Experiment


```bash
cat ./experiments/sklearn1.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
cat ./experiments/sklearn2.yaml 
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
kubectl create -f ./experiments/sklearn1.yaml
kubectl create -f ./experiments/sklearn2.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
    model.mlops.seldon.io/iris2 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris condition met
    model.mlops.seldon.io/iris2 condition met
```
````

```bash
cat ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
      namespace: seldon-mesh
    spec:
      defaultModel: iris
      candidates:
      - modelName: iris
        weight: 50
      - modelName: iris2
        weight: 50
```
````

```bash
kubectl create -f ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```json

    experiment.mlops.seldon.io/experiment-sample created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s experiment --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    experiment.mlops.seldon.io/experiment-sample condition met
```
````

```bash
seldon model infer --inference-host ${MESH_IP}:80 -i 50 iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    map[iris2_1:27 iris_1:23]
```
````

```bash
kubectl delete -f ./experiments/ab-default-model.yaml 
kubectl delete -f ./experiments/sklearn1.yaml
kubectl delete -f ./experiments/sklearn2.yaml
```
````{collapse} Expand to see output
```json

    experiment.mlops.seldon.io "experiment-sample" deleted
    model.mlops.seldon.io "iris" deleted
    model.mlops.seldon.io "iris2" deleted
```
````
### Pipeline - model chain


```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
```
````

```bash
kubectl create -f ./models/tfsimple1.yaml
kubectl create -f ./models/tfsimple2.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 condition met
    model.mlops.seldon.io/tfsimple2 condition met
```
````

```bash
cat ./pipelines/tfsimples.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimples
      namespace: seldon-mesh
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
          inputs:
          - tfsimple1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple1.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple2
```
````

```bash
kubectl create -f ./pipelines/tfsimples.yaml
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/tfsimples created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/tfsimples condition met
```
````

```bash
seldon pipeline infer tfsimples --inference-mode grpc --inference-host ${MESH_IP}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
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
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
```
````

```bash
kubectl delete -f ./pipelines/tfsimples.yaml
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io "tfsimples" deleted
```
````

```bash
kubectl delete -f ./models/tfsimple1.yaml
kubectl delete -f ./models/tfsimple2.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "tfsimple1" deleted
    model.mlops.seldon.io "tfsimple2" deleted
```
````
### Pipeline - model join


```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
cat ./models/tfsimple3.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple3
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
```
````

```bash
kubectl create -f ./models/tfsimple1.yaml
kubectl create -f ./models/tfsimple2.yaml
kubectl create -f ./models/tfsimple3.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created
    model.mlops.seldon.io/tfsimple3 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 condition met
    model.mlops.seldon.io/tfsimple2 condition met
    model.mlops.seldon.io/tfsimple3 condition met
```
````

```bash
cat ./pipelines/tfsimples-join.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: join
      namespace: seldon-mesh
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
        - name: tfsimple3      
          inputs:
          - tfsimple1.outputs.OUTPUT0
          - tfsimple2.outputs.OUTPUT1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple2.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple3
```
````

```bash
kubectl create -f ./pipelines/tfsimples-join.yaml
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/join created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/join condition met
```
````

```bash
seldon pipeline infer join --inference-mode grpc --inference-host ${MESH_IP}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
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
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
```
````

```bash
kubectl delete -f ./pipelines/tfsimples-join.yaml
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io "join" deleted
```
````

```bash
kubectl delete -f ./models/tfsimple1.yaml
kubectl delete -f ./models/tfsimple2.yaml
kubectl delete -f ./models/tfsimple3.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "tfsimple1" deleted
    model.mlops.seldon.io "tfsimple2" deleted
    model.mlops.seldon.io "tfsimple3" deleted
```
````

```python

```
