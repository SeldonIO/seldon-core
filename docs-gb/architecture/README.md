# Architecture

Seldon Core 2 uses a microservice architecture where each service has limited and well-defined responsibilities working together to orchestrate scalable and fault-tolerant ML serving and management. These components communicate internally using gRPC and they can be scaled independently. Seldon Core 2 services can be split into two categories: 

* **Control Plane** services are responsible for managing the operations and configurations of your ML models and workflows. This includes functionality to instantiate new inference servers, load models, update new versions of models, configure model experiments and pipelines, and expose endpoints that may receive inference requests. The main control plane component is the **Scheduler** that is responsible for managing the loading and unloading of resources (models, pipelines, experiments) onto the respective components.

* **Data Plane** services are responsible for managing the flow of data between components or models. Core 2 supports REST and gRPC payloads that follow the Open Inference Protocol (OIP). The main data plane service is **Envoy**, which acts as a single ingress for all data plane load and routes data to the relevant servers internally (e.g. Seldon MLServer or NVidia Triton pods). 

{% hint style="info" %}
**Note**: Because Core 2 architecture separates control plane and data plane responsibilities, when control plane services are down (e.g. the Scheduler), data plane inference can still be served. In this manner the system is more resilient to failures. For example, an outage of control plane services does not impact the ability of the system to respond to end user traffic. Core 2 can be provisioned to be **highly available** on the data plane path. 
{% endhint %}


The current set of services used in Seldon Core 2 is shown below. Following the diagram, we will describe each control plane and data plan service.

![architecture](../images/architecture.png)

## Control Plane

### Scheduler
This service manages the loading and unloading of Models, Pipelines and Experiments on the relevant micro services. It is also responsible for matching Models with available Servers in a way that optimises infrastructure use. In the current design we can only have _one_ instance of the Scheduler as its internal state is persisted on disk.

When the Scheduler (re)starts there is a synchronisation flow to coordinate the startup process and to attempt to wait for expected Model Servers to connect before proceeding with control plane operations. This is important so that ongoing data plane operations are not interrupted. This then introduces a delay on any control plane operations until the process has finished (including control plan resources status updates). This synchronisation process has a timeout, which has a default of 10 minutes. It can be changed by setting helm seldon-core-v2-components value `scheduler.schedulerReadyTimeoutSeconds`.

### Agent 
This service manages the loading and unloading of models on a server and access to the server over REST/gRPC. It acts as a reverse proxy to connect end users with the actual Model Servers. In this way the system collects stats and metrics about data plane inferences that helps with observability and scaling. 

### Controller
We also provide a Kubernetes Operator to allow Kubernetes usage. This is implemented in the Controller Manager microservice, which manages CRD reconciliation with Scheduler. Currently Core 2 supports _one_ instance of the Controller.

{% hint style="info" %}
**Note**: All services besides the Controller are Kubernetes agnostic and can run locally, e.g. on Docker Compose.
{% endhint %}

## Data Plane

### Pipeline Gateway 
This service handles REST/gRPC calls to Pipelines. It translates between synchronous requests to Kafka operations, producing a message on the relevant input topic for a Pipeline and consuming from the output topic to return inference results back to the users.

### Model Gateway 
This service handles the flow of data from models to inference requests on servers and passes on the responses via Kafka.

### Dataflow Engine 
This service handles the flow of data between components in a pipeline, using Kafka Streams. It enables Core 2 to chain and join Models together to provide complex Pipelines.

### Envoy 
This service manages the proxying of requests to the correct servers including load balancing.

## Dataflow Architecture and Pipelines

To support the movement towards data centric machine learning Seldon Core 2 follows a dataflow paradigm. By taking a decentralized route that focuses on the flow of data, users can have more flexibility and insight as they build and manage complex AI applications in production. This contrasts with more centralized orchestration approaches where data is secondary.

![dataflow](../images/dataflow.png)

### Kafka
Kafka is used as the backbone for Pipelines allowing decentralized, synchronous and asynchronous usage. This enables Models to be connected together into arbitrary directed acyclic graphs. Models can be reused in different Pipelines. The flow of data between models is handled by the dataflow engine using [Kafka Streams](https://docs.confluent.io/platform/current/streams/concepts.html).

![kafka](../images/kafka.png)

By focusing on the data we allow users to join various flows together using stream joining concepts as shown below.

![joins](../images/joins.png)

We support several types of joins:
* _inner joins_, where all inputs need to be present for a transaction to join the tensors passed through the Pipeline;
* _outer joins_, where only a subset needs to be available during the join window
* _triggers_, in which data flows need to wait until records on one or more trigger data flows appear. The data in these triggers is not passed onwards from the join.

These techniques allow users to create complex pipeline flows of data between machine learning components.

More discussion on the data flow view of machine learning and its effect on v2 design can be found [here](dataflow.md).
