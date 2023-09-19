# Scaling

## Models

Models can be scaled by setting their replica count, e.g.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.2.3/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki
  replicas: 3
```

Currently, the number of replicas will need not to exceed the replicas of the Server the model is scheduled to.

## Servers

Servers can be scaled by setting their replica count, e.g.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver
  namespace: seldon
spec:
  replicas: 4
  serverConfig: mlserver
```

Currently, models scheduled to a server can only scale up to the server replica count.

## Internal Components

Seldon Core v2 runs with several control and dataplane components. The scaling of these resources is discussed below:

- Pipeline gateway.
    - This pipeline gateway handles REST and gRPC synchronous requests to Pipelines. It is stateless and can be scaled based on traffic demand.
- Model gateway.
    - This component pulls model requests from Kafka and sends them to inference servers. It can be scaled up to the partition factor of your Kafka topics. At present we set a uniform partition factor for all topics in one installation of Seldon Core V2.
- Dataflow engine.
    - The dataflow engine runs KStream topologies to manage Pipelines. It can run as multiple replicas and the scheduler will balance Pipelines to run across it with a consistent hashing load balancer. Each Pipeline is managed up to the partition factor of Kafka (presently hardwired to one).
- Scheduler.
    - This manages the control plane operations. It is presently required to be one replica as it maintains internal state within a BadgerDB held on local persistent storage (stateful set in Kubernetes). Performance tests have shown this not to be a bottleneck at present.
- Kubernetes Controller.
    - The Kubernetes controller manages resources updates on the cluster which it passes on to the Scheduler. It is by default one replica but has the ability to scale.
- Envoy
    - Envoy replicas get their state from the scheduler for routing information and can be scaled as needed.


### Future Enhancements

 * Allow configuration of partition factor for data plane consistent hashing load balancer.
 * Allow Model gateway and Pipeline gateway to use consistent hashing load balancer.
 * Consider control plane scaling options.
