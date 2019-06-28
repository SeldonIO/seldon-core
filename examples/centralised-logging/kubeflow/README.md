# Setup on existing kubeflow

## Installation

The request logging setup includes knative, which includes istio. If you've an existing kubeflow, you can instead use the kubeflow istio.

Using the kubeflow istio also provides the possibility to put services behind its authentication. See

https://www.kubeflow.org/docs/started/getting-started-k8s/

(Pending discussion at https://github.com/kubeflow/website/issues/840)

To setup seldon and supporting services on top of kubeflow run ./full-setup-existing-kubeflow.sh from the centralised-logging dir.

## Accessing services

The final output of the full-setup-existing-kubeflow.sh script includes URLs to access services such as kibana and grafana.

The path to seldon services can be found by inspecting the prefix section of `kubectl get vs -n default seldon-single-model-seldon-single-model-http -o yaml`

You can curl a service directly within the cluster - there is an example in the [request logging README](../request-logging/README.md).
