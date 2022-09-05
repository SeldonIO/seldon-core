## Seldon V2 Experiment Version Tests


### Test change candidate for a model

We will use three SKlearn Iris classification models to illustrate experiments.

Load both models.


```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
seldon model load -f ./models/sklearn3.yaml
```
```json

    {}
    {}
    {}
```

```bash
seldon model status iris -w ModelAvailable
seldon model status iris2 -w ModelAvailable
seldon model status iris3 -w ModelAvailable
```
```json

    {}
    {}
    {}
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris_1::50]
```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::50]
```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris3_1::50]
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
seldon experiment start -f ./experiments/ab-default-model.yaml 
```
```json

    {}
```

```bash
seldon experiment status experiment-sample -w | jq -M .
```
```json

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::26 :iris_1::24]
```

```bash
cat ./experiments/ab-default-model2.yaml 
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
      - name: iris3
        weight: 50
```

```bash
seldon experiment start -f ./experiments/ab-default-model2.yaml 
```
```json

    {}
```

```bash
seldon experiment status experiment-sample -w | jq -M .
```
```json

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris3_1::24 :iris_1::26]
```

```bash
seldon experiment stop experiment-sample
```
```json

    {}
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris_1::50]
```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::50]
```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris3_1::50]
```

```bash
seldon model unload iris
seldon model unload iris2
seldon model unload iris3
```
```json

    {}
    {}
    {}
```
### Test change default model in an experiment

We will use three SKlearn Iris classification models to illustrate experiments.


```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
seldon model load -f ./models/sklearn3.yaml
```
```json

    {}
    {}
    {}
```

```bash
seldon model status iris -w ModelAvailable
seldon model status iris2 -w ModelAvailable
seldon model status iris3 -w ModelAvailable
```
```json

    {}
    {}
    {}
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris_1::50]
```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::50]
```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris3_1::50]
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
seldon experiment start -f ./experiments/ab-default-model.yaml 
```
```json

    {}
```

```bash
seldon experiment status experiment-sample -w | jq -M .
```
```json

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }
```
Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.


```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::28 :iris_1::22]
```

```bash
cat ./experiments/ab-default-model3.yaml 
```
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
    spec:
      default: iris3
      candidates:
      - name: iris3
        weight: 50
      - name: iris2
        weight: 50
```

```bash
seldon experiment start -f ./experiments/ab-default-model3.yaml 
```
```json

    {}
```

```bash
seldon experiment status experiment-sample -w | jq -M .
```
```json

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris_1::50]
```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::24 :iris3_1::26]
```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::50]
```

```bash
seldon experiment stop experiment-sample
```
```json

    {}
```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris_1::50]
```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris2_1::50]
```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json

    map[:iris3_1::50]
```

```bash
seldon model unload iris
seldon model unload iris2
seldon model unload iris3
```
```json

    {}
    {}
    {}
```

```python

```
