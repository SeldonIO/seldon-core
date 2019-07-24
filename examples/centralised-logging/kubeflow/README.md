# Setup on existing kubeflow

## Installation

The request logging setup includes knative, which includes istio. If you've an existing kubeflow, you can instead use the kubeflow istio.

For kubeflow cluster setup and installation see - we recommend installing with istio into an existing cluster:

https://www.kubeflow.org/docs/started/getting-started-k8s/

To setup seldon and supporting services on top of kubeflow, using its istio, run ./full-setup-existing-kubeflow.sh from the centralised-logging dir.

## Accessing services

The final output of the full-setup-existing-kubeflow.sh script includes URLs to access services such as kibana and grafana.

The path to seldon services can be found by inspecting the prefix section of `kubectl get vs -n default seldon-single-model-seldon-single-model-http -o yaml`

You can curl a service directly within the cluster - there is an example in the [request logging README](../request-logging/README.md).
