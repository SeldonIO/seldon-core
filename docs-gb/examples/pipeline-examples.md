# Pipeline examples

These examples illustrates a series of Pipelines showing of different ways of combining flows of\
data and conditional logic. We assume you have Seldon Core 2 running locally.

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

### Models Used

* `gs://seldon-models/triton/simple` an example Triton tensorflow model that takes 2 inputs INPUT0\
  and INPUT1 and adds them to produce OUTPUT0 and also subtracts INPUT1 from INPUT0 to produce OUTPUT1.\
  See [here](https://github.com/triton-inference-server/server/tree/main/docs/examples/model_repository/simple)\
  for the original source code and license.
* Other models can be found at https://github.com/SeldonIO/triton-python-examples

### Model Chaining

Chain the output of one model into the next. Also shows chaning the tensor names via `tensorMap` to conform to the expected input tensor names of the second model.

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
---  
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

```output
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

This pipeline chains the output of `tfsimple1` into `tfsimple2`. As these models have compatible shape and data type this can be done. However, the output tensor names from `tfsimple1` need to be renamed to match the input tensor names for `tfsimple2`. You can do this with the `tensorMap` feature.

The output of the Pipeline is the output from `tfsimple2`.

```bash
cat ./pipelines/tfsimples.yaml
```

{% embed url="https://github.com/SeldonIO/seldon-core/blob/v2/samples/pipelines/tfsimples.yaml" %}

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
curl -k http://<INGRESS_IP>:80/v2/models/tfsimples/infer \
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


You can use the Seldon CLI `pipeline inspect` feature to look at the data for all steps of the pipeline for the last data item passed through the pipeline (the default). This can be useful for debugging.

```bash
seldon pipeline inspect tfsimples
```

```bash
seldon.default.model.tfsimple1.inputs	ciep298fh5ss73dpdir0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
seldon.default.model.tfsimple1.outputs	ciep298fh5ss73dpdir0	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
seldon.default.model.tfsimple2.inputs	ciep298fh5ss73dpdir0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
seldon.default.model.tfsimple2.outputs	ciep298fh5ss73dpdir0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}
seldon.default.pipeline.tfsimples.inputs	ciep298fh5ss73dpdir0	{"modelName":"tfsimples", "inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
seldon.default.pipeline.tfsimples.outputs	ciep298fh5ss73dpdir0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}

```

Next, take a look get the output as json and use the `jq` tool to get just one value.

```bash
seldon pipeline inspect tfsimples --format json | jq -M .topics[0].msgs[0].value
```

```json
{
  "inputs": [
    {
      "name": "INPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
        ]
      }
    },
    {
      "name": "INPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
        ]
      }
    }
  ]
}

```

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

### Model Chaining from inputs

Chain the output of one model into the next. Shows using the input and outputs and combining.

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

```output
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
cat ./pipelines/tfsimples-input.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimples-input
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
      inputs:
      - tfsimple1.inputs.INPUT0
      - tfsimple1.outputs.OUTPUT1
      tensorMap:
        tfsimple1.outputs.OUTPUT1: INPUT1
  output:
    steps:
    - tfsimple2

```

```bash
kubectl create -f ./pipelines/tfsimples-input.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimples-input created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimples-input condition met
```

{% tabs %}

{% tab title="curl" %}
```bash
curl -k http://<INGRESS_IP>:80/v2/models/tfsimples-input/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: tfsimples-input.pipeline" \
  -d '{
    "model_name": "simple",
    "inputs": [
      {
        "name": "INPUT0",
        "datatype": "INT32",
        "shape": [1,16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      },
      {
        "name": "INPUT1",
        "datatype": "INT32",
        "shape": [1,16],
        "data": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
      }
    ]
  }' | jq -M .
```   

```bash
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
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
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
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
seldon pipeline infer tfsimples-input \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```    
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
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
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
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
{% endtabs %}


```bash
kubectl delete -f ./pipelines/tfsimples-input.yaml -n seldon-mesh
```
```outputs
pipeline.mlops.seldon.io "tfsimples-input" deleted
```
```bash
kubectl delete -f ./models/tfsimple1.yaml -n seldon-mesh
kubectl delete -f ./models/tfsimple2.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
```

### Model Join

Join two flows of data from two models as input to a third model. This shows how individual flows of data can be combined.

```bash
cat ./models/tfsimple1.yaml
echo "---"
cat ./models/tfsimple2.yaml
echo "---"
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
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
---
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
In this pipeline for the input to `tfsimple3` and join 1 output tensor each from the two previous models `tfsimple1` and `tfsimple2`. You need to use the `tensorMap` feature to rename each output tensor to one of the expected input tensors for the `tfsimple3` model.

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

The outputs are the sequence "2,4,6..." which conforms to the logic of this model (addition and subtraction) when fed the output of the first two models.

