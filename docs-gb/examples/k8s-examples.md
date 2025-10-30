---
---

# Kubernetes examples

{% hint style="info" %}
**Note**:  The Seldon CLI allows you to view information about underlying Seldon resources and make changes to them through the scheduler in non-Kubernetes environments. However, it cannot modify underlying manifests within a Kubernetes cluster. Therefore, using the Seldon CLI for control plane operations in a Kubernetes environment is not recommended. For more details, see [Seldon CLI](../getting-started/cli.md).
{% endhint %}

## Before you begin

1. Ensure that you have [installed Seldon Core 2](../installation/production-environment/README.md#installing-seldon-core-2) in the namespace `seldon-mesh`.

2. Ensure that you are performing these steps in the directory where you have downloaded the [samples](https://github.com/SeldonIO/seldon-core/tree/v2/samples).

3. Get the IP address of the Seldon Core 2 instance running with Istio:

  ```bash
  ISTIO_INGRESS=$(kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

  echo "Seldon Core 2: http://$ISTIO_INGRESS"
  ```
  {% hint style="info" %}
  Make a note of the IP address that is displayed in the output. Replace <INGRESS_IP> with your service mesh's ingress IP address in the following commands.
  {% endhint %}

#### Create a Model

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
kubectl create -f ./models/sklearn-iris-gs.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/iris condition met

```

```bash
kubectl get model iris -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T10:01:52Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-06-30T10:01:52Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}
Make a REST inference call
```
{% tabs %}
{% tab title="curl" %}
```bash
curl -k http://<INGRESS_IP>:80/v2/models/iris/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: iris" \
  -d '{
    "inputs": [
      {
        "name": "predict",
        "shape": [1, 4],
        "datatype": "FP32",
        "data": [[1, 2, 3, 4]]
      }
    ]
  }' | jq

```

{% endtab %}

{% tab title="seldon-cli" %}
```bash
seldon model infer iris --inference-host <INGRESS_IP>:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

{% endtab %}
{% endtabs %}

Output is similar to:

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "7fd401e1-3dce-46f5-9668-902aea652b89",
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
Make a gRPC inference call

```bash
seldon model infer iris --inference-mode grpc --inference-host <INGRESS_IP>:80 \
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
kubectl get server mlserver -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T09:59:12Z",
      "status": "True",
      "type": "Ready"
    },
    {
      "lastTransitionTime": "2023-06-30T09:59:12Z",
      "reason": "StatefulSet replicas matches desired replicas",
      "status": "True",
      "type": "StatefulSetReady"
    }
  ],
  "loadedModels": 1,
  "replicas": 1,
  "selector": "seldon-server-name=mlserver"
}

```

Delete the model

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io "iris" deleted

```

#### Experiment

```bash
cat ./models/sklearn1.yaml
```

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

```bash
cat ./models/sklearn2.yaml
```

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

```bash
kubectl create -f ./models/sklearn1.yaml -n seldon-mesh
kubectl create -f ./models/sklearn2.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/iris created
model.mlops.seldon.io/iris2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/iris condition met
model.mlops.seldon.io/iris2 condition met

```

```bash
cat ./experiments/ab-default-model.yaml
```

```yaml
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

```

```bash
kubectl create -f ./experiments/ab-default-model.yaml -n seldon-mesh
```

```outputs
experiment.mlops.seldon.io/experiment-sample created

```

```bash
kubectl wait --for condition=ready --timeout=300s experiment --all -n seldon-mesh
```

```outputs
experiment.mlops.seldon.io/experiment-sample condition met

```

{% tabs %}
{% tab title="curl" %}
```bash
for i in {1..10}; do 
  curl -s -k <INGRESS_IP>:80/v2/models/experiment-sample/infer \
    -H "Host: seldon-mesh.inference.seldon" \
    -H "Content-Type: application/json" \
    -H "Seldon-Model: experiment-sample.experiment" \
    -d '{"inputs":[{"name":"predict","shape":[1,4],"datatype":"FP32","data":[[1,2,3,4]]}]}' \
    | jq -r .model_name
