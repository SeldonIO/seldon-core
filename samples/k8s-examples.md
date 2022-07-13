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



    '172.24.255.9'
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

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

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
          "lastTransitionTime": "2022-06-03T14:35:59Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-06-03T14:35:59Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
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
    	"id": "3be6542c-5ad2-4ebc-a0d4-842377653b5d",
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
## Explainer


```bash
cat ./models/income.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/income"
      requirements:
      - sklearn
```
````

```bash
kubectl create -f ./models/income.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income condition met
```
````

```bash
kubectl get model income -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-06-25T10:22:17Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-06-25T10:22:17Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }
```
````

```bash
seldon model infer income --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[53,4,0,2,8,4,2,0,0,0,60,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "8b8ac132-ae7d-44a7-86eb-385092dd8703",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
    	"outputs": [
    		{
    			"name": "predict",
    			"shape": [
    				1
    			],
    			"datatype": "INT64",
    			"parameters": null,
    			"data": [
    				0
    			]
    		}
    	]
    }
```
````

```bash
cat ./models/income-explainer.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/alibi-explain/income"
      explainer:
        type: anchor_tabular
        modelRef: income
```
````

```bash
kubectl create -f ./models/income-explainer.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income-explainer created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income condition met
    model.mlops.seldon.io/income-explainer condition met
```
````

```bash
kubectl get model income-explainer -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-06-25T10:22:32Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-06-25T10:22:32Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }
```
````

```bash
seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[53,4,0,2,8,4,2,0,0,0,60,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "8fb6f648-88bf-4d59-8bd4-42d215e72b28",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
    	"outputs": [
    		{
    			"name": "explanation",
    			"shape": [
    				1
    			],
    			"datatype": "BYTES",
    			"parameters": {
    				"content_type": "str",
    				"headers": null
    			},
    			"data": [
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.7.0\"}, \"data\": {\"anchor\": [\"Marital Status = Separated\", \"Capital Loss <= 0.00\"], \"precision\": 0.9813084112149533, \"coverage\": 0.17423333333333332, \"raw\": {\"feature\": [3, 9], \"mean\": [0.96875, 0.9813084112149533], \"precision\": [0.96875, 0.9813084112149533], \"coverage\": [0.18063333333333334, 0.17423333333333332], \"examples\": [{\"covered_true\": [[24, 4, 4, 2, 2, 5, 4, 0, 0, 0, 40, 9], [41, 4, 4, 2, 1, 5, 4, 0, 0, 0, 40, 9], [41, 4, 4, 2, 7, 1, 4, 1, 0, 0, 21, 9], [31, 6, 1, 2, 8, 0, 4, 1, 0, 0, 60, 0], [32, 4, 4, 2, 2, 1, 4, 1, 0, 0, 40, 9], [44, 4, 4, 2, 8, 5, 4, 0, 0, 0, 70, 9], [20, 0, 3, 2, 0, 1, 3, 1, 0, 1602, 40, 9], [22, 4, 4, 2, 2, 2, 4, 1, 0, 0, 55, 9], [21, 4, 4, 2, 1, 1, 4, 0, 0, 0, 40, 9], [34, 4, 4, 2, 7, 0, 1, 1, 0, 0, 40, 7]], \"covered_false\": [[26, 4, 6, 2, 5, 0, 4, 1, 0, 1977, 40, 9], [44, 4, 6, 2, 5, 0, 4, 1, 99999, 0, 65, 9], [39, 4, 2, 2, 5, 4, 4, 0, 0, 0, 80, 9], [34, 4, 5, 2, 5, 1, 4, 1, 0, 2258, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[55, 4, 1, 2, 5, 0, 4, 1, 0, 0, 40, 9], [65, 4, 3, 2, 6, 2, 4, 0, 0, 0, 20, 9], [18, 0, 3, 2, 0, 3, 1, 0, 0, 0, 24, 3], [38, 4, 5, 2, 1, 1, 4, 0, 0, 0, 50, 3], [39, 4, 0, 2, 4, 0, 2, 1, 0, 0, 60, 9], [29, 4, 4, 2, 7, 4, 2, 0, 0, 0, 25, 9], [30, 4, 1, 2, 1, 1, 4, 0, 0, 0, 40, 9], [49, 4, 0, 2, 5, 0, 4, 1, 0, 0, 45, 9], [35, 4, 1, 2, 6, 0, 4, 1, 0, 0, 52, 9], [57, 4, 4, 2, 1, 2, 2, 0, 0, 0, 40, 5]], \"covered_false\": [[44, 4, 1, 2, 8, 0, 4, 1, 7298, 0, 48, 9], [67, 5, 4, 2, 5, 5, 4, 0, 20051, 0, 30, 1], [57, 1, 5, 2, 8, 0, 2, 1, 15024, 0, 40, 9], [60, 7, 2, 2, 5, 0, 4, 1, 0, 0, 55, 9], [32, 4, 4, 2, 2, 0, 4, 1, 99999, 0, 40, 9], [25, 4, 0, 2, 6, 1, 4, 1, 27828, 0, 40, 9], [65, 4, 1, 2, 1, 0, 4, 1, 10605, 0, 20, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Separated\", \"Capital Loss <= 0.00\"], \"prediction\": [0], \"instance\": [53.0, 4.0, 0.0, 2.0, 8.0, 4.0, 2.0, 0.0, 0.0, 0.0, 60.0, 9.0], \"instances\": [[53.0, 4.0, 0.0, 2.0, 8.0, 4.0, 2.0, 0.0, 0.0, 0.0, 60.0, 9.0]]}}}"
    			]
    		}
    	]
    }
```
````

```bash
kubectl delete -f ./models/income.yaml
kubectl delete -f ./models/income-explainer.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "income" deleted
    model.mlops.seldon.io "income-explainer" deleted
```
````
## Custom Server


```bash
cat ./servers/custom-mlserver.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Server
    metadata:
      name: mlserver-custom
      namespace: seldon-mesh
    spec:
      serverConfig: mlserver
      podSpec:
        containers:
        - image: cliveseldon/mlserver:1.1.0.explain
          name: mlserver
```
````

```bash
kubectl create -f ./servers/custom-mlserver.yaml
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io/mlserver-custom created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io/mlserver condition met
    server.mlops.seldon.io/mlserver-custom condition met
```
````

```bash
cat ./models/iris-custom-server.yaml
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
      server: mlserver-custom
```
````

```bash
kubectl create -f ./models/iris-custom-server.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris condition met
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
    	"id": "1bc7c802-b380-480c-96be-95472b76c2dc",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
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
kubectl delete -f ./models/iris-custom-server.yaml
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "iris" deleted
```
````

```bash
kubectl delete -f ./servers/custom-mlserver.yaml
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io "mlserver-custom" deleted
```
````

```python

```
