# Seldon Core Helm Charts

# Seldon Core Setup

The core charts for installing Seldon Core.

 * seldon-core-crd
   * SeldonDeployment Custom Resource Definition and Spartakus usage metrics
 * seldon-core
   * Main helm chart for installing Seldon Core
 * seldon-core-analytics
   * Example Prometheus and Grafana setup with demonstration Grafana dashboard for Seldon Core


See the [Installation](../docs/install.md)  documentation for details of how to install Seldon Core using the above charts.

# Seldon Core Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core

 * seldon-single-model
   * Serve a single model with optional Outlier Detector and attached Persistent Volume.
 * seldon-abtest
   * Serve an AB test between two models.
 * seldon-mab
   * Serve a multi-armed bandit between two models.
 * seldon-openvino
   * Deploy a single model with Intel OpenVINO model server.

For examples of using some of the above charts see [here](../notebooks/helm_examples.ipynb).

# Misc

 * seldon-core-loadtesting
   * Utility to load test