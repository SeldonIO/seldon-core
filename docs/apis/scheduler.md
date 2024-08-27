# Scheduler

The Seldon scheduler API provides a gRPC service to allow Models, Servers, Experiments, and Pipelines to be managed. In Kubernetes the manager deployed by Seldon translates Kubernetes custom resource definitions into calls to the Seldon Scheduler.

In non-Kubernetes environments users of Seldon could create a client to directly control Seldon resources using this API.

## Proto Definition

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/apis/mlops/scheduler/scheduler.proto" %}
