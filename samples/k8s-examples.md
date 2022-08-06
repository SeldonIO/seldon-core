## Seldon V2 Kubernetes Examples



```python
import os
os.environ["NAMESPACE"] = "test"
```


```python
MESH_IP=kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```
````{collapse} Expand to see output
```bash



    '172.18.255.9'
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
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
      memory: 100Ki
```
````

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris condition met
```
````

```bash
kubectl get model iris -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-08-04T14:33:14Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-08-04T14:33:14Z",
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
    	"id": "d04092bb-b434-4acb-9985-941169be52e3",
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
kubectl get server mlserver -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-08-04T14:31:33Z",
          "status": "True",
          "type": "Ready"
        },
        {
          "lastTransitionTime": "2022-08-04T14:31:33Z",
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
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "iris" deleted
```
````
### Experiment


```bash
cat ./models/sklearn1.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
cat ./models/sklearn2.yaml 
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
kubectl create -f ./models/sklearn1.yaml -n ${NAMESPACE}
kubectl create -f ./models/sklearn2.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
    model.mlops.seldon.io/iris2 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
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
    spec:
      default: iris
      candidates:
      - modelName: iris
        weight: 50
      - modelName: iris2
        weight: 50
```
````

```bash
kubectl create -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    experiment.mlops.seldon.io/experiment-sample created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s experiment --all -n ${NAMESPACE}
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

    map[:iris2_1::25 :iris_1::25]
```
````

```bash
kubectl delete -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
kubectl delete -f ./models/sklearn1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/sklearn2.yaml -n ${NAMESPACE}
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
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
```
````

```bash
kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
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
kubectl create -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/tfsimples created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
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
kubectl delete -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io "tfsimples" deleted
```
````

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
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
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple3
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
```
````

```bash
kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple3.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created
    model.mlops.seldon.io/tfsimple3 created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
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
kubectl create -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io/join created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
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
kubectl delete -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    pipeline.mlops.seldon.io "join" deleted
```
````

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple3.yaml -n ${NAMESPACE}
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
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
      requirements:
      - sklearn
```
````

```bash
kubectl create -f ./models/income.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income condition met
```
````

```bash
kubectl get model income -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-08-04T11:59:29Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-08-04T11:59:29Z",
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
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "389eebf2-069b-4a4b-8ae1-005402493469",
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
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
      explainer:
        type: anchor_tabular
        modelRef: income
```
````

```bash
kubectl create -f ./models/income-explainer.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income-explainer created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/income condition met
    model.mlops.seldon.io/income-explainer condition met
```
````

```bash
kubectl get model income-explainer -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-08-04T11:59:53Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-08-04T11:59:53Z",
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
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "02b6d055-e587-4485-a36c-d09b5830748e",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.7.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9938650306748467, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7913148371531966, 0.9178082191780822, 0.9938650306748467], \"precision\": [0.7913148371531966, 0.9178082191780822, 0.9938650306748467], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[66, 0, 1, 1, 0, 0, 4, 1, 0, 0, 6, 9], [36, 2, 1, 1, 5, 1, 4, 0, 0, 0, 60, 9], [40, 4, 1, 1, 8, 0, 4, 1, 0, 0, 75, 9], [42, 4, 1, 1, 6, 4, 4, 1, 0, 0, 45, 0], [49, 2, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [37, 4, 1, 1, 5, 0, 4, 1, 0, 0, 55, 9], [61, 4, 5, 1, 6, 0, 4, 1, 3103, 0, 50, 9], [54, 4, 1, 1, 7, 0, 4, 1, 0, 0, 40, 9], [35, 2, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [24, 4, 1, 1, 6, 1, 4, 1, 0, 0, 47, 9]], \"covered_false\": [[38, 6, 5, 1, 5, 0, 2, 1, 99999, 0, 60, 9], [40, 4, 1, 1, 7, 5, 1, 0, 7688, 0, 52, 6], [43, 4, 5, 1, 5, 1, 4, 1, 0, 0, 50, 9], [51, 4, 1, 1, 8, 1, 0, 1, 0, 0, 50, 9], [42, 4, 1, 1, 8, 0, 4, 1, 0, 2415, 60, 9], [22, 4, 1, 1, 5, 1, 0, 0, 14344, 0, 40, 9], [46, 4, 5, 1, 8, 1, 4, 0, 27828, 0, 50, 9], [50, 4, 1, 1, 8, 0, 4, 1, 7298, 0, 50, 9], [43, 1, 1, 1, 8, 1, 4, 1, 0, 0, 40, 9], [56, 6, 5, 1, 2, 0, 4, 1, 0, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[36, 4, 1, 1, 6, 3, 4, 1, 0, 1902, 45, 9], [55, 4, 2, 1, 8, 3, 4, 1, 0, 0, 60, 9], [46, 4, 2, 1, 5, 3, 4, 1, 0, 0, 70, 9], [45, 2, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [55, 5, 5, 1, 8, 3, 4, 1, 0, 0, 60, 9], [24, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [58, 6, 1, 1, 8, 3, 4, 1, 0, 0, 50, 9], [34, 4, 5, 1, 4, 3, 2, 0, 0, 0, 40, 0], [30, 4, 1, 1, 4, 3, 4, 1, 0, 0, 60, 4], [39, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 1]], \"covered_false\": [[45, 2, 1, 1, 2, 3, 1, 1, 7298, 0, 40, 9], [42, 1, 5, 1, 8, 3, 4, 0, 14084, 0, 60, 9], [46, 4, 1, 1, 2, 3, 4, 1, 15024, 0, 40, 9], [36, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [47, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [42, 2, 1, 1, 1, 3, 4, 0, 99999, 0, 40, 9], [55, 4, 5, 1, 5, 3, 4, 1, 15024, 0, 50, 0], [48, 1, 5, 1, 8, 3, 4, 0, 10520, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[27, 4, 5, 1, 8, 3, 4, 0, 0, 0, 25, 9], [45, 1, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [32, 4, 1, 1, 5, 3, 4, 1, 0, 0, 50, 9], [81, 6, 1, 1, 5, 3, 4, 0, 0, 1668, 3, 4], [44, 2, 1, 1, 5, 3, 4, 1, 0, 0, 50, 9], [44, 7, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [41, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [48, 7, 2, 1, 5, 3, 4, 1, 0, 0, 50, 9], [32, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [68, 4, 5, 1, 5, 3, 4, 1, 0, 0, 18, 0]], \"covered_false\": [[67, 5, 1, 1, 8, 3, 4, 1, 0, 2392, 75, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
    			]
    		}
    	]
    }
```
````

```bash
kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
kubectl delete -f ./models/income-explainer.yaml -n ${NAMESPACE}
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
    spec:
      serverConfig: mlserver
      podSpec:
        containers:
        - image: cliveseldon/mlserver:1.2.0.dev1
          name: mlserver
```
````

```bash
kubectl create -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io/mlserver-custom created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io/mlserver condition met
    server.mlops.seldon.io/mlserver-custom condition met
    server.mlops.seldon.io/triton condition met
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
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      server: mlserver-custom
```
````

```bash
kubectl create -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io/iris created
```
````

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
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
    	"id": "ddfb14fa-dd0e-4960-9a89-4137570f5feb",
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
kubectl delete -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    model.mlops.seldon.io "iris" deleted
```
````

```bash
kubectl delete -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```
````{collapse} Expand to see output
```json

    server.mlops.seldon.io "mlserver-custom" deleted
```
````

```python

```
