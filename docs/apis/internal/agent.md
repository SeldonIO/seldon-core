# Agent API

This API is for communication between the Seldon Scheduler and the Seldon Agent which runs next to each inference server and manages the loading and unloading of models onto the server as well as acting as a reverse proxy in the data plane for handling requests to the inference server.

## Proto Definition

```proto
syntax = "proto3";

package seldon.mlops.agent;

option go_package = "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent";

import "mlops/scheduler/scheduler.proto";

// [START Messages]

message ModelEventMessage {
  string serverName = 1;
  uint32 replicaIdx = 2;
  string modelName = 3;
  uint32 modelVersion = 4;
  enum Event {
      UNKNOWN_EVENT = 0;
      LOAD_FAIL_MEMORY = 1;
      LOADED = 2;
      LOAD_FAILED = 3;
      UNLOADED = 4;
      UNLOAD_FAILED = 5;
      REMOVED = 6; // unloaded and removed from local PVC
      REMOVE_FAILED = 7;
      RSYNC = 9; // Ask server for all models that need to be loaded
      }
  Event event = 5;
  string message = 6;
  uint64 availableMemoryBytes = 7;
}

message ModelEventResponse {

}

message ModelScalingTriggerMessage {
  string serverName = 1;
  uint32 replicaIdx = 2;
  string modelName = 3;
  uint32 modelVersion = 4;
  enum Trigger {
      SCALE_UP = 0;
      SCALE_DOWN = 1;
      }
  Trigger trigger = 5;
  uint32 amount = 6;  // number of replicas required
  map<string,uint32> metrics = 7;  // optional metrics to expose to the scheduler
}

message ModelScalingTriggerResponse {

}

message AgentDrainRequest {
  string serverName = 1;
  uint32 replicaIdx = 2;
}

message AgentDrainResponse {
  bool success = 1;
}

message AgentSubscribeRequest {
  string serverName = 1;
  bool shared = 2;
  uint32 replicaIdx = 3;
  ReplicaConfig replicaConfig = 4;
  repeated ModelVersion loadedModels = 5;
  uint64 availableMemoryBytes = 6;
}

message ReplicaConfig {
  string inferenceSvc = 1; // inference DNS service name
  int32 inferenceHttpPort = 2; // inference HTTP port
  int32 inferenceGrpcPort = 3; // Inference grpc port
  uint64 memoryBytes = 4; // The memory capacity of the server replica
  repeated string capabilities = 5; // The list of capabilities of the server, e.g. sklearn, pytorch, xgboost, mlflow
  uint32 overCommitPercentage = 6; // The percentage of over commit to allow, set to 0 (%) to disable over commit
}

message ModelOperationMessage {
  enum Operation {
    UNKNOWN_EVENT = 0;
    LOAD_MODEL = 1;
    UNLOAD_MODEL = 2;
  }
  Operation operation = 1;
  ModelVersion modelVersion = 2;
  bool autoscalingEnabled = 3;
}

message ModelVersion {
  scheduler.Model model = 1;
  uint32 version = 2;
}

// [END Messages]

// [START Services]

service AgentService {
  rpc AgentEvent(ModelEventMessage) returns (ModelEventResponse) {};
  rpc Subscribe(AgentSubscribeRequest) returns (stream ModelOperationMessage) {};
  rpc ModelScalingTrigger(stream ModelScalingTriggerMessage) returns (ModelScalingTriggerResponse) {};
  rpc AgentDrain(AgentDrainRequest) returns (AgentDrainResponse) {};
}

// [END Services]
```
