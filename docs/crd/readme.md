# Custom Resource Definitions

## Seldon Deployment

The runtime inference graph for a machine learning deployment is described as a SeldonDeployment Kubernetes resource. The structure of this manifest is defined as a [proto buffer](../reference/seldon-deployment.md). This doc will describe the SeldonDeployment resource in general and how to create one for your runtime inference graph.

## Creating your resource definition

The full specification can be found [here](../reference/seldon-deployment.md). Below we highlight various parts and describe their intent.

The core goal is to describe your runtime inference graph(s) and deploy it with appropriate resources and scale. Example illustrative graphs are shown below:

![graph](../reference/graph.png)

The top level SeldonDeployment has standard Kubernetes meta data and consists of a spec which is defined by the user and a status which will be set by the system to represent the current state of the SeldonDeployment.

```proto
message SeldonDeployment {
  required string apiVersion = 1;
  required string kind = 2;
  optional k8s.io.apimachinery.pkg.apis.meta.v1.ObjectMeta metadata = 3;
  required DeploymentSpec spec = 4;
  optional DeploymentStatus status = 5;
}
```

The core deployment spec consists of a set of ```predictors```. Each predictor represents a seperate runtime serving graph. The set of predictors will serve request as controlled by a load balancer. At present the share of traffic will be in relation to the number of replicas each predictor has. A use case for two predictors would be a main deployment and a canary, with the main deployment having 9 replicas and the canary 1, so the canary receives 10% of the overall traffic. Each predictor will be a seperately set of managed deployments with Kubernetes so it is safe to add and remove predictors without affecting existing predictors.

To allow an OAuth API to be provisioned you should specify an OAuth key and secret. If you are using Ambassador you will not need this as you can plug in your own external authentication using Ambassador.

```proto

message DeploymentSpec {
  optional string name = 1; // A unique name within the namespace.
  repeated PredictorSpec predictors = 2; // A list of 1 or more predictors describing runtime machine learning deployment graphs.
  optional string oauth_key = 6; // The oauth key for external users to use this deployment via an API.
  optional string oauth_secret = 7; // The oauth secret for external users to use this deployment via an API.
  map<string,string> annotations = 8; // Arbitrary annotations.
}

```

For each predictor you should at a minimum specify:

 * A unique name
 * A PredictiveUnit graph that presents the tree of components to deploy.
 * One or more componentSpecs which describes the set of images for parts of your container graph that will be instigated as microservice containers. These containers will have been wrapped to work within the [internal API](../reference/internal-api.md). This component spec is a standard [PodTemplateSpec](https://kubernetes.io/docs/api-reference/extensions/v1beta1/definitions/#_v1_podtemplatespec). For complex grahs you can decide to use several componentSpecs so as to separate your components into separate Pods each with their own resource requirements.
     * If you leave the ports empty for each container they will be added automatically and matched to the ports in the graph specification. If you decide to specify the ports manually they should match the port specified for the matching component in the graph specification.
 * the number of replicas of this predictor to deploy

```proto
message PredictorSpec {
  required string name = 1; // A unique name not used by any other predictor in the deployment.
  required PredictiveUnit graph = 2; // A graph describing how the predictive units are connected together.
  repeated k8s.io.api.core.v1.PodTemplateSpec componentSpecs = 3; // A description of the set of containers used by the graph. One for each microservice defined in the graph. Can be split over 1 or more PodTemplateSpecs.
  optional int32 replicas = 4; // The number of replicas of the predictor to create.
  map<string,string> annotations = 5; // Arbitrary annotations.
  optional k8s.io.api.core.v1.ResourceRequirements engineResources = 6; // Optional set of resources for the Seldon engine which is added to each Predictor graph to manage the request/response flow
  map<string,string> labels = 7; // labels to be attached to entry deplyment for this predictor
}

```

The predictive unit graph is a tree. Each node is of a particular type. If the implementation is not specified then a microservice is assumed and you must define a matching named container within the componentSpec above. Each type of PredictiveUnit has a standard set of methods it is expected to manage, see [here](../reference/seldon-deployment.md). 

For each node in the graph:

 * A unique name. If the node describes a microservice then it must match a named container with the componentSpec.
 * The children nodes.
 * The type of the predictive unit : MODEL, ROUTER, COMBINER, TRANSFORMER or OUTPUT_TRANSFORMER.
 * The implementation. This can be left blank if it will be a microserice as this is the default otherwise choose from the available appropriate implementations provided internally.
 * Methods. This can be left blank if you wish to follow the standard methods for your PredictiveNode type : see [here](../reference/seldon-deployment.md). 
 * Endpoint. In here you should minimally if this a microservice specify whether the PredictiveUnit will use REST or gRPC. Ports will be defined automatically if not specified.
 * Parameters. Specify any parameters you wish to pass to the PredictiveUnit. These will be passed in an environment variable called PREDICTIVE_UNIT_PARAMETERS as a JSON list.

```proto

message PredictiveUnit {


  required string name = 1; //must match container name of component if no implementation
  repeated PredictiveUnit children = 2; // The child predictive units.
  optional PredictiveUnitType type = 3;
  optional PredictiveUnitImplementation implementation = 4;
  repeated PredictiveUnitMethod methods = 5;
  optional Endpoint endpoint = 6; // The exposed endpoint for this unit.
  repeated Parameter parameters = 7; // Customer parameter to pass to the unit.
}


```
