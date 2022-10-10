### Seldon V2 Batch Examples


```python
import os
os.environ["NAMESPACE"] = "seldon"
```


```python
MESH_IP=kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```
```bash



    '172.19.255.12'
```

## Deploy Models and Pipelines


```bash
cat models/sklearn-iris-gs.yaml
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
      memory: 100Ki
```

```bash
cat pipelines/iris.yaml
```
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: iris-pipeline
    spec:
      steps:
        - name: iris
      output:
        steps:
        - iris
```

```bash
cat models/tfsimple1.yaml
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
```

```bash
cat pipelines/tfsimple.yaml
```
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1
```

```bash
kubectl apply -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
kubectl apply -f pipelines/iris.yaml -n ${NAMESPACE}

kubectl apply -f models/tfsimple1.yaml -n ${NAMESPACE}
kubectl apply -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```
```json

    model.mlops.seldon.io/iris unchanged
    pipeline.mlops.seldon.io/iris-pipeline unchanged
    model.mlops.seldon.io/tfsimple1 unchanged
    pipeline.mlops.seldon.io/tfsimple unchanged
```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
kubectl wait --for condition=ready --timeout=300s pipelines --all -n ${NAMESPACE}

kubectl wait --for condition=ready --timeout=300s model tfsimple1 -n ${NAMESPACE}
kubectl wait --for condition=ready --timeout=300s pipelines tfsimple -n ${NAMESPACE}
```
```json

    model.mlops.seldon.io/iris condition met
    model.mlops.seldon.io/tfsimple condition met
    model.mlops.seldon.io/tfsimple1 condition met
    pipeline.mlops.seldon.io/iris-pipeline condition met
    pipeline.mlops.seldon.io/tfsimple condition met
    pipeline.mlops.seldon.io/tfsimple-pipeline condition met
    model.mlops.seldon.io/tfsimple1 condition met
    pipeline.mlops.seldon.io/tfsimple condition met