done | sort | uniq -c

```
```outputs
 4 iris2_1
 6 iris_1

```
{% endtab %}

{% tab title="seldon-cli" %}
```bash
seldon model infer --inference-host <INGRESS_IP>:80 -i 10 iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```
```outputs
Success: map[:iris2_1::4 :iris_1::6]

```

{% endtab %}
{% endtabs %}


```bash
kubectl delete -f ./experiments/ab-default-model.yaml -n seldon-mesh
kubectl delete -f ./models/sklearn1.yaml -n seldon-mesh
kubectl delete -f ./models/sklearn2.yaml -n seldon-mesh
```

```outputs
experiment.mlops.seldon.io "experiment-sample" deleted
model.mlops.seldon.io "iris" deleted
model.mlops.seldon.io "iris2" deleted

```

#### Pipeline - model chain

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```

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

```bash
kubectl create -f ./models/tfsimple1.yaml -n seldon-mesh
kubectl create -f ./models/tfsimple2.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met

```

```bash
cat ./pipelines/tfsimples.yaml
```

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

```bash
kubectl create -f ./pipelines/tfsimples.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimples condition met

```
{% tabs %}
{% tab title="curl" %}
```bash
curl -k <INGRESS_IP>:80/v2/models/tfsimples/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: tfsimples.pipeline" \
  -d '{
    "inputs": [
      {
        "name": "INPUT0",
        "datatype": "INT32",
        "shape": [1, 16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      },
      {
        "name": "INPUT1",
        "datatype": "INT32",
        "shape": [1, 16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      }
    ]
  }' |jq 

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
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
      ],
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    },
    {
      "data": [
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
      ],
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    }
  ]
}
```

{% endtab %}

{% tab title="seldon-cli" %}
```bash
seldon pipeline infer tfsimples --inference-mode grpc --inference-host <INGRESS_IP>:80 \
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

{% endtab %}
{% endtabs %}


```bash
kubectl delete -f ./pipelines/tfsimples.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io "tfsimples" deleted

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n seldon-mesh
kubectl delete -f ./models/tfsimple2.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted

```

#### Pipeline - model join

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
cat ./models/tfsimple3.yaml
```

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

```bash
kubectl create -f ./models/tfsimple1.yaml -n seldon-mesh
kubectl create -f ./models/tfsimple2.yaml -n seldon-mesh
kubectl create -f ./models/tfsimple3.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created
model.mlops.seldon.io/tfsimple3 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met
model.mlops.seldon.io/tfsimple3 condition met

```

```bash
cat ./pipelines/tfsimples-join.yaml
```

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

```bash
kubectl create -f ./pipelines/tfsimples-join.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/join created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/join condition met

```
{% tabs %}

{% tab title="curl" %} 
```bash
curl -k <INGRESS_IP>:80/v2/models/join/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: join.pipeline" \
  -d '{
    "model_name": "simple",
    "inputs": [
      {
        "name": "INPUT0",
        "datatype": "INT32",
        "shape": [1, 16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      },
      {
        "name": "INPUT1",
        "datatype": "INT32",
        "shape": [1, 16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      }
    ]
  }' |jq

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
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
      ],
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    },
    {
      "data": [
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
      ],
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    }
  ]
}
```

