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




    '172.19.255.1'



### Model


```python
!cat ./models/sklearn-iris-gs.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
      memory: 100Ki



```python
!kubectl create -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris condition met



```python
!kubectl get model iris -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-16T18:11:41Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-16T18:11:41Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "de1b3d19-3fcb-4865-b59b-615a5b5f1e69",
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



```python
!seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP}:80 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```

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



```python
!kubectl get server mlserver -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-16T18:06:01Z",
          "status": "True",
          "type": "Ready"
        },
        {
          "lastTransitionTime": "2022-11-16T18:06:01Z",
          "reason": "StatefulSet replicas matches desired replicas",
          "status": "True",
          "type": "StatefulSetReady"
        }
      ],
      "loadedModels": 1
    }



```python
!kubectl delete -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "iris" deleted


### Experiment


```python
!cat ./models/sklearn1.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!cat ./models/sklearn2.yaml 
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!kubectl create -f ./models/sklearn1.yaml -n ${NAMESPACE}
!kubectl create -f ./models/sklearn2.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris created
    model.mlops.seldon.io/iris2 created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris condition met
    model.mlops.seldon.io/iris2 condition met



```python
!cat ./experiments/ab-default-model.yaml 
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
    spec:
      default: iris
      candidates:
      - name: iris
        weight: 50
      - name: iris2
        weight: 50



```python
!kubectl create -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
```

    experiment.mlops.seldon.io/experiment-sample created



```python
!kubectl wait --for condition=ready --timeout=300s experiment --all -n ${NAMESPACE}
```

    experiment.mlops.seldon.io/experiment-sample condition met



```python
!seldon model infer --inference-host ${MESH_IP}:80 -i 50 iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    Success: map[:iris2_1::27 :iris_1::23]



```python
!kubectl delete -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/sklearn1.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/sklearn2.yaml -n ${NAMESPACE}
```

    experiment.mlops.seldon.io "experiment-sample" deleted
    model.mlops.seldon.io "iris" deleted
    model.mlops.seldon.io "iris2" deleted


### Pipeline - model chain


```python
!cat ./models/tfsimple1.yaml 
!cat ./models/tfsimple2.yaml
```

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



```python
!kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/tfsimple1 condition met
    model.mlops.seldon.io/tfsimple2 condition met



```python
!cat ./pipelines/tfsimples.yaml
```

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



```python
!kubectl create -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/tfsimples created



```python
!kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/tfsimples condition met



```python
!seldon pipeline infer tfsimples --inference-mode grpc --inference-host ${MESH_IP}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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



```python
!kubectl delete -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io "tfsimples" deleted



```python
!kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "tfsimple1" deleted
    model.mlops.seldon.io "tfsimple2" deleted


### Pipeline - model join


```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
!cat ./models/tfsimple3.yaml
```

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



```python
!kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
!kubectl create -f ./models/tfsimple3.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created
    model.mlops.seldon.io/tfsimple3 created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/tfsimple1 condition met
    model.mlops.seldon.io/tfsimple2 condition met
    model.mlops.seldon.io/tfsimple3 condition met



```python
!cat ./pipelines/tfsimples-join.yaml
```

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



```python
!kubectl create -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/join created



```python
!kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/join condition met



```python
!seldon pipeline infer join --inference-mode grpc --inference-host ${MESH_IP}:80 \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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



```python
!kubectl delete -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io "join" deleted



```python
!kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/tfsimple3.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "tfsimple1" deleted
    model.mlops.seldon.io "tfsimple2" deleted
    model.mlops.seldon.io "tfsimple3" deleted


## Explainer


```python
!cat ./models/income.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
      requirements:
      - sklearn



```python
!kubectl create -f ./models/income.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met



```python
!kubectl get model income -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-16T18:16:34Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-16T18:16:34Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "11e439cf-c967-44fd-aa13-baefd9c4d407",
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



```python
!cat ./models/income-explainer.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
      explainer:
        type: anchor_tabular
        modelRef: income



