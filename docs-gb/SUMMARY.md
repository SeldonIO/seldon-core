
# Table of contents

## Getting Started
* [Quick Start Guide](README.md)
* [License](LICENSE.md)
* Installation
    * [Install in Kubernetes](install/installation.md)
    * [Install Locally](install/kind.md)
    * [Install on Google Cloud](install/gcp.md)
    * [Install on Azure](install/azure.md)
    * [Install on AWS](install/aws.md)
* [Community](developer/community.md)

## Concepts
* [Overview of Components](overview.md)
  
## Configuration
  * Installation Parameters
    * [Helm Configurations](install/advanced-helm-chart-configuration.md)
    * [Usage Reporting](install/usage-reporting.md)
    * [Change Log](https://github.com/SeldonIO/seldon-core/tree/master/CHANGELOG.md)
    * [Upgrading](upgrading.md)
  * Deployments
    * [Deployment Techniques](deployments/deploying.md)
    * [Supported API Protocols](deployments/protocols.md)
    * [Testing Model Endpoints](deployments/serving.md)
    * [Replica Scaling](deployments/scaling.md)
    * [Model Metadata](deployments/metadata.md)
    * [Budgeting Disruptions](deployments/disruption-budgets.md)
    * [Graph Deployment Options](deployments/graph-modes.md)
    * [AB Tests and Progerssive Rollouts](deployments/abtests.md)
    * [Troubleshooting Deployments](deployments/troubleshooting.md)
  * Servers
    * [Custom Inference Servers](servers/custom.md)
    * [Storage Initializers]
    * [Inference Optimization](servers/optimization.md)
    * [XGBoost Server](servers/xgboost.md)
    * [Triton Inference Server](servers/triton.md)
    * [SKLearn Server](servers/sklearn.md)
    * [Tempo Server](servers/tempo.md)
    * [MLFlow Server](servers/mlflow.md)
    * [HuggingFace Server](servers/huggingface.md)
    * [TensorFlow Serving](servers/tensorflow.md)
  * Routing
    * [Ingress with Istio](routing/istio.md)
    * [Using the Istio Service Mesh](routing/istio.md#using-the-istio-service-mesh)
    * [Ambassador Ingress](routing/ambassador.md)
    * [OpenShift](routing/openshift.md)
    * [Routers Including Multi armed Bandits](routing/routers.md)
    * [Infrence Graphs](routing/inference-graph.md)
  * Wrappers and SDKs
    * Python Language Wrapper
      * [Install the Seldon Core Python module](wrappers/python/python_module.md)
      * [Creating your Python inference class](wrappers/python/python_component.md)
      * [Create imgae S2I](wrappers/python/python_wrapping_s2i.md)
      * [Create image with Dockerfile](wrappers/python/python_wrapping_docker.md)
      * [Seldon Python server configuration](wrappers/python/python_server.md)
      * [Calling the Seldon API with Seldon Python client](wrappers/python/seldon_client.md)
      * [Python API reference]
      * [Development Tips](wrappers/python/developer_notes.md)
    * [Go Language Wrapper]
    * [Java Language Wrapper](wrappers/java.md)
    * [Nodejs Language Wrapper](wrappers/nodejs.md)
    * [C++ Language Wrapper](wrappers/cpp.md)
    * [R Language Wrapper](wrappers/r.md)

  * Integrations
    * [CI/CD MLOps at Scale](integrations/cicd-mlops.md)
    * [Payload Logging](integrations/logging.md)
    * [Distributed Tracing with Jaeger](integrations/distributed-tracing.md)
    * [Batch Processing](integrations/batch.md)
    * [Stream Processing with KNative](integrations/knative_eventing.md)
    * [Metrics with Prometheus](integrations/analytics.md)
    * [Native Kafka Integration](integrations/kafka.md)
    * [Model Explanations](integrations/explainers.md)
    * [Outliner Detection](integrations/outlier_detection.md)
    * [Drift Detection](integrations/drift_detection.md)

## Tutorials
 * [Notebooks](install/notebooks.md)
 
## Reference

## Contributing
 * [Overview](developer/readme.md)
 * [Seldon Core Licensing](developer/contributing.md)
 * [End to End Tests](developer/e2e.md)
 * [Roadmap](developer/roadmap.md)
 * [Build using Private repo](developer/buid-using-private-repo.md)