{% tabs %}

{% tab title="curl" %} 
```bash
curl -k http://<INGRESS_IP>:80/v2/models/join/infer \
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

### Conditional

Shows conditional data flows - one of two models is run based on output tensors from first.

```bash
cat ./models/conditional.yaml
echo "---"
cat ./models/add10.yaml
echo "---"
cat ./models/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: conditional
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/conditional"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python

```

```bash
kubectl create -f ./models/conditional.yaml -n seldon-mesh
kubectl create -f ./models/add10.yaml -n seldon-mesh
kubectl create -f ./models/mul10.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/conditional created
model.mlops.seldon.io/add10 created
model.mlops.seldon.io/mul10 created
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/conditional condition met
model.mlops.seldon.io/add10 condition met
model.mlops.seldon.io/mul10 condition met

```

Here we assume the `conditional` model can output two tensors OUTPUT0 and OUTPUT1 but only outputs the former if the CHOICE input tensor is set to 0 otherwise it outputs tensor OUTPUT1. By this means only one of the two downstream models will receive data and run. The `output` steps does an `any` join from both models and whichever data appears first will be sent as output to pipeline. As in this case only 1 of the two models `add10` and `mul10` runs we will receive their output.

```bash
cat ./pipelines/conditional.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple-conditional
spec:
  steps:
  - name: conditional
  - name: mul10
    inputs:
    - conditional.outputs.OUTPUT0
    tensorMap:
      conditional.outputs.OUTPUT0: INPUT
  - name: add10
    inputs:
    - conditional.outputs.OUTPUT1
    tensorMap:
      conditional.outputs.OUTPUT1: INPUT
  output:
    steps:
    - mul10
    - add10
    stepsJoin: any

```

```bash
kubectl create -f ./pipelines/conditional.yaml -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimple-conditional created
```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/tfsimple-conditional condition met
```

The `mul10` model runs as the CHOICE tensor is set to 0.

{% tabs %}

{% tab title="curl" %} 
```bash
curl -k http://<INGRESS_IP>:80/v2/models/tfsimple-conditional/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: tfsimple-conditional.pipeline" \
  -d '{
    "model_name": "conditional",
    "inputs": [
      {
        "name": "CHOICE",
        "datatype": "INT32",
        "shape": [1],
        "data": [0]
      },
      {
        "name": "INPUT0",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      },
      {
        "name": "INPUT1",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        10,
        20,
        30,
        40
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    }
  ]
}

```

{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc --inference-host <INGRESS_IP>:80 \
 '{"model_name":"conditional","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}

```
{% endtab %}


{% endtabs %}



The `add10` model will run as the CHOICE tensor is not set to zero.

{% tabs %}

{% tab title="curl" %} 
```bash
curl -k http://<INGRESS_IP>:80/v2/models/tfsimple-conditional/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: tfsimple-conditional.pipeline" \
  -d '{
    "model_name": "conditional",
    "inputs": [
      {
        "name": "CHOICE",
        "datatype": "INT32",
        "shape": [1],
        "data": [1]
      },
      {
        "name": "INPUT0",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      },
      {
        "name": "INPUT1",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        11,
        12,
        13,
        14
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    }
  ]
}

```

{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc --inference-host <INGRESS_IP>:80 \ 
 '{"model_name":"conditional","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```
{% endtab %}


{% endtabs %}



```bash
kubectl delete -f ./pipelines/conditional.yaml -n seldon-mesh
```

```bash
kubectl delete -f ./models/conditional.yaml -n seldon-mesh
kubectl delete -f ./models/add10.yaml -n seldon-mesh
kubectl delete -f ./models/mul10.yaml -n seldon-mesh
```

### Pipeline Input Tensors

Access to indivudal tensors in pipeline inputs

```bash
cat ./models/mul10.yaml
echo "---"
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python

```

```bash
kubectl create -f ./models/mul10.yaml -n seldon-mesh
kubectl create -f ./models/add10.yaml -n seldon-mesh
```

```outputs
model.mlops.seldon.io/mul10 created
model.mlops.seldon.io/add10 created
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/mul10 condition met
model.mlops.seldon.io/add10 condition met
```

This pipeline shows how we can access pipeline inputs INPUT0 and INPUT1 from different steps.

```bash
cat ./pipelines/pipeline-inputs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-inputs
spec:
  steps:
  - name: mul10
    inputs:
    - pipeline-inputs.inputs.INPUT0
    tensorMap:
      pipeline-inputs.inputs.INPUT0: INPUT
  - name: add10
    inputs:
    - pipeline-inputs.inputs.INPUT1
    tensorMap:
      pipeline-inputs.inputs.INPUT1: INPUT
  output:
    steps:
    - mul10
    - add10

```

