# External Prediction API

![API](./api.png)

The Seldon Core exposes a generic external API to connect your ML runtime prediction to external business applications.

## REST API

### Prediction

 - endpoint : POST /api/v1.0/predictions
 - payload : JSON representation of ```SeldonMessage``` - see [proto definition](./prediction.md#proto-buffer-and-grpc-definition)
 - example payload :

   ```json
   {"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
   ```

### Feedback

 - endpoint : POST /api/v1.0/feedback
 - payload : JSON representation of ```Feedback``` - see [proto definition](./prediction.md#proto-buffer-and-grpc-definition)

## gRPC

```protobuf
service Seldon {
  rpc Predict(SeldonMessage) returns (SeldonMessage) {};
  rpc SendFeedback(Feedback) returns (SeldonMessage) {};
 }
```

see full [proto definition](./prediction.md#proto-buffer-and-grpc-definition)

