# Table of contents

## About
* [Overview](README.md)
* [Concepts](/docs-gb/concepts/README.md)
* [Architecture](architecture/README.md)

## Installation
* [Installation Overview](installation/README.md)
* [Learning Environment](installation/learning-environment/README.md)
  * [Self-hosted Kafka](installation/learning-environment/self-hosted-kafka.md)
* [Production Environment](installation/production-environment/README.md)
  * [Kafka Integration](installation/production-environment/kafka/README.md)
  * [Managed Kafka](installation/production-environment/kafka/managed-kafka.md) 
  * [Ingress Controller](installation/production-environment/ingress-controller/istio.md)
* [Test the Installation](installation/test-installation.md)
    
* Advanced Configurations
  * [Server Config](kubernetes/resources/serverconfig.md)
  * [Server Runtime](kubernetes/resources/seldonruntime.md)
  * [Seldon Config](kubernetes/resources/seldonconfig.md)
  * [Pipeline Config](kubernetes/resources/pipeline.md)  
* [Upgrading](upgrading.md) 
## User Guide
<!-->
* Getting Started
  * Deploy Model (OIP+MLServer Link)
  * Pipeline
  * Inference Server
  * Run Inference -->
  * [Kubernetes Resources](kubernetes/resources/README.md) 
* [Servers](servers.md)
    * [Resource allocation](resource-allocation/README.md)
      * [Example: Serving models on dedicated GPU nodes](resource-allocation/example-serving-models-on-dedicated-gpu-nodes.md)
* [Models](models/README.md)
  <!-->  
   * CRD
    * Registration
    * Versioning
    * LLM
    * Parameterized Models
    * Links to Secret Management -->
    * [Multi-Model Serving](models/mms.md)
    * [Inference Artifacts](models/inference-artifacts.md)
    * [rClone](models/rclone.md)
    * [Parameterized Models](models/parameterized-models/README.md)
    * [Pandas Query](models/parameterized-models/pandasquery.md) 
    * [Storage Secrets](kubernetes/storage-secrets.md)