```bash
kubectl create -f ./pipelines/pipeline-inputs.yaml -n seldon-mesh
```
```outputs
pipeline.mlops.seldon.io/pipeline-inputs created
```
```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/pipeline-inputs condition met
```

{% tabs %}

{% tab title="curl" %} 
```bash
curl -k http://<INGRESS_IP>:80/v2/models/pipeline-inputs/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: pipeline-inputs.pipeline" \
  -d '{
    "model_name": "pipeline",
    "inputs": [
      {
        "name": "INPUT0",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      },
      {
        "name": "INPUT1",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        10,
        20,
        30,
        40
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    },
    {
      "data": [
        11,
        12,
        13,
        14
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    }
  ]
}

```

{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer pipeline-inputs --inference-mode grpc --inference-host <INGRESS-IP>:80 \
    '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    },
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```
{% endtab %}


{% endtabs %}


```bash
kubectl delete -f ./pipelines/pipeline-inputs.yaml -n seldon-mesh
```

```bash
kubectl delete -f ./models/mul10.yaml -n seldon-mesh
kubectl delete -f ./models/add10.yaml -n seldon-mesh
```

### Trigger Joins

Shows how joins can be used for triggers as well.

```bash
cat ./models/mul10.yaml
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python

```

```bash
kubectl create -f ./models/mul10.yaml -n seldon-mesh
kubectl create -f ./models/add10.yaml -n seldon-mesh

```

```outputs
model.mlops.seldon.io/mul10 created
model.mlops.seldon.io/add10 created
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```outputs
model.mlops.seldon.io/mul10 condition met
model.mlops.seldon.io/add10 condition met
```

Here we required tensors names `ok1` or `ok2` to exist on pipeline inputs to run the `mul10` model but require tensor `ok3` to exist on pipeline inputs to run the `add10` model. The logic on `mul10` is handled by a trigger join of `any` meaning either of these input data can exist to satisfy the trigger join.

```bash
cat ./pipelines/trigger-joins.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: trigger-joins
spec:
  steps:
  - name: mul10
    inputs:
    - trigger-joins.inputs.INPUT
    triggers:
    - trigger-joins.inputs.ok1
    - trigger-joins.inputs.ok2
    triggersJoinType: any
  - name: add10
    inputs:
    - trigger-joins.inputs.INPUT
    triggers:
    - trigger-joins.inputs.ok3
  output:
    steps:
    - mul10
    - add10
    stepsJoin: any

```

```bash
kubectl create -f ./pipelines/trigger-joins.yaml -n seldon-mesh
```
```outputs
pipeline.mlops.seldon.io/trigger-joins created
```
```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n seldon-mesh
```

```outputs
pipeline.mlops.seldon.io/trigger-joins condition met
```

Trigger the first join.

{% tabs %}

{% tab title="curl" %}

```bash
curl -k http://<INGRESS_IP>:80/v2/models/trigger-joins/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: trigger-joins.pipeline" \
  -d '{
    "model_name": "pipeline",
    "inputs": [
      {
        "name": "ok1",
        "datatype": "FP32",
        "shape": [1],
        "data": [1]
      },
      {
        "name": "INPUT",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .
```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        10,
        20,
        30,
        40
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    }
  ]
}
```

{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer trigger-joins --inference-mode grpc --inference-host <INGRESS_IP>:80\
    '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}
```
{% endtab %}

{% endtabs %}

Now, you can trigger the second join.

{% tabs %}

{% tab title="curl" %} 

```bash
curl -k http://<INGRESS_IP>:80/v2/models/trigger-joins/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: trigger-joins.pipeline" \
  -d '{
    "model_name": "pipeline",
    "inputs": [
      {
        "name": "ok3",
        "datatype": "FP32",
        "shape": [1],
        "data": [1]
      },
      {
        "name": "INPUT",
        "datatype": "FP32",
        "shape": [4],
        "data": [1,2,3,4]
      }
    ]
  }' | jq -M .

```
```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        11,
        12,
        13,
        14
      ],
      "name": "OUTPUT",
      "shape": [
        4
      ],
      "datatype": "FP32"
    }
  ]
}

```
{% endtab %}

{% tab title="seldon-cli" %}

```bash
seldon pipeline infer trigger-joins --inference-mode grpc --inference-host <INGRESS_IP>:80 \
    '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```
{% endtab %}

{% endtabs %}

```bash
kubectl delete -f ./pipelines/trigger-joins.yaml -n seldon-mesh
```
```outputs
pipeline.mlops.seldon.io "trigger-joins" deleted
```
```bash
kubectl delete -f ./models/mul10.yaml -n seldon-mesh
kubectl delete -f ./models/add10.yaml -n seldon-mesh
```
```outputs
model.mlops.seldon.io "mul10" deleted
model.mlops.seldon.io "add10" deleted
```