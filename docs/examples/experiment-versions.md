# Seldon V2 Experiment Version Tests

This notebook will show how we can update running experiments.

### Test change candidate for a model

We will use three SKlearn Iris classification models to illustrate experiment updates.

Load all models.

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

Let's call all three models individually first.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris3_1::50]

```

We will start an experiment to change the iris endpoint to split traffic with the `iris2` model.

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

Now when we call the iris model we should see a roughly 50/50 split between the two models.

```bash
seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::48 :iris_1::52]

```

Now we update the experiment to change to a split with the `iris3` model.

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

Now we should see a split with the `iris3` model.

```bash
seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris3_1::42 :iris_1::58]

```

```bash
seldon experiment stop experiment-sample
```

```json
{}

```

Now the experiment has been stopped we check everything as before.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris3_1::50]

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

Here we test changing the model we want to split traffic on. We will use three SKlearn Iris classification models to illustrate.

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

Let's call all three models to verify initial conditions.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris3_1::50]

```

Now we start an experiment to change calls to the `iris` model to split with the `iris2` model.

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
seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::51 :iris_1::49]

```

Now let's change the model we want to experiment to modify to the `iris3` model. Splitting between that and `iris2`.

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

Let's check the iris model is now as before but the iris3 model has traffic split.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::25 :iris3_1::25]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

```bash
seldon experiment stop experiment-sample
```

```json
{}

```

Finally, let's check now the experiment has stopped as is as at the start.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

```bash
seldon model infer iris3 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris3_1::50]

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