```python
!kubectl create -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income-explainer created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met
    model.mlops.seldon.io/income-explainer condition met



```python
!kubectl get model income-explainer -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-16T18:16:51Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-16T18:16:51Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "7aa0f8ec-a68a-4141-a757-56ebd39f2845",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Occupation = Admin\", \"Capital Gain <= 0.00\", \"Capital Loss <= 0.00\"], \"precision\": 0.9888392857142857, \"coverage\": 0.03760569648420116, \"raw\": {\"feature\": [3, 4, 8, 9], \"mean\": [0.7713498622589532, 0.9002217294900222, 0.9636363636363636, 0.9888392857142857], \"precision\": [0.7713498622589532, 0.9002217294900222, 0.9636363636363636, 0.9888392857142857], \"coverage\": [0.3037383177570093, 0.040943480195816645, 0.03871829105473965, 0.03760569648420116], \"examples\": [{\"covered_true\": [[39, 4, 2, 1, 5, 1, 4, 1, 0, 0, 45, 9], [56, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [62, 6, 1, 1, 5, 0, 4, 1, 0, 0, 60, 9], [48, 4, 1, 1, 6, 0, 4, 1, 0, 0, 45, 9], [29, 4, 1, 1, 1, 1, 4, 0, 0, 0, 40, 9], [46, 1, 1, 1, 5, 1, 4, 1, 0, 0, 40, 9], [55, 4, 2, 1, 5, 0, 4, 1, 0, 0, 40, 9], [37, 4, 1, 1, 8, 0, 4, 1, 0, 0, 50, 9], [30, 4, 1, 1, 8, 0, 4, 1, 0, 0, 50, 9], [39, 4, 5, 1, 1, 1, 4, 1, 0, 0, 40, 0]], \"covered_false\": [[25, 5, 1, 1, 6, 1, 4, 1, 0, 0, 45, 9], [54, 4, 5, 1, 8, 0, 4, 1, 0, 0, 45, 9], [38, 4, 5, 1, 8, 0, 4, 1, 0, 0, 60, 9], [50, 4, 1, 1, 6, 1, 4, 1, 0, 0, 45, 9], [43, 4, 1, 1, 5, 5, 4, 0, 15024, 0, 50, 9], [46, 2, 1, 1, 5, 1, 4, 0, 0, 1408, 40, 9], [42, 4, 1, 1, 4, 0, 4, 1, 7298, 0, 45, 9], [46, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 45, 9], [49, 7, 5, 1, 6, 0, 4, 1, 99999, 0, 80, 9], [47, 4, 1, 1, 8, 0, 4, 1, 0, 0, 60, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[46, 4, 1, 1, 1, 0, 4, 1, 0, 0, 35, 9], [49, 7, 2, 1, 1, 4, 4, 1, 0, 0, 50, 9], [45, 4, 1, 1, 1, 5, 1, 0, 5178, 0, 40, 7], [63, 0, 1, 1, 1, 0, 4, 1, 0, 0, 40, 9], [47, 4, 1, 1, 1, 0, 4, 1, 0, 0, 60, 9], [24, 4, 1, 1, 1, 1, 4, 1, 0, 0, 30, 9], [40, 4, 5, 1, 1, 3, 4, 0, 0, 0, 40, 9], [50, 4, 5, 1, 1, 0, 4, 1, 0, 0, 50, 9], [55, 1, 5, 1, 1, 0, 4, 1, 0, 0, 60, 9], [46, 1, 5, 1, 1, 0, 4, 1, 0, 0, 40, 1]], \"covered_false\": [[63, 4, 1, 1, 1, 5, 4, 0, 7688, 0, 36, 9], [46, 2, 2, 1, 1, 5, 4, 0, 0, 1902, 52, 9], [32, 4, 5, 1, 1, 1, 4, 1, 0, 2824, 55, 9], [56, 4, 1, 1, 1, 0, 4, 1, 5178, 0, 44, 9], [40, 4, 2, 1, 1, 0, 4, 1, 0, 0, 40, 9], [45, 4, 5, 1, 1, 0, 4, 1, 15024, 0, 60, 9], [44, 2, 5, 1, 1, 0, 4, 1, 15024, 0, 35, 9], [58, 4, 1, 1, 1, 0, 4, 1, 7688, 0, 50, 9], [40, 5, 5, 1, 1, 5, 4, 0, 15024, 0, 30, 6]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[52, 7, 1, 1, 1, 0, 4, 1, 0, 0, 40, 9], [55, 4, 1, 1, 1, 0, 4, 1, 0, 0, 40, 9], [32, 6, 1, 1, 1, 4, 2, 0, 0, 0, 30, 9], [24, 4, 1, 1, 1, 1, 4, 1, 0, 0, 55, 9], [44, 4, 1, 1, 1, 5, 4, 0, 0, 0, 20, 9], [34, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [39, 4, 1, 1, 1, 4, 4, 0, 0, 0, 50, 9], [40, 6, 1, 1, 1, 0, 4, 1, 0, 0, 45, 9], [31, 2, 1, 1, 1, 1, 4, 1, 0, 0, 45, 9], [43, 4, 5, 1, 1, 4, 4, 0, 0, 2547, 40, 9]], \"covered_false\": [[51, 0, 2, 1, 1, 1, 4, 1, 0, 2824, 40, 9], [46, 4, 2, 1, 1, 0, 4, 1, 0, 2415, 55, 9], [44, 5, 5, 1, 1, 0, 4, 1, 0, 2415, 55, 9], [68, 7, 2, 1, 1, 0, 4, 1, 0, 2377, 60, 9], [61, 7, 2, 1, 1, 0, 4, 1, 0, 0, 40, 0]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[47, 4, 1, 1, 1, 0, 4, 1, 0, 0, 40, 8], [32, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [38, 1, 5, 1, 1, 0, 4, 1, 0, 0, 40, 6], [31, 4, 1, 1, 1, 1, 4, 0, 0, 0, 42, 9], [37, 4, 1, 1, 1, 0, 4, 1, 0, 0, 45, 9], [47, 4, 1, 1, 1, 0, 4, 1, 0, 0, 50, 9], [22, 7, 1, 1, 1, 0, 4, 1, 0, 0, 20, 9], [66, 4, 1, 1, 1, 0, 4, 1, 0, 0, 99, 9], [23, 4, 1, 1, 1, 3, 2, 1, 0, 0, 10, 9], [40, 4, 1, 1, 1, 0, 4, 1, 0, 0, 45, 9]], \"covered_false\": [[33, 4, 2, 1, 1, 0, 4, 1, 0, 0, 60, 9], [47, 1, 2, 1, 1, 0, 4, 1, 0, 0, 40, 9], [46, 4, 1, 1, 1, 1, 2, 0, 0, 0, 40, 9], [52, 5, 2, 1, 1, 0, 4, 1, 0, 0, 65, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Occupation = Admin\", \"Capital Gain <= 0.00\", \"Capital Loss <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
    			]
    		}
    	]
    }



```python
!kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "income" deleted
    model.mlops.seldon.io "income-explainer" deleted


## Custom Server


```python
!cat ./servers/custom-mlserver.yaml
```

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



```python
!kubectl create -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```

    server.mlops.seldon.io/mlserver-custom created



```python
!kubectl wait --for condition=ready --timeout=300s server --all -n ${NAMESPACE}
```

    server.mlops.seldon.io/mlserver condition met
    server.mlops.seldon.io/mlserver-custom condition met
    server.mlops.seldon.io/triton condition met



```python
!cat ./models/iris-custom-server.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      server: mlserver-custom



```python
!kubectl create -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris condition met



```python
!seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

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



```python
!kubectl delete -f ./models/iris-custom-server.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "iris" deleted



```python
!kubectl delete -f ./servers/custom-mlserver.yaml -n ${NAMESPACE}
```

    server.mlops.seldon.io "mlserver-custom" deleted



```python

```
