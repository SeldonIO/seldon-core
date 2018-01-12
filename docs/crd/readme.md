# Custom Resource Definitions

## Seldon Deployment

The runtime inference graph for a machine learning deployment is described as a SeldonDeployment Kubernetes resourse. The structure of this manifest is defined as a [proto buffer](../reference/seldon-deployment.md). This doc will describe the SeldonDeployment resource in general and how to create one for your runtime inference graph.

## Creating your resource definition

The full specification can be found [here](../reference/seldon-deployment.md). Below we highlight various parts and describe their intent.


The core deployment spec consists of a set of ```predictors```. Each predictor represents a seperate runtime serving graph. To allow an OAuth API to be provision you should specify an OAuth key and secret.

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

 * a unique name
 * a PredictiveUnit graph that presents the tree of components to deploy.
 * A componentSpec which describes the set of images for parts of your container graph that will be instigated as microservice containers. These containers will have been wrapped to work within the [internal API](../reference/internal-api.md). This component spec is a standard [PodTemplateSpec](https://kubernetes.io/docs/api-reference/extensions/v1beta1/definitions/#_v1_podtemplatespec).
     * If you leave the ports empty for each container they will be added automatically and matched to the ports in the graph specification. If you decide to specify the ports manually they should match the port specified for the matching component in the grph specification.
 * the number of replicas of this predictor to deploy

```proto

message PredictorSpec {
  required string name = 1; // A unique name not used by any other predictor in the deployment.
  required PredictiveUnit graph = 2; // A graph describing how the predictive units are connected together.
  required k8s.io.api.core.v1.PodTemplateSpec componentSpec = 3; // A description of the set of containers used by the graph. One for each microservice defined in the graph.
  optional int32 replicas = 4; // The number of replicas of the predictor to create.
  map<string,string> annotations = 5; // Arbitrary annotations.
}

```

The predictive unit graph is a tree. Each node is of a particular type with the leaf nodes being models. If not implementation is specified then a microservice is assumed and you must define a matching named container within the componentSpec above. Each type of PredictiveUnit has a standard set of methods it is expected to manage, see [here](../reference/seldon-deployment.md). T

For each node in the graph:

 * A unique name. If the node describes a mircoservice then it must match a named container with the componentSpec
 * The children nodes.
 * The type of the predictive unit : MODEL, ROUTER, COMBINER etc
 * The implementation. This can be left blank if it will be a microserice as this is the default otherwise choose from the avilable apropriate implementions.


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


message Endpoint {

  enum EndpointType {
    REST = 0; // REST endpoints with JSON payloads
    GRPC = 1; // gRPC endpoints
  }

  optional string service_host = 1; // Hostname for endpoint.
  optional int32 service_port = 2; // The port to connect to the service.
  optional EndpointType type = 3; // The protocol handled by the endpoint.
}

message Parameter {

  enum ParameterType {
    INT = 0;
    FLOAT = 1;
    DOUBLE = 2;
    STRING = 3;
    BOOL = 4;
  }  

  required string name = 1;
  required string value = 2;
  required ParameterType type = 3;

}


```
