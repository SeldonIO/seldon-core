# Local Batch Inference Example

This example runs you through a series of batch inference requests made to both models and pipelines running on Seldon Core locally. 

## Setup

If you haven't already, you'll need to [clone the Seldon Core repository and run it locally](../getting-started/docker-installation/index) before you run through this example.

```{important}
By default, the CLI will expect your inference endpoint to be at `0.0.0.0:9000`. If you have customized this, you'll need to [redirect the CLI](../cli/index).
```

## Deploy Models and Pipelines

First, let's jump in to the `samples` folder where we'll find some sample models and pipelines we can use:

```bash
cd samples/
```

### Deploy the Iris Model

Let's take a look at a sample model before we deploy it:

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

The above manifest will deploy a simple [sci-kit learn](https://scikit-learn.org/stable/) model based on the [iris dataset](https://archive.ics.uci.edu/ml/datasets/iris).

Let's now deploy that model using the Seldon CLI:

```bash
seldon model load -f models/sklearn-iris-gs.yaml
```

### Deploy the Iris Pipeline

Now that we've deployed our iris model, let's create a [pipeline](../pipelines/index) around the model. 

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

We see that this pipeline only has one step, which is to call the `iris` model we deployed earlier. We can create the pipeline by running:

```bash
seldon pipeline load -f pipelines/iris.yaml
```

### Deploy the Tensorflow Model

To demonstrate batch inference requests to different types of models, we'll also deploy a simple [tensorflow](https://www.tensorflow.org/) model:

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

The tensorflow model takes two arrays as inputs and returns two arrays as outputs. The first output is the addition of the two inputs and the second output is the value of (first input - second input).

Let's deploy the model:

```bash
seldon model load -f models/tfsimple1.yaml
```

### Deploy the Tensorflow Pipeline

Just as we did for the scikit-learn model, we'll deploy a simple pipeline for our tensorflow model:

Inspect the pipeline manifest:
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
and deploy it:
```bash
seldon pipeline load -f pipelines/tfsimple.yaml
```

### Check Model and Pipeline Status

Once we've deployed a model or pipeline to Seldon Core, we can list them and check their status by running:

```bash
seldon model list
```
and
```bash
seldon pipeline list
```

Your models and pieplines should be showing a state of `ModelAvailable` and `PipelineReady` respectively.

## Test Predictions

Before we run a large batch job of predictions through our models and pipelines, let's quickly check that they work with a single standalone inference request. We can do this using the `seldon model infer` command.

### Scikit-learn Model
```bash
seldon model infer iris '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' | jq
```
```json
{
  "model_name": "iris_1",
  "model_version": "1",
  "id": "a67233c2-2f8c-4fbc-a87e-4e4d3d034c9f",
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

The preidiction request body needs to be an [Open Inference Protocol](../apis/inference/v2.md) compatible payload and also match the expected inputs for the model you've deployed. In this case, the iris model expects data of shape `[1, 4]` and of type `FP32`. 

You'll notice that the prediction results for this request come back on `outputs[0].data`.

### Scikit-learn Pipeline

```bash
seldon pipeline infer iris-pipeline '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' |  jq
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

### Tensorflow Model

```bash
seldon model infer tfsimple1 '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq
```
```json
{
  "model_name": "tfsimple1_1",
  "model_version": "1",
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        1,
        16
      ],
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
      ]
    }
  ]
}
```

You'll notice that the inputs for our tensorflow model look different from the ones we sent to the iris model. This time, we're sending two arrays of shape `[1,16]`. When sending an inference request, we can optionally chose which outputs we want back by including an `{"outputs":...}` object. 

### Tensorflow Pipeline

```bash
seldon pipeline infer tfsimple '"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq
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
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
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

## Running the Scikit-Learn Batch Job

In the samples folder there is a batch request input file: `batch-inputs/iris-input.txt`. It contains 100 input payloads for our iris model. Let's take a look at the first line in that file:

```bash
cat batch-inputs/iris-input.txt | head -n 1 | jq
```
```json
{
  "inputs": [
    {
      "name": "predict",
      "data": [
        0.38606369295833043,
        0.006894049558299753,
        0.6104082981607108,
        0.3958954239450676
      ],
      "datatype": "FP64",
      "shape": [
        1,
        4
      ]
    }
  ]
}

```

To run a batch inference job we'll use the [MLServer CLI](https://mlserver.readthedocs.io/en/latest/reference/cli.html). If you don't already have it installed you can install it using:

```bash
pip install mlserver
```
### Iris Model

The inference job can be executed by running the following command:
```bash
mlserver infer -u localhost:9000 -m iris -i batch-inputs/iris-input.txt -o /tmp/iris-output.txt --workers 5
```
```output
2023-01-22 18:24:17,272 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-22 18:24:17,273 [mlserver] INFO - server url: localhost:9000
2023-01-22 18:24:17,273 [mlserver] INFO - model name: iris
2023-01-22 18:24:17,273 [mlserver] INFO - request headers: {}
2023-01-22 18:24:17,273 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-01-22 18:24:17,273 [mlserver] INFO - output file path: /tmp/iris-output.txt
2023-01-22 18:24:17,273 [mlserver] INFO - workers: 5
2023-01-22 18:24:17,273 [mlserver] INFO - retries: 3
2023-01-22 18:24:17,273 [mlserver] INFO - batch interval: 0.0
2023-01-22 18:24:17,274 [mlserver] INFO - batch jitter: 0.0
2023-01-22 18:24:17,274 [mlserver] INFO - connection timeout: 60
2023-01-22 18:24:17,274 [mlserver] INFO - micro-batch size: 1
2023-01-22 18:24:17,420 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-22 18:24:17,421 [mlserver] INFO - Total processed instances: 100
2023-01-22 18:24:17,421 [mlserver] INFO - Time taken: 0.15 seconds

```

The mlserver batch component will take your input file `batch-inputs/iris-input.txt`, distribute those payloads across 5 different workers (`--workers 5`), collect the responses and write them to a file `/tmp/iris-output.txt`. For a full set of options check out the [MLServer CLI Reference](https://mlserver.readthedocs.io/en/latest/reference/cli.html#mlserver-infer).

#### Checking the Output

We can check the inference responses by looking at the contents of the output file:
```bash
cat /tmp/iris-output.txt | head -n 1 | jq
```

### Iris Pipeline

We can run the same batch job for our iris pipeline and store the outputs in a different file:

```bash
mlserver infer -u localhost:9000 -m iris-pipeline.pipeline -i batch-inputs/iris-input.txt -o /tmp/iris-pipeline-output.txt --workers 5
```
```output
2023-01-22 18:25:18,651 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-22 18:25:18,653 [mlserver] INFO - server url: localhost:9000
2023-01-22 18:25:18,653 [mlserver] INFO - model name: iris-pipeline.pipeline
2023-01-22 18:25:18,653 [mlserver] INFO - request headers: {}
2023-01-22 18:25:18,653 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-01-22 18:25:18,653 [mlserver] INFO - output file path: /tmp/iris-pipeline-output.txt
2023-01-22 18:25:18,653 [mlserver] INFO - workers: 5
2023-01-22 18:25:18,653 [mlserver] INFO - retries: 3
2023-01-22 18:25:18,653 [mlserver] INFO - batch interval: 0.0
2023-01-22 18:25:18,653 [mlserver] INFO - batch jitter: 0.0
2023-01-22 18:25:18,653 [mlserver] INFO - connection timeout: 60
2023-01-22 18:25:18,653 [mlserver] INFO - micro-batch size: 1
2023-01-22 18:25:18,963 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-22 18:25:18,963 [mlserver] INFO - Total processed instances: 100
2023-01-22 18:25:18,963 [mlserver] INFO - Time taken: 0.31 seconds
```

#### Checking the Output

We can check the inference responses by looking at the contents of the output file:
```bash
cat /tmp/iris-pipeline-output.txt | head -n 1 | jq
```

## Running the Tensorflow Batch Job

The samples folder contains an example batch input for the tensorflow model, just as it did for the scikit-learn model. You can find it at `batch-inputs/tfsimple-input.txt`. Let's take a look at the first inference request in the file:

```bash
cat batch-inputs/tfsimple-input.txt | head -n 1 | jq
```
```json
{
  "inputs": [
    {
      "name": "INPUT0",
      "data": [
        75,
        39,
        9,
        44,
        32,
        97,
        99,
        40,
        13,
        27,
        25,
        36,
        18,
        77,
        62,
        60
      ],
      "datatype": "INT32",
      "shape": [
        1,
        16
      ]
    },
    {
      "name": "INPUT1",
      "data": [
        39,
        7,
        14,
        58,
        13,
        88,
        98,
        66,
        97,
        57,
        49,
        3,
        49,
        63,
        37,
        12
      ],
      "datatype": "INT32",
      "shape": [
        1,
        16
      ]
    }
  ]
}
```

### Tensorflow Model

As before, we can run the inference batch job using the `mlserver infer` command:
```bash
mlserver infer -u localhost:9000 -m tfsimple1 -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-output.txt --workers 10

```
```
2023-01-23 14:56:10,870 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-23 14:56:10,872 [mlserver] INFO - server url: localhost:9000
2023-01-23 14:56:10,872 [mlserver] INFO - model name: tfsimple1
2023-01-23 14:56:10,872 [mlserver] INFO - request headers: {}
2023-01-23 14:56:10,872 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-01-23 14:56:10,872 [mlserver] INFO - output file path: /tmp/tfsimple-output.txt
2023-01-23 14:56:10,872 [mlserver] INFO - workers: 10
2023-01-23 14:56:10,872 [mlserver] INFO - retries: 3
2023-01-23 14:56:10,872 [mlserver] INFO - batch interval: 0.0
2023-01-23 14:56:10,872 [mlserver] INFO - batch jitter: 0.0
2023-01-23 14:56:10,872 [mlserver] INFO - connection timeout: 60
2023-01-23 14:56:10,872 [mlserver] INFO - micro-batch size: 1
2023-01-23 14:56:11,077 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-23 14:56:11,077 [mlserver] INFO - Total processed instances: 100
2023-01-23 14:56:11,078 [mlserver] INFO - Time taken: 0.21 seconds
```

#### Checking the Output

We can check the inference responses by looking at the contents of the output file:
```bash
cat /tmp/tfsimple-output.txt | head -n 1 | jq
```

You should get the following response:
```json
{
  "model_name": "tfsimple1_1",
  "model_version": "1",
  "id": "54e6c237-8356-4c3c-96b5-2dca4596dbe9",
  "parameters": {
    "batch_index": 0,
    "inference_id": "54e6c237-8356-4c3c-96b5-2dca4596dbe9"
  },
  "outputs": [
    {
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "parameters": {},
      "data": [
        114,
        46,
        23,
        102,
        45,
        185,
        197,
        106,
        110,
        84,
        74,
        39,
        67,
        140,
        99,
        72
      ]
    },
    {
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "parameters": {},
      "data": [
        36,
        32,
        -5,
        -14,
        19,
        9,
        1,
        -26,
        -84,
        -30,
        -24,
        33,
        -31,
        14,
        25,
        48
      ]
    }
  ]
}
```

### Tensorflow Pipeline
```bash
mlserver infer -u localhost:9000 -m tfsimple1 -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-pipeline-output.txt --workers 10
```
```output
2023-01-23 14:56:10,870 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-23 14:56:10,872 [mlserver] INFO - server url: localhost:9000
2023-01-23 14:56:10,872 [mlserver] INFO - model name: tfsimple1
2023-01-23 14:56:10,872 [mlserver] INFO - request headers: {}
2023-01-23 14:56:10,872 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-01-23 14:56:10,872 [mlserver] INFO - output file path: /tmp/tfsimple-pipeline-output.txt
2023-01-23 14:56:10,872 [mlserver] INFO - workers: 10
2023-01-23 14:56:10,872 [mlserver] INFO - retries: 3
2023-01-23 14:56:10,872 [mlserver] INFO - batch interval: 0.0
2023-01-23 14:56:10,872 [mlserver] INFO - batch jitter: 0.0
2023-01-23 14:56:10,872 [mlserver] INFO - connection timeout: 60
2023-01-23 14:56:10,872 [mlserver] INFO - micro-batch size: 1
2023-01-23 14:56:11,077 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-23 14:56:11,077 [mlserver] INFO - Total processed instances: 100
2023-01-23 14:56:11,078 [mlserver] INFO - Time taken: 0.25 seconds

```

#### Checking the Output

We can check the inference responses by looking at the contents of the output file:
```bash
cat /tmp/tfsimple-pipeline-output.txt | head -n 1 | jq
```

## Cleaning Up

Now that we've run our batch examples, let's remove the models and pipelines we created:

```bash
seldon model unload iris
```

```bash
seldon model unload tfsimple1
```

```bash
seldon pipeline unload iris-pipeline
```

```bash
seldon pipeline unload tfsimple
```

And finally let's spin down our local instance of Seldon Core:

```bash
cd ../ && make undeploy-local
```

