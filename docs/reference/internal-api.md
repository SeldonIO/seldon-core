# Internal Microservice API

![graph](./graph.png)

To add microservice components to a runtime prediction graph users need to create service that respects the internal API. The API provides a default service for each type of component within the system:

 - [Model](#model)
 - [Router](#router)
 - [Combiner](#combiner)
 - [Transformer](#transformer)
 - [Output_Transformer](#output_transformer)


## Model

A service to return predictions. 

### REST API

#### Predict

 | | |
 | - |- |
 | Endpoint | POST /predict |
 | Request | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |

Example request payload: 

```json
{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
```

### gRPC

```protobuf
service Model {
  rpc Predict(SeldonMessage) returns (SeldonMessage) {};
 }
``` 

See full [proto definition](./prediction.md/#proto-buffer-and-grpc-definition).

## Router

A service to route requests to one of its children and receive feedback rewards for them.

### REST API

#### Route

 | | |
 | - |- |
 | Endpoint | POST /route |
 | Request | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |

Example request payload: 

```json
{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
```

#### Send Feedback

 | | |
 | - |- |
 | Endpoint | POST /send-feedback |
 | Request | JSON representation of [```Feedback```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |

Example request payload: 

```json
    "request": {
        "data": {
            "names": ["a", "b"],
            "tensor": {
                "shape": [1, 2],
                "values": [0, 1]
            }
        }
    },
    "response": {
        "data": {
            "names": ["a", "b"],
            "tensor": {
                "shape": [1, 1],
                "values": [0.9]
            }
        }
    },
    "reward": 1.0
}
```


### gRPC

```protobuf
service Router {
  rpc Route(SeldonMessage) returns (SeldonMessage) {};
  rpc SendFeedback(Feedback) returns (SeldonMessage) {};
 }
``` 

See full [proto definition](./prediction.md/#proto-buffer-and-grpc-definition).

## Combiner

A service to combine reponses from its children into a single response. 

### REST API

#### Combine

 | | |
 | - |- |
 | Endpoint | POST /combine |
 | Request | JSON representation of [```SeldonMessageList```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |


### gRPC

```protobuf
service Combiner {
  rpc Aggregate(SeldonMessageList) returns (SeldonMessage) {};
}
``` 

See full [proto definition](./prediction.md/#proto-buffer-and-grpc-definition).



## Transformer

A service to transform its input.

### REST API

#### Transform

 | | |
 | - |- |
 | Endpoint | POST /transform-input |
 | Request | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |

Example request payload: 

```json
{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
```

### gRPC

```protobuf
service Transformer {
  rpc TransformInput(SeldonMessage) returns (SeldonMessage) {};
}
``` 

See full [proto definition](./prediction.md/#proto-buffer-and-grpc-definition).



## Output_Transformer

A service to transform the response from its child.

### REST API

#### Transform

 | | |
 | - |- |
 | Endpoint | POST /transform-output |
 | Request | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |
 | Response | JSON representation of [```SeldonMessage```](./prediction.md/#proto-buffer-and-grpc-definition) |

Example request payload: 

```json
{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
```

### gRPC

```protobuf
service OutputTransformer {
  rpc TransformOutput(SeldonMessage) returns (SeldonMessage) {};
}
``` 

See full [proto definition](./prediction.md/#proto-buffer-and-grpc-definition).

