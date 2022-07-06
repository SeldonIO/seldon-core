# Hodometer

Hodometer collects and publishes anonymous usage metrics for Seldon Core v2.

## Usage

The metrics collected by Hodometer are used to understand how people interact with Seldon Core.
This includes how many Seldon Core clusters there are and how features are used.

[See below](#metrics) for the metrics we collect.

Enabling full metrics collection allows Seldon to best understand what users want.

However, you can reduce the level of information provided by changing the collection level.
Which metrics are covered by which level is documented below.

Alernatively, you can opt out entirely by simply not installing Hodometer, or by removing your existing installation of it.
This does not affect your usage of the software in any way.

## Installation

_TODO_

## Metrics

There are three levels of metrics that can be enabled:
* Cluster-level -- basic information about the installation
* Resource-level -- high-level details about which Seldon Core v2 resources are used
* Feature-level -- more detailed information about how resources are used

The set of metrics available at each level is:

<!-- start list metrics -->

| Metric name | Level | Format | Notes |
| --- | --- | --- | --- |
| `cluster_id` | cluster | UUID | A random identifier for this cluster for de-duplication |
| `seldon_core_version` | cluster | Version number | E.g. 1.2.3 |
| `is_global_installation` | cluster | Boolean | Whether installation is global or namespaced |
| `is_kubernetes` | cluster | Boolean | Whether or not the installation is in Kubernetes |
| `kubernetes_version` | cluster | Version number | Kubernetes server version, if inside Kubernetes |
| `node_count` | cluster | Integer | Number of nodes in the cluster, if inside Kubernetes |
| `model_count` | resource | Integer | Number of `Model` resources |
| `pipeline_count` | resource | Integer | Number of `Pipeline` resources |
| `experiment_count` | resource | Integer | Number of `Experiment` resources |
| `server_count` | resource | Integer | Number of `Server` resources |
| `server_replica_count` | resource | Integer | Total number of `Server` resource replicas |
| `multimodel_enabled_count` | feature | Integer | Number of `Server` resources with multi-model serving enabled |
| `overcommit_enabled_count` | feature | Integer | Number of `Server` resources with overcommitting enabled |
| `gpu_enabled_count` | feature | Integer | Number of `Server` resources with GPUs attached |
| `inference_server_name` | feature | String | Name of inference server, e.g. MLServer or Triton |
| `server_cpu_cores_sum` | feature | Float | Total of CPU limits across all `Server` resource replicas, in cores |
| `server_memory_gb_sum` | feature | Float | Total of memory limits across all `Server` resource replicas, in GiB |

<!-- end list metrics -->

## Privacy

We aim to not collect any sensitive or identifying information.
For example, we do **not** collect IP addresses, machine names, or company information.

If you are concerned about any of the information being collected, please raise an issue or PR, or contact Seldon directly.

