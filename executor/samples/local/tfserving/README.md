# Test Executor with Tensorflow Serving

You will need:

 * docker
 * curl
 * [grpcurl](https://github.com/fullstorydev/grpcurl)
 * a built executor

Clone the tensorflow repo:

```bash
make serving
```


## REST - single model

Run the following commands in different terminals.

Start tensorflow serving model
```bash
make run_tensorflow_serving
```

Start the executor locally.
```bash
make run_rest_executor
```

This will run the model shown below:

```JSON
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  annotations:
    seldon.io/executor: "true"
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: tensorflow/serving:latest
          name: half_plus_two
    graph:
      children: []
      endpoint:
        type: REST
        service_host: 0.0.0.0
        service_port: 8501
      name: half_plus_two
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

Send a request
```bash
make curl_rest
```

You should see a response:
```
{
    "predictions": [2.5, 3.0, 4.5]
}
```


## REST - chained model

Run the following commands in different terminals.

Start tensorflow serving model
```bash
make run_tensorflow_serving
```

Start the executor locally.
```bash
make run_rest_executor_chain
```

This will run against the SeldonDeployment with 2 Tensorflow models one after the other:

```JSON
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  annotations:
    seldon.io/executor: "true"
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: tensorflow/serving:latest
          name: half_plus_two
    graph:
      endpoint:
        type: REST
        service_host: 0.0.0.0
        service_port: 8501
      name: half_plus_two
      type: MODEL
      children: 
      - endpoint:
          type: REST
          service_host: 0.0.0.0
          service_port: 8501
        name: half_plus_two
        type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

Send a request
```bash
make curl_rest
```

You should see a response:
```
{
    "predictions": [3.25, 3.5, 4.25]
}
```

## gRPC

Run the following commands in different terminals.

Start tensorflow serving model
```bash
make run_tensorflow_serving
```

Start the executor locally.
```bash
make run_grpc_executor
```

This will run the model shown below:

```JSON
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  annotations:
    seldon.io/executor: "true"
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: tensorflow/serving:latest
          name: half_plus_two
    graph:
      children: []
      endpoint:
        type: GRPC
        service_host: 0.0.0.0
        service_port: 8500
      name: half_plus_two
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1
```

Send a request
```bash
make grpc_test
```

You should see a response:
```
{
  "outputs": {
    "x": {
      "dtype": "DT_FLOAT",
      "tensorShape": {
        "dim": [
          {
            "size": "3"
          }
        ]
      },
      "floatVal": [
        2.5,
        3,
        3.5
      ]
    }
  },
  "modelSpec": {
    "name": "half_plus_two",
    "version": "123",
    "signatureName": "serving_default"
  }
}
```


## gRPC Chained

Run the following commands in different terminals.

Start tensorflow serving model
```bash
make run_tensorflow_serving
```

Start the executor locally.
```bash
make run_grpc_executor_chain
```

This will run the model shown below:

```JSON
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  annotations:
    seldon.io/executor: "true"
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: tensorflow/serving:latest
          name: half_plus_two
    graph:
      endpoint:
        type: GRPC
        service_host: 0.0.0.0
        service_port: 8500
      name: half_plus_two
      type: MODEL
      children: 
      - endpoint:
          type: GRPC
          service_host: 0.0.0.0
          service_port: 8500
        name: half_plus_two
        type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1
```

Send a request
```bash
make grpc_test
```

You should see a response:
```
{
  "outputs": {
    "x": {
      "dtype": "DT_FLOAT",
      "tensorShape": {
        "dim": [
          {
            "size": "3"
          }
        ]
      },
      "floatVal": [
        3.25,
        3.5,
        3.75
      ]
    }
  },
  "modelSpec": {
    "name": "half_plus_two",
    "version": "123",
    "signatureName": "serving_default"
  }
}
```

