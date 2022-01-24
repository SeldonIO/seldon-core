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

Check status of model
```bash
make curl_status
```

You should see a response:
```
{
 "model_version_status": [
  {
   "version": "123",
   "state": "AVAILABLE",
   "status": {
    "error_code": "OK",
    "error_message": ""
   }
  }
 ]
}
```

Check model metadata
```bash
make curl_metadata
```

You should see a response like:
```
{
"model_spec":{
 "name": "half_plus_two",
 "signature_name": "",
 "version": "123"
}
,
"metadata": {"signature_def": {
 "signature_def": {
  "regress_x_to_y2": {
   "inputs": {
    "inputs": {
     "dtype": "DT_STRING",
     "tensor_shape": {
      "dim": [],
      "unknown_rank": true
     },
     "name": "tf_example:0"
    }
   },
   "outputs": {
    "outputs": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "y2:0"
    }
   },
   "method_name": "tensorflow/serving/regress"
  },
  "classify_x_to_y": {
   "inputs": {
    "inputs": {
     "dtype": "DT_STRING",
     "tensor_shape": {
      "dim": [],
      "unknown_rank": true
     },
     "name": "tf_example:0"
    }
   },
   "outputs": {
    "scores": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "y:0"
    }
   },
   "method_name": "tensorflow/serving/classify"
  },
  "regress_x2_to_y3": {
   "inputs": {
    "inputs": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "x2:0"
    }
   },
   "outputs": {
    "outputs": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "y3:0"
    }
   },
   "method_name": "tensorflow/serving/regress"
  },
  "serving_default": {
   "inputs": {
    "x": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "x:0"
    }
   },
   "outputs": {
    "x": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "y:0"
    }
   },
   "method_name": "tensorflow/serving/predict"
  },
  "regress_x_to_y": {
   "inputs": {
    "inputs": {
     "dtype": "DT_STRING",
     "tensor_shape": {
      "dim": [],
      "unknown_rank": true
     },
     "name": "tf_example:0"
    }
   },
   "outputs": {
    "outputs": {
     "dtype": "DT_FLOAT",
     "tensor_shape": {
      "dim": [
       {
        "size": "-1",
        "name": ""
       },
       {
        "size": "1",
        "name": ""
       }
      ],
      "unknown_rank": false
     },
     "name": "y:0"
    }
   },
   "method_name": "tensorflow/serving/regress"
  }
 }
}
}
}

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

Check Status of model
```bash
make grpc_status
```

You should see a response:
```
{
  "model_version_status": [
    {
      "version": "123",
      "state": "AVAILABLE",
      "status": {
        
      }
    }
  ]
}
```

Check model metadata
```bash
make grpc_metadata
```

You should see a reponse:
```
{
  "modelSpec": {
    "name": "half_plus_two",
    "version": "123"
  },
  "metadata": {
    "signature_def": {
      "@type": "type.googleapis.com/tensorflow.serving.SignatureDefMap",
      "signatureDef": {
        "classify_x_to_y": {
          "inputs": {
            "inputs": {
              "name": "tf_example:0",
              "dtype": "DT_STRING",
              "tensorShape": {
                "unknownRank": true
              }
            }
          },
          "outputs": {
            "scores": {
              "name": "y:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "methodName": "tensorflow/serving/classify"
        },
        "regress_x2_to_y3": {
          "inputs": {
            "inputs": {
              "name": "x2:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "outputs": {
            "outputs": {
              "name": "y3:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "methodName": "tensorflow/serving/regress"
        },
        "regress_x_to_y": {
          "inputs": {
            "inputs": {
              "name": "tf_example:0",
              "dtype": "DT_STRING",
              "tensorShape": {
                "unknownRank": true
              }
            }
          },
          "outputs": {
            "outputs": {
              "name": "y:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "methodName": "tensorflow/serving/regress"
        },
        "regress_x_to_y2": {
          "inputs": {
            "inputs": {
              "name": "tf_example:0",
              "dtype": "DT_STRING",
              "tensorShape": {
                "unknownRank": true
              }
            }
          },
          "outputs": {
            "outputs": {
              "name": "y2:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "methodName": "tensorflow/serving/regress"
        },
        "serving_default": {
          "inputs": {
            "x": {
              "name": "x:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "outputs": {
            "x": {
              "name": "y:0",
              "dtype": "DT_FLOAT",
              "tensorShape": {
                "dim": [
                  {
                    "size": "-1"
                  },
                  {
                    "size": "1"
                  }
                ]
              }
            }
          },
          "methodName": "tensorflow/serving/predict"
        }
      }
    }
  }
}

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

