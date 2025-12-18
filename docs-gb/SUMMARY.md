# Table of contents

## Getting Started

* [Quick Start Guide](README.md)
* [License](LICENSE.md)
* [Installation](getting-started/installation/README.md)
  * [Install in Kubernetes](install/installation.md)
  * [Install Locally](install/kind.md)
  * [Install on Google Cloud](install/gcp.md)
  * [Install on Azure](install/azure.md)
  * [Install on AWS](install/aws.md)
* [Community](developer/community.md)

## Concepts

* [Overview of Components](overview.md)

## Configuration

* [Installation Parameters](configuration/installation-parameters/README.md)
  * [Helm Configurations](install/advanced-helm-chart-configuration.md)
  * [Usage Reporting](install/usage-reporting.md)
  * [Change Log](https://github.com/SeldonIO/seldon-core/tree/master/CHANGELOG.md)
  * [Upgrading](upgrading.md)
* [Deployments](configuration/deployments/README.md)
  * [Deployment Techniques](deployments/deploying.md)
  * [Supported API Protocols](deployments/protocols.md)
  * [Testing Model Endpoints](deployments/serving.md)
  * [Replica Scaling](deployments/scaling.md)
  * [Model Metadata](deployments/metadata.md)
  * [Budgeting Disruptions](deployments/disruption-budgets.md)
  * [Graph Deployment Options](deployments/graph-modes.md)
  * [AB Tests and Progerssive Rollouts](deployments/abtests.md)
  * [Troubleshooting Deployments](deployments/troubleshooting.md)
* [Servers](configuration/servers/README.md)
  * [Custom Inference Servers](servers/custom.md)
  * [\[Storage Initializers\]](configuration/servers/storage-initializers.md)
  * [Prepackaged Model Servers](servers/overview.md)
  * [Inference Optimization](servers/optimization.md)
  * [XGBoost Server](servers/xgboost.md)
  * [Triton Inference Server](servers/triton.md)
  * [SKLearn Server](servers/sklearn.md)
  * [Tempo Server](servers/tempo.md)
  * [MLFlow Server](servers/mlflow.md)
  * [HuggingFace Server](servers/huggingface.md)
  * [TensorFlow Serving](servers/tensorflow.md)
* [Routing](configuration/routing/README.md)
  * [Ingress with Istio](routing/istio.md)
  * [Using the Istio Service Mesh](routing/istio.md#using-the-istio-service-mesh)
  * [Ambassador Ingress](routing/ambassador.md)
  * [OpenShift](routing/openshift.md)
  * [Routers Including Multi armed Bandits](routing/routers.md)
  * [Infrence Graphs](routing/inference-graph.md)
* [Wrappers and SDKs](configuration/wrappers-and-sdks/README.md)
  * [Python Language Wrapper](configuration/wrappers-and-sdks/python-language-wrapper/README.md)
    * [Install the Seldon Core Python module](wrappers/python/python_module.md)
    * [Creating your Python inference class](wrappers/python/python_component.md)
    * [Create image S2I](wrappers/python/python_wrapping_s2i.md)
    * [Create image with Dockerfile](wrappers/python/python_wrapping_docker.md)
    * [Seldon Python server configuration](wrappers/python/python_server.md)
    * [Calling the Seldon API with Seldon Python client](wrappers/python/seldon_client.md)
    * [\[Python API reference\]](configuration/wrappers-and-sdks/python-language-wrapper/python-api-reference.md)
    * [Development Tips](wrappers/python/developer_notes.md)
  * [\[Go Language Wrapper\]](configuration/wrappers-and-sdks/go-language-wrapper.md)
  * [Java Language Wrapper](wrappers/java.md)
  * [Nodejs Language Wrapper](wrappers/nodejs.md)
  * [C++ Language Wrapper](wrappers/cpp.md)
  * [R Language Wrapper](wrappers/r.md)
* [Integrations](configuration/integrations/README.md)
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

* [Notebooks](notebooks/README.md)
  * [Install Seldon Core](notebooks/seldon-core-setup.md)
  * [Install MinIO](notebooks/minio_setup.md)
  * [Deploy a Scikit-learn Model Binary](servers/sklearn.md)
  * [Deploy a Tensorflow Exported Model](servers/tensorflow.md)
  * [MLflow Pre-packaged Model Server A/B Test](notebooks/mlflow_server_ab_test_ambassador.md)
  * [MLflow Open Inference Protocol End to End Workflow](notebooks/mlflow_v2_protocol_end_to_end.md)
  * [Deploy a XGBoost Model Binary](servers/xgboost.md)
  * [Deploy Pre-packaged Model Server with Cluster's MinIO](notebooks/minio-sklearn.md)
  * [Custom Pre-packaged LightGBM Server](notebooks/custom_server.md)
  * [SKLearn Spacy NLP](notebooks/sklearn_spacy_text_classifier_example.md)
  * [SKLearn Iris Classifier](notebooks/iris.md)
  * [Sagemaker SKLearn Example](notebooks/sagemaker_sklearn.md)
  * [TFserving MNIST](notebooks/tfserving_mnist.md)
  * [Statsmodels Holt-Winter's time-series model](notebooks/statsmodels.md)
  * [Runtime Metrics & Tags](notebooks/runtime_metrics_tags.md)
  * [Triton GPT2 Example](notebooks/triton_gpt2_example.md)
  * [NVIDIA TensorRT MNIST](notebooks/tensorrt.md)
  * [OpenVINO ImageNet](notebooks/openvino.md)
  * [OpenVINO ImageNet Ensemble](notebooks/openvino_ensemble.md)
  * [Triton Examples](notebooks/triton_examples.md)
  * [Kubeflow Seldon E2E Pipeline](notebooks/kubeflow_seldon_e2e_pipeline.md)
  * [H2O Java MoJo](notebooks/cifar10_kafka.mdh2o_mojo.md)
  * [Outlier Detection with Combiner](notebooks/outlier_combiner.md)
  * [Stream Processing with KNative Eventing](notebooks/knative_eventing_streaming.md)
  * [Kafka CIFAR10](notebooks/cifar10_kafka.md)
  * [Kafka SpaCy SKlearn NLP](notebooks/kafka_spacy_sklearn.md)
  * [Kafka KEDA Autoscaling](notebooks/kafka_keda.md)
  * [CPP Wrapper Simple Single File](notebooks/cpp_simple.md)
  * [Advanced CPP Buildsystem Override](notebooks/cpp_advanced.md)
  * [Environment Variables](notebooks/cpp_advanced.md#environment-variables)
  * [AWS EKS Tensorflow Deep MNIST](notebooks/aws_eks_deep_mnist.md)
  * [Azure AKS Tensorflow Deep MNIST](notebooks/azure_aks_deep_mnist.md)
  * [GKE with GPU Tensorflow Deep MNIST](notebooks/gpu_tensorflow_deep_mnist.md)
  * [Alibaba Cloud Tensorflow Deep MNIST](notebooks/alibaba_ack_deep_mnist.md)
  * [Triton GPT2 Example Azure](notebooks/triton_gpt2_example_azure.md)
  * [Setup for Triton GPT2 Example Azure](notebooks/triton_gpt2_example_azure_setup.md)
  * [Real Time Monitoring of Statistical Metrics](notebooks/feedback_reward_custom_metrics.md)
  * [Model Explainer Example](notebooks/iris_explainer_poetry.md)
  * [Model Explainer Open Inference Protocol Example](notebooks/iris_anchor_tabular_explainer_v2.md)
  * [Outlier Detection on CIFAR10](notebooks/outlier_cifar10.md)
  * [Training Outlier Detector for CIFAR10 with Poetry](notebooks/cifar10_od_poetry.md)
  * [Batch Processing with Argo Workflows and S3 / Minio](notebooks/argo_workflows_batch.md)
  * [Batch Processing with Argo Workflows and HDFS](notebooks/argo_workflows_hdfs_batch.md)
  * [Batch Processing with Kubeflow Pipelines](notebooks/kubeflow_pipelines_batch.md)
  * [Autoscaling Example](notebooks/autoscaling_example.md)
  * [KEDA Autoscaling example](notebooks/keda.md)
  * [Request Payload Logging with ELK](notebooks/payload_logging.md)
  * [Custom Metrics with Grafana & Prometheus](notebooks/metrics.md)
  * [Distributed Tracing with Jaeger](notebooks/tracing.md)
  * [Replica control](notebooks/scale.md)
  * [Example Helm Deployments](notebooks/helm_examples.md)
  * [Max gRPC Message Size](notebooks/max_grpc_msg_size.md)
  * [Deploy Multiple Seldon Core Operators](notebooks/multiple_operators.md)
  * [Protocol Examples](notebooks/protocol_examples.md)
  * [Configurable timeouts](notebooks/timeouts.md)
  * [Custom Protobuf Data Example](notebooks/customdata_example.md)
  * [Disruption Budgets Example](notebooks/pdbs_example.md)
  * [Istio AB Test](notebooks/istio_canary.md)
  * [Ambassador AB Test](notebooks/ambassador_canary.md)
  * [Seldon/Iter8 - Progressive AB Test with Single Seldon Deployment](notebooks/iter8-single.md)
  * [Seldon/Iter8 - Progressive AB Test with Multiple Seldon Deployments](notebooks/iter8-separate.md)
  * [Chainer MNIST](notebooks/chainer_mnist.md)
  * [Custom pre-processors with the Open Inference Protocol](notebooks/transformers-v2-protocol.md)
  * [Graph Examples](notebooks/graph-examples.md)
  * [Ambassador Canary](notebooks/ambassador_canary.md)
  * [Ambassador Shadow](notebooks/ambassador_shadow.md)
  * [Ambassador Headers](notebooks/ambassador_headers.md)
  * [Istio Examples](notebooks/istio_examples.md)
  * [Istio Canary](notebooks/istio_canary.md)
  * [Patch Volumes for Version 1.2.0 Upgrade](notebooks/patch_1_2.md)
  * [Service Orchestrator Overhead](notebooks/bench_svcOrch.md)
  * [Tensorflow Benchmark](notebooks/bench_tensorflow.md)
  * [Argo Workflows Benchmarking](notebooks/vegeta_bench_argo_workflows.md)
  * [Python Serialization Cost Benchmark](notebooks/python_serialization.md)
  * [KMP\_AFFINITY Benchmarking Example](notebooks/python_kmp_affinity.md)
  * [Kafka Payload Logging](notebooks/kafka_logger.md)
  * [RClone Storage Initializer - testing new secret format](notebooks/rclone-upgrade.md)
  * [RClone Storage Initializer - upgrading your cluster (AWS S3 / MinIO)](notebooks/global-rclone-upgrade.md)

## Reference

* [Annotation Based Configuration](reference/annotations.md)
* [Benchmarking](reference/benchmarking.md)
* [General Availability](reference/ga.md)
* [Helm Charts](reference/helm_charts.md)
* [Images](reference/images.md)
* [Logging and Log Level](reference/log_level.md)
* [Private Docker Registry](reference/private_registries.md)
* [Prediction APIs](reference/prediction-apis/README.md)
  * [Open Inference Protocol](reference/prediction-apis/v2-protocol/README.md)
    * [REST](reference/prediction-apis/v2-protocol/rest/README.md)
      * ```yaml
        props:
          models: true
        type: builtin:openapi
        dependencies:
          spec:
            ref:
              kind: openapi
              spec: open-inference-protocol-v2
        ```
    * [gRPC](reference/prediction-apis/v2-protocol/grpc.md)
  * [Scalar Value Types](reference/prediction-apis/v2-protocol/grpc.md#scalar-value-types)
  * [Microservice API](reference/internal-api.md)
  * [External API](reference/external-prediction.md)
  * [Prediction Proto Buffer Spec](reference/prediction.md)
  * [Prediction Open API Spec](reference/prediction-apis/prediction-open-api-spec/README.md)
    * [Seldon Core External via Ambassador](reference/prediction-apis/prediction-open-api-spec/seldon-core-external-via-ambassador/README.md)
      * ```yaml
        props:
          models: true
        type: builtin:openapi
        dependencies:
          spec:
            ref:
              kind: openapi
              spec: engine
        ```
    * [Seldon Core Internal microservice API](reference/prediction-apis/prediction-open-api-spec/seldon-core-internal-microservice-api/README.md)
      * ```yaml
        props:
          models: true
        type: builtin:openapi
        dependencies:
          spec:
            ref:
              kind: openapi
              spec: wrapper
        ```
* [Release Highlights](reference/release-highlights/README.md)
  * [Release 1.7.0 Hightlights](reference/release-highlights/release-1.7.0.md)
  * [Release 1.6.0 Hightlights](reference/release-highlights/release-1.6.0.md)
  * [Release 1.5.0 Hightlights](reference/release-highlights/release-1.5.0.md)
  * [Release 1.1.0 Hightlights](reference/release-highlights/release-1.1.0.md)
  * [Release 1.0.0 Hightlights](reference/release-highlights/release-1.0.0.md)
  * [Release 0.4.1 Hightlights](reference/release-highlights/release-0.4.1.md)
  * [Release 0.4.0 Hightlights](reference/release-highlights/release-0.4.0.md)
  * [Release 0.3.0 Hightlights](reference/release-highlights/release-0.3.0.md)
  * [Release 0.2.7 Hightlights](reference/release-highlights/release-0.2.7.md)
  * [Release 0.2.6 Hightlights](reference/release-highlights/release-0.2.6.md)
  * [Release 0.2.5 Hightlights](reference/release-highlights/release-0.2.5.md)
  * [Release 0.2.3 Hightlights](reference/release-highlights/release-0.2.3.md)
* [Seldon Deployment CRD](reference/seldon-deployment-crd.md)
* [Service Orchestrator](reference/svcorch.md)
* [Kubeflow](reference/kubeflow.md)
* [Archived Docs](https://docs.seldon.io/projects/seldon-core/en/1.18/nav/archive.html)

## Contributing

* [Overview](developer/readme.md)
* [Seldon Core Licensing](developer/contributing.md)
* [End to End Tests](developer/e2e.md)
* [Roadmap](developer/roadmap.md)
* [Build using Private Repo](developer/buid-using-private-repo.md)
* [Seldon Docs Home](https://docs.seldon.ai/home)