```
## Test Predictions


```bash
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' | jq -M .
```
```json

    {
      "model_name": "iris_1",
      "model_version": "1",
      "id": "5b88fb3f-b467-4c36-b172-0ca81ecfab2e",
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

```bash
seldon pipeline infer iris-pipeline --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' |  jq -M .
```
```json

    {
      "model_name": "",
      "outputs": [
        {
          "data": [
            2
          ],
          "name": "predict",
          "shape": [
            1
          ],
          "datatype": "INT64"
        }
      ]
    }
```

```bash
seldon model infer tfsimple --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq .
```
```json

    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m"tfsimple_1"[0m[1;39m,
      [0m[34;1m"model_version"[0m[1;39m: [0m[0;32m"1"[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT0"[0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m2[0m[1;39m,
            [0;39m4[0m[1;39m,
            [0;39m6[0m[1;39m,
            [0;39m8[0m[1;39m,
            [0;39m10[0m[1;39m,
            [0;39m12[0m[1;39m,
            [0;39m14[0m[1;39m,
            [0;39m16[0m[1;39m,
            [0;39m18[0m[1;39m,
            [0;39m20[0m[1;39m,
            [0;39m22[0m[1;39m,
            [0;39m24[0m[1;39m,
            [0;39m26[0m[1;39m,
            [0;39m28[0m[1;39m,
            [0;39m30[0m[1;39m,
            [0;39m32[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m
```

```bash
seldon pipeline infer tfsimple-pipeline --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq .
```
```json

    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m""[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m2[0m[1;39m,
            [0;39m4[0m[1;39m,
            [0;39m6[0m[1;39m,
            [0;39m8[0m[1;39m,
            [0;39m10[0m[1;39m,
            [0;39m12[0m[1;39m,
            [0;39m14[0m[1;39m,
            [0;39m16[0m[1;39m,
            [0;39m18[0m[1;39m,
            [0;39m20[0m[1;39m,
            [0;39m22[0m[1;39m,
            [0;39m24[0m[1;39m,
            [0;39m26[0m[1;39m,
            [0;39m28[0m[1;39m,
            [0;39m30[0m[1;39m,
            [0;39m32[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT0"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m
        [1;39m}[0m[1;39m,
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m,
            [0;39m0[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT1"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m
```
## MLServer Iris Batch Job


```bash
cat batch-inputs/iris-input.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"predict"[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m0.38606369295833043[0m[1;39m,
            [0;39m0.006894049558299753[0m[1;39m,
            [0;39m0.6104082981607108[0m[1;39m,
            [0;39m0.3958954239450676[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"FP64"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m4[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m
```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris -i batch-inputs/iris-input.txt -o /tmp/iris-output.txt --workers 5
```
```bash
    2022-10-10 08:40:18,854 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 08:40:18,857 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 08:40:18,857 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
    2022-10-10 08:40:18,857 [mlserver] INFO - output file path: /tmp/iris-output.txt
    2022-10-10 08:40:19,066 [mlserver] INFO - Time taken: 0.21 seconds
```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris-pipeline.pipeline -i batch-inputs/iris-input.txt -o /tmp/iris-pipeline-output.txt --workers 5
```
```bash
    2022-10-10 08:40:20,849 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 08:40:20,852 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 08:40:20,852 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
    2022-10-10 08:40:20,852 [mlserver] INFO - output file path: /tmp/iris-pipeline-output.txt
    2022-10-10 08:40:21,307 [mlserver] INFO - Time taken: 0.45 seconds
```

```bash
cat /tmp/iris-output.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m"iris_1"[0m[1;39m,
      [0m[34;1m"model_version"[0m[1;39m: [0m[0;32m"1"[0m[1;39m,
      [0m[34;1m"id"[0m[1;39m: [0m[0;32m"08359080-6fe2-4429-8c37-cc6276f4340b"[0m[1;39m,
      [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{
        [0m[34;1m"content_type"[0m[1;39m: [0m[1;30mnull[0m[1;39m,
        [0m[34;1m"headers"[0m[1;39m: [0m[1;30mnull[0m[1;39m,
        [0m[34;1m"batch_index"[0m[1;39m: [0m[0;39m0[0m[1;39m
      [1;39m}[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"predict"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT64"[0m[1;39m,
          [0m[34;1m"parameters"[0m[1;39m: [0m[1;30mnull[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m0[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m
```

```bash
cat /tmp/iris-pipeline-output.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m""[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m0[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"predict"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT64"[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m,
      [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{
        [0m[34;1m"batch_index"[0m[1;39m: [0m[0;39m0[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m}[0m
```
## Triton TFSimple Batch Job


```bash
cat batch-inputs/tfsimple-input.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"INPUT0"[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m75[0m[1;39m,
            [0;39m39[0m[1;39m,
            [0;39m9[0m[1;39m,
            [0;39m44[0m[1;39m,
            [0;39m32[0m[1;39m,
            [0;39m97[0m[1;39m,
            [0;39m99[0m[1;39m,
            [0;39m40[0m[1;39m,
            [0;39m13[0m[1;39m,
            [0;39m27[0m[1;39m,
            [0;39m25[0m[1;39m,
            [0;39m36[0m[1;39m,
            [0;39m18[0m[1;39m,
            [0;39m77[0m[1;39m,
            [0;39m62[0m[1;39m,
            [0;39m60[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m,
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"INPUT1"[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m39[0m[1;39m,
            [0;39m7[0m[1;39m,
            [0;39m14[0m[1;39m,
            [0;39m58[0m[1;39m,
            [0;39m13[0m[1;39m,
            [0;39m88[0m[1;39m,
            [0;39m98[0m[1;39m,
            [0;39m66[0m[1;39m,
            [0;39m97[0m[1;39m,
            [0;39m57[0m[1;39m,
            [0;39m49[0m[1;39m,
            [0;39m3[0m[1;39m,
            [0;39m49[0m[1;39m,
            [0;39m63[0m[1;39m,
            [0;39m37[0m[1;39m,
            [0;39m12[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m
```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-output.txt --workers 5
```
```bash
    2022-10-10 08:40:23,543 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 08:40:23,546 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 08:40:23,546 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
    2022-10-10 08:40:23,546 [mlserver] INFO - output file path: /tmp/tfsimple-output.txt
    2022-10-10 08:40:23,733 [mlserver] INFO - Time taken: 0.19 seconds
```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple-pipeline.pipeline -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-pipeline-output.txt --workers 5
```
```bash
    2022-10-10 08:40:25,397 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 08:40:25,400 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 08:40:25,400 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
    2022-10-10 08:40:25,400 [mlserver] INFO - output file path: /tmp/tfsimple-pipeline-output.txt
    2022-10-10 08:40:25,752 [mlserver] INFO - Time taken: 0.35 seconds
```

```bash
cat /tmp/tfsimple-output.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"id"[0m[1;39m: [0m[0;32m"168eef39-0510-4729-b8df-3d0ec063df55"[0m[1;39m,
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m"tfsimple_1"[0m[1;39m,
      [0m[34;1m"model_version"[0m[1;39m: [0m[0;32m"1"[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT0"[0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{}[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m114[0m[1;39m,
            [0;39m46[0m[1;39m,
            [0;39m23[0m[1;39m,
            [0;39m102[0m[1;39m,
            [0;39m45[0m[1;39m,
            [0;39m185[0m[1;39m,
            [0;39m197[0m[1;39m,
            [0;39m106[0m[1;39m,
            [0;39m110[0m[1;39m,
            [0;39m84[0m[1;39m,
            [0;39m74[0m[1;39m,
            [0;39m39[0m[1;39m,
            [0;39m67[0m[1;39m,
            [0;39m140[0m[1;39m,
            [0;39m99[0m[1;39m,
            [0;39m72[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m,
        [1;39m{
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT1"[0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{}[0m[1;39m,
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m36[0m[1;39m,
            [0;39m32[0m[1;39m,
            [0;39m-5[0m[1;39m,
            [0;39m-14[0m[1;39m,
            [0;39m19[0m[1;39m,
            [0;39m9[0m[1;39m,
            [0;39m1[0m[1;39m,
            [0;39m-26[0m[1;39m,
            [0;39m-84[0m[1;39m,
            [0;39m-30[0m[1;39m,
            [0;39m-24[0m[1;39m,
            [0;39m33[0m[1;39m,
            [0;39m-31[0m[1;39m,
            [0;39m14[0m[1;39m,
            [0;39m25[0m[1;39m,
            [0;39m48[0m[1;39m
          [1;39m][0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m,
      [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{
        [0m[34;1m"batch_index"[0m[1;39m: [0m[0;39m0[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m}[0m
```

```bash
cat /tmp/tfsimple-pipeline-output.txt | head -n 1 | jq .
```
```yaml
    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m""[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m114[0m[1;39m,
            [0;39m46[0m[1;39m,
            [0;39m23[0m[1;39m,
            [0;39m102[0m[1;39m,
            [0;39m45[0m[1;39m,
            [0;39m185[0m[1;39m,
            [0;39m197[0m[1;39m,
            [0;39m106[0m[1;39m,
            [0;39m110[0m[1;39m,
            [0;39m84[0m[1;39m,
            [0;39m74[0m[1;39m,
            [0;39m39[0m[1;39m,
            [0;39m67[0m[1;39m,
            [0;39m140[0m[1;39m,
            [0;39m99[0m[1;39m,
            [0;39m72[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT0"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m
        [1;39m}[0m[1;39m,
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m36[0m[1;39m,
            [0;39m32[0m[1;39m,
            [0;39m-5[0m[1;39m,
            [0;39m-14[0m[1;39m,
            [0;39m19[0m[1;39m,
            [0;39m9[0m[1;39m,
            [0;39m1[0m[1;39m,
            [0;39m-26[0m[1;39m,
            [0;39m-84[0m[1;39m,
            [0;39m-30[0m[1;39m,
            [0;39m-24[0m[1;39m,
            [0;39m33[0m[1;39m,
            [0;39m-31[0m[1;39m,
            [0;39m14[0m[1;39m,
            [0;39m25[0m[1;39m,
            [0;39m48[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"OUTPUT1"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m,
            [0;39m16[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT32"[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m,
      [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{
        [0m[34;1m"batch_index"[0m[1;39m: [0m[0;39m0[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m}[0m
```
## Cleanup


```bash
kubectl delete -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
kubectl delete -f pipelines/iris.yaml -n ${NAMESPACE}
```
```json

    model.mlops.seldon.io "iris" deleted
    pipeline.mlops.seldon.io "iris-pipeline" deleted
```

```bash
kubectl delete -f models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```
```json

    model.mlops.seldon.io "tfsimple1" deleted
    pipeline.mlops.seldon.io "tfsimple" deleted
```