{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer join --inference-mode grpc --inference-host <INGRESS_IP>:80 \
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
{% endtab %}


{% endtabs %}


```bash
kubectl delete -f ./pipelines/tfsimples-join.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io "join" deleted

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n seldon-mesh
kubectl delete -f ./models/tfsimple2.yaml -n seldon-mesh
kubectl delete -f ./models/tfsimple3.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
model.mlops.seldon.io "tfsimple3" deleted

```

### Explainer

```bash
cat ./models/income.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/classifier"
  requirements:
  - sklearn

```

```bash
kubectl create -f ./models/income.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/income created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/income condition met

```

```bash
kubectl get model income -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```

```json
{
  "availableReplicas": 1,
  "conditions": [
    {
      "lastTransitionTime": "2025-10-30T08:37:24Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2025-10-30T08:37:24Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "modelgwReady": "ModelAvailable(1/1 ready ) ",
  "replicas": 1,
  "selector": "server=mlserver"
}
```

```bash
seldon model infer income --inference-host <INGRESS_IP>:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "cdf32df2-eb42-42d8-9f66-404bcab95540",
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
				0
			]
		}
	]
}

```

```bash
cat ./models/income-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/explainer"
  explainer:
    type: anchor_tabular
    modelRef: income

```

```bash
kubectl create -f ./models/income-explainer.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/income-explainer created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/income condition met
model.mlops.seldon.io/income-explainer condition met

```

```bash
kubectl get model income-explainer -n seldon-mesh -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T10:03:07Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-06-30T10:03:07Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income-explainer --inference-host <INGRESS_IP>:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "3028a904-9bb3-42d7-bdb7-6e6993323ed7",
	"parameters": {},
	"outputs": [
		{
			"name": "explanation",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str"
			},
			"data": [
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.1\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"precision\": 0.9705882352941176, \"coverage\": 0.0699, \"raw\": {\"feature\": [3, 5], \"mean\": [0.8094218415417559, 0.9705882352941176], \"precision\": [0.8094218415417559, 0.9705882352941176], \"coverage\": [0.3036, 0.0699], \"examples\": [{\"covered_true\": [[23, 4, 1, 1, 5, 1, 4, 0, 0, 0, 40, 9], [44, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [60, 2, 5, 1, 5, 1, 4, 0, 0, 0, 25, 9], [52, 4, 1, 1, 2, 0, 4, 1, 0, 0, 50, 9], [66, 6, 1, 1, 8, 0, 4, 1, 0, 0, 8, 9], [52, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [27, 4, 1, 1, 1, 1, 4, 1, 0, 0, 35, 9], [48, 4, 1, 1, 6, 0, 4, 1, 0, 0, 45, 9], [45, 6, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [40, 2, 1, 1, 5, 4, 4, 0, 0, 0, 45, 9]], \"covered_false\": [[42, 6, 5, 1, 6, 0, 4, 1, 99999, 0, 80, 9], [29, 4, 1, 1, 8, 1, 4, 1, 0, 0, 50, 9], [49, 4, 1, 1, 8, 0, 4, 1, 0, 0, 50, 9], [34, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [38, 2, 1, 1, 5, 5, 4, 0, 7688, 0, 40, 9], [45, 7, 5, 1, 5, 0, 4, 1, 0, 0, 45, 9], [43, 4, 2, 1, 5, 0, 4, 1, 99999, 0, 55, 9], [47, 4, 5, 1, 6, 1, 4, 1, 27828, 0, 60, 9], [42, 6, 1, 1, 2, 0, 4, 1, 15024, 0, 60, 9], [56, 4, 1, 1, 6, 0, 2, 1, 7688, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[23, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [50, 2, 5, 1, 8, 3, 2, 1, 0, 0, 45, 9], [24, 4, 1, 1, 7, 3, 4, 0, 0, 0, 40, 3], [62, 4, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [44, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [46, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [44, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [25, 4, 5, 1, 5, 3, 4, 1, 0, 0, 35, 9], [32, 2, 5, 1, 5, 3, 4, 1, 0, 0, 50, 9]], \"covered_false\": [[57, 5, 5, 1, 6, 3, 4, 1, 99999, 0, 40, 9], [44, 4, 1, 1, 8, 3, 4, 1, 7688, 0, 60, 9], [43, 2, 5, 1, 4, 3, 2, 0, 8614, 0, 47, 9], [56, 5, 2, 1, 5, 3, 4, 1, 99999, 0, 70, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
			]
		}
	]
}

```

```bash
kubectl delete -f ./models/income.yaml -n seldon-mesh
kubectl delete -f ./models/income-explainer.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io "income" deleted
model.mlops.seldon.io "income-explainer" deleted

```