* Inference
    * [Inference Server](https://docs.seldon.io/projects/seldon-core/en/v2/contents/about/index.html#inference-servers)
    * [Run Inference](https://docs.seldon.io/projects/seldon-core/en/v2/contents/inference/index.html)
    * [OIP](apis/inference/v2.md)
    * [Batch](examples/batch-examples-k8s.md) and (examples/batch-examples-local.md)
* [Pipelines](pipelines.md)
* [Scaling](kubernetes/scaling.md)
  <!-- 
  * Server Scaling
  * Component Scaling -->
  * [Autoscaling](kubernetes/autoscaling.md)
  <!--
  * Multi-Component Serving and Overcommit -->
  * [HPA Autoscaling in single-model serving](kubernetes/hpa-rps-autoscaling.md)
* Data Science Monitoring
    * [Dataflow with Kafka](architecture/dataflow.md)
    * Request & Response Logging
    * [Model Performance Metrics](performance-tests.md)
    * [Drift Detection](drift.md)
    * [Outlier Detection](outlier.md)
    * [Explainability](explainers.md)
* Operational Monitoring
    * [Operational Metrics](metrics/operational.md)
    * Kubernetes Metrics
    * [Usage Metrics](metrics/usage.md)
    * [Tracing](kubernetes/tracing.md)
    * [Local Metrics](metrics/local-metrics-test.md)
    * [Performance Tests](performance-tests.md)
    <!--
    * Performance Tuning -->
<!--    
* Rollouts & Experiments
    * Rollout Strategies
        * Progressive Rollouts
        * Rollbacks -->
    * [Experiments](kubernetes/resources/experiment.md)
  <!-->
      * A/B Testing
      * Traffic Splitting
      * Canary
      * Shadow 
    * CI/CD
      * Link to Component Versioning -->
* [Examples](examples/README.md)
  * [Local examples](examples/local-examples.md)
  * [Kubernetes examples](examples/k8s-examples.md)
  * [Huggingface models](examples/huggingface.md)
  * [Model zoo](examples/model-zoo.md)
  * [Artifact versions](examples/multi-version.md)
  * [Pipeline examples](examples/pipeline-examples.md)
  * [Pipeline to pipeline examples](examples/pipeline-to-pipeline.md)
  * [Explainer examples](examples/explainer-examples.md)
  * [Custom Servers](examples/custom-servers.md)
  * [Local experiments](examples/local-experiments.md)
  * [Experiment version examples](examples/experiment-versions.md)
  * [Inference examples](examples/inference.md)
  * [Tritonclient examples](examples/tritonclient-examples.md)
  * [Batch Inference examples (kubernetes)](examples/batch-examples-k8s.md)
  * [Batch Inference examples (local)](examples/batch-examples-local.md)
  * [Checking Pipeline readiness](examples/pipeline-ready-and-metadata.md)
  * [Multi-Namespace Kubernetes](examples/k8s-clusterwide.md)
  * [Huggingface speech to sentiment with explanations pipeline](examples/speech-to-sentiment.md)
  * [Production image classifier with drift and outlier monitoring](examples/cifar10.md)
  * [Production income classifier with drift, outlier and explanations](examples/income.md)
  * [Conditional pipeline with pandas query model](examples/pandasquery.md)
  * [Kubernetes Server with PVC](examples/k8s-pvc.md)  


## Integrations
  * [Service Meshes](kubernetes/service-meshes/README.md)
    * [Ambassador](kubernetes/service-meshes/ambassador.md)
    * [Istio](kubernetes/service-meshes/istio.md)
    * [Traefik](kubernetes/service-meshes/traefik.md)
    * [Secure Model Endpoints](models/securing-endpoints.md)
<!--   
  * Audit Trails
  * Alerts
  * Data Management
  * Modules 
  -->
## Resources
<!--
* Troubleshooting
* Tutorials -->
* [Security](/getting-started/kubernetes-installation/security/index.html)
<!--
  * Authentication
  * Authorization
  * Secrets Management 
  -->
* [APIs](apis/README.md)
  * [Internal](apis/internal/README.md)
    * [Chainer](apis/internal/chainer.md)
    * [Agent](apis/internal/agent.md)
  * [Inference](apis/inference/README.md)
    * [Open Inference Protocol](apis/inference/v2.md)
  * [Scheduler](apis/scheduler.md)
  * [Seldon CLI](getting-started/cli.md)
    * [CLI](cli/README.md)
    * [Seldon](cli/seldon.md)
      * [Config](cli/seldon\_config.md)
      * [Config Activate](cli/seldon\_config\_activate.md)
      * [Config Deactivate](cli/seldon\_config\_deactivate.md)
      * [Config Add](cli/seldon\_config\_add.md)
      * [Config List](cli/seldon\_config\_list.md)
      * [Config Remove](cli/seldon\_config\_remove.md)
    * [Experiment](cli/seldon\_experiment.md)
      * [Experiment Start](cli/seldon\_experiment\_start.md)
      * [Experiment Status](cli/seldon\_experiment\_status.md)
      * [Experiment List](cli/seldon\_experiment\_list.md)
      * [Experiment Stop](cli/seldon\_experiment\_stop.md)
    * [Model](cli/seldon\_model.md)
      * [Model Status](cli/seldon\_model\_status.md)
      * [Model Load](cli/seldon\_model\_load.md)
      * [Model List](cli/seldon\_model\_list.md)
      * [Model Infer](cli/seldon\_model\_infer.md)
      * [Model Metadata](cli/seldon\_model\_metadata.md)
      * [Model Unload](cli/seldon\_model\_unload.md)
    * [Pipeline](cli/seldon\_pipeline.md)
      * [Pipeline Load](cli/seldon\_pipeline\_load.md)
      * [Pipeline Status](cli/seldon\_pipeline\_status.md)
      * [Pipeline List](cli/seldon\_pipeline\_list.md)
      * [Pipeline Inspect](cli/seldon\_pipeline\_inspect.md)
      * [Pipeline Infer](cli/seldon\_pipeline\_infer.md)
      * [Pipeline Unload](cli/seldon\_pipeline\_unload.md)
    * [Server](cli/seldon\_server.md)
      * [Server List](cli/seldon\_server\_list.md)
      * [Server Status](cli/seldon\_server\_status.md)
<!--    
* Reference
    * Glossary -->
* [FAQs](faqs.md)          
 
* [Development](development/README.md)
  * [License](development/licenses.md)
  * [Release](development/release.md)


