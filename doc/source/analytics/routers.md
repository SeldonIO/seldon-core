# Routers in Seldon Core

## Definitions
A router is one of the pre-defined types of [predictive
units](../reference/apis/prediction.md#proto-buffer-and-grpc-definition) in
Seldon Core.
It is a microservice to route requests to one of its children and optionally
receive feedback rewards for making the routing choices.
The REST and gRPC internal APIs that the router components must conform to are
covered in the [internal API](../reference/apis/internal-api.md#route)
reference.

## Implementations
Currently we provide two reference implementations of routers in Python. Both are instances of [multi-armed bandits](https://en.wikipedia.org/wiki/Multi-armed_bandit#Semi-uniform_strategies):
* [Epsilon-greedy router](https://github.com/SeldonIO/seldon-core/tree/master/components/routers/epsilon-greedy)
* [Thompson Sampling](https://github.com/SeldonIO/seldon-core/tree/master/components/routers/thompson-sampling)

## Implementing custom routers
A router component must implement a `Route` method which will return one of the children that the router component is connected to for routing an incoming request. The options for the return value for a custom router at present are

 * -1 : Route to all children
 * -2 : Route to no children and return the current request as the response
 * N >= 0 : Route to child N

The response for REST calls should be returned as a SeldonMessage with the payload containing the route value or a JSON array containing a single integer.

Optionally a `SendFeedback` method can be implemented to provide a mechanism for informing the router on the quality of its decisions. This would be used in adaptive routers such as multi-armed bandits, refer to the [epsilon-greedy](https://github.com/SeldonIO/seldon-core/tree/master/components/routers/epsilon-greedy) example for more detail.

As an example, consider writing a custom A/B/C... testing component with a
user-specified number of children and routing probabilities (two-model routing
is already supported in Seldon Core:
[RANDOM_ABTEST](../reference/apis/prediction.md#proto-buffer-and-grpc-definition)).
In this scenario because the routing logic is static there is no need to
implement `SendFeedback` as we will not be dynamically changing the routing by
providing feedback for its routing choices.
On the other hand, an adaptive router whose routing is required to change
dynamically by providing feedback will need to implement the `SendFeedback`
method.

Because routers are generic components that only need to implement the `Route` method, there is considerable flexibility in designing the routing logic. Some example concepts going beyond random testing and multi-armed bandits:
* Routing depending on external conditions, e.g. use the time of day to route traffic to a model that has been known to perform best during a particular time period.
* Model as a router: use a predictive model within a router component to first determine a higher level class membership (e.g. cat vs dog) and according to the decision route traffic to more specific models (e.g. dog-specific model to infer a breed).

## Limitations

The current default orchestrator in Go the "executor" does not return routing meta data in request calls. This is a [known issue](https://github.com/SeldonIO/seldon-core/issues/1823). 
