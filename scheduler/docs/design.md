# Design

![architecture](architecture.png)

 * **Operator**
    * Calls scheduler with load/unload model requests
 * **Agent**
    * Runs on each server pod. On start calls scheduler to inform of new server replica with given capabilities, memory.
    * Handles load requests:
      * Tell Rclone server to download artifacts
      * Tell server to load/unload model
 * **Scheduler**
    * Assigns models to server replicas
    * Manages gRPC connections to:
       * Agent to load/unload models assigned to a replica
       * Envoy to update routing to replicas for models
 * **Server**
    * V2 compatible ML server
 * **Rclone**
    * Rclone server to download artifacts onto local PVC
 * **DB**
    * State of truth store for model->server mapping
 * **Scaler**
    * Handles possble:
       * Scale to zero
       * Scale from zero
       * Payload logging

## Agent-Scheduler design

Requirements:

  * Handle server updates, e.g. user changes server configuration (more memory, different image with capabilities (sklearn, alibi etc))
  * Handle server failures

Due to above current design:
  * Agent calls scheduler on startup, when pod is ready, and informs scheduler of new replica (if any, there may be 0 replicas) with a server with given capabilities
  * Scheduler tells agent of model(s) to load
  * Scheduler handles loss of gRPC connection by rescheduling models (if possible)
  * Scheduler reschedules failed scheduling models when replicas restart

## Scheduler design

Requirements:

 * Handle core state of truth for model->server mapping
 * Update via a "scheduling algorithm"
 * Handle syncing of Envoy and Agents when model->server changes
 * Handle running as multiple pods with remote DB storage

## gRPC Services

 * [Scheduler](../apis/mlops/scheduler/scheduler.proto)
 * [Agent](../apis/mlops/agent/agent.proto)


## Model Replica State

A model state can be in a set of states.

```golang
const (
    ModelStateUnknown ModelState = iota
    ModelProgressing
    ModelAvailable
    ModelFailed
    ModelTerminating
    ModelTerminated
    ModelTerminateFailed
    ScheduleFailed
    ModelScaledDown
)
```

And the underlying model replicas can be in a set of states.

```golang
const (
	ModelReplicaStateUnknown ModelReplicaState = iota
	LoadRequested
	Loading
	Loaded
	LoadFailed
	UnloadEnvoyRequested
	UnloadRequested
	Unloading
	Unloaded
	UnloadFailed
	Available
	LoadedUnavailable
	Draining
)
```

The idea is the core scheduler, the scheduler-agent server and the scheduler-envoy can in theory all work independently with the later two syncing when a model state changes.

### Scheduler
#### Loading
 1. Scheduler gprc receives load model RPC
 1. Scheduler assigns model to 1 or more replicas updating the core model->server state
    1. Model replica state it set to `LoadRequested` and Model state is set to `ModelProgressing`

#### Unloading
 1. Scheduler gprc receives unload model RPC
 1. Scheduler removes model replicas updating the core model->server state
    1. Model replica state it set to `UnloadEnvoyRequested` and Model state is set to `ModelTerminating`

### Scheduler-Agent

When it syncs

#### Loading
 1. Agent-server sees model replica is `LoadRequested`
    1. A load request to agent for desired replicas and changes state to `Loading`
    1. When model is loaded agent sends Event update to scheduler and scheduler sets state to `Loaded`

#### Unloading
 1. Agent-server sees model replica is `UnloadRequested`
    1. An unload request to agent for desired replicas and changes state to `Unloading`
    1. When model is unloaded agent sends Event update to scheduler and scheduler sets state to `Unloaded`

### Scheduler-Envoy
 1. Envoy syncs and updates mapping for any models it sees that have state `Loaded` to be `Available` and removes any whose state is not `Loaded`
 1. Envoy sets all model replicas marked as `UnloadEnvoyRequested` to `UnloadRequested`, which would trigger Agent-server model replica unload


Below a diagram of the interaction of creating a new Model

```mermaid
sequenceDiagram
    participant User
    participant Operator
    participant Scheduler
    participant ModelStore
    participant EventBus
    participant AgentClient as Agent Client
    participant AgentServer
    participant k8sAPI
    participant MLServer
    participant IncrementalProcessor
    participant Envoy
    
    User->>Operator: Loads a new model First Time
    Operator->>Scheduler: Request model load
    Scheduler->>ModelStore: Create a new Model in store
    ModelStore->>Scheduler: get all servers
    Note right of Scheduler: filter servers
    Note right of Scheduler: sort servers
    Scheduler->>ModelStore: Update loaded Models (find and update to servers)
    loop update loaded models per server and replicas
    Note right of ModelStore: iterate through every assigned replica and set load requested
    Note right of ModelStore: unload other models in other replicas that weren't requested
      ModelStore->>ModelStore: update model status to Progressing
      ModelStore->>EventBus: modelUpdateEventSource
      EventBus-->>Scheduler: Handle Model events
      Scheduler->>Operator: Send Status Event
      Operator->>Operator: Update Model status to Progressing
    end
    EventBus->>AgentClient: model update event
    loop through each model replica with Load Requested in Agent Client
      AgentClient->>AgentServer: GRPC load request
      AgentClient->>ModelStore: Update Replica state for model to Load Requested
      ModelStore-->>EventBus: Model Update Event
    end
    
    AgentServer->>k8sAPI: Get rclone config
    AgentServer->>AgentServer: copy model artifact into pvc
    AgentServer->>MLServer: Load model
    
    AgentServer->>AgentClient: New Agent event Model Loaded
    AgentClient->>ModelStore: update model status to Loaded
    ModelStore-->>EventBus: new Model Update Event Available
    
    EventBus-->>Scheduler: Handle Model events
    EventBus-->>IncrementalProcessor: new Model Event with status Available
    IncrementalProcessor->>Envoy: Add a new route for Model
    Scheduler->>Operator: Send Status Event
    Operator->>Operator: Update Model status to Available
```