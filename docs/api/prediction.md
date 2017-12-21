---
title: "Prediction API"
date: 2017-12-09T17:49:41Z
weight: 1
---

## Prediction API


 * [Design]({{< ref "#design" >}})
 * [Definiton]({{< ref "#definition" >}})

### Design

![graph](./prediction.png)


### Definition

```js
syntax = "proto3";

import "google/protobuf/struct.proto";

package seldon.protos;

option java_package = "io.seldon.protos";
option java_outer_classname = "PredictionProtos";

message SeldonMessage {

  Status status = 1;
  Meta meta = 2;
  oneof data_oneof {
    DefaultData data = 3;
    bytes binData = 4;
    string strData = 5;
  }
}

message DefaultData {
  repeated string names = 1;
  oneof data_oneof {
    Tensor tensor = 2;
    google.protobuf.ListValue ndarray = 3;
  }
}

message Tensor {
  repeated int32 shape = 1 [packed=true];
  repeated double values = 2 [packed=true];
}

message Meta {
  string puid = 1; 
  map<string,google.protobuf.Value> tags = 2;
  map<string,int32> routing = 3;
  OutlierStatus outlierStatus = 4;  
}

message Status {

    enum StatusFlag {
        SUCCESS = 0;
        FAILURE = 1;
    }

    int32 code = 1;
    string info = 2;
    string reason = 3;
    StatusFlag status = 4;
}

message Feedback {
  SeldonMessage request = 1;
  SeldonMessage response = 2;
  float reward = 3;
  SeldonMessage truth = 4;
}

message RequestResponse {
  SeldonMessage request = 1;
  SeldonMessage response = 2;
}


message OutlierStatus{
    bool isOutlier = 1;
    double score = 2;
}

service Model {
  rpc Predict(SeldonMessage) returns (SeldonMessage) {};
  rpc SendFeedback(Feedback) returns (SeldonMessage) {};
 }

service Router {
  rpc Route(SeldonMessage) returns (SeldonMessage) {};
  rpc SendFeedback(Feedback) returns (SeldonMessage) {};
 }

service Transformer {
  rpc TransformInput(SeldonMessage) returns (SeldonMessage) {};
  rpc TransformOutput(SeldonMessage) returns (SeldonMessage) {};
 }


```

