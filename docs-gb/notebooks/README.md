# Notebooks

## Seldon Core Setup

* [Install Seldon Core](seldon-core-setup.md)
* [Install MinIO](minio_setup.md)

## Prepackaged Inference Server Examples

* [Deploy a Scikit-learn Model Binary](../servers/sklearn.md)
* [Deploy a Tensorflow Exported Model](../servers/tensorflow.md)
* [MLflow Pre-packaged Model Server A/B Test](mlflow_server_ab_test_ambassador.md)
* [MLflow Open Inference Protocol End to End Workflow](mlflow_v2_protocol_end_to_end.md)
* [Deploy a XGBoost Model Binary](../servers/xgboost.md)
* [Deploy Pre-packaged Model Server with Cluster's MinIO](minio-sklearn.md)
* [Custom Pre-packaged LightGBM Server](custom_server.md)

## Python Language Wrapper Examples

* [SKLearn Spacy NLP](sklearn_spacy_text_classifier_example.md)
* [SKLearn Iris Classifier](iris.md)
* [Sagemaker SKLearn Example](sagemaker_sklearn.md)
* [TFserving MNIST](tfserving_mnist.md)
* [Statsmodels Holt-Winter's time-series model](statsmodels.md)
* [Runtime Metrics & Tags](runtime_metrics_tags.md)
* [Triton GPT2 Example](triton_gpt2_example.md)

## Specialized Framework Examples

* [NVIDIA TensorRT MNIST](tensorrt.md)
* [OpenVINO ImageNet](openvino.md)
* [OpenVINO ImageNet Ensemble](openvino_ensemble.md)
* [Triton Examples](triton_examples.md)

## Incubating Projects Examples

* [Kubeflow Seldon E2E Pipeline](kubeflow_seldon_e2e_pipeline.md)
* [H2O Java MoJo](h2o_mojo.md)
* [Outlier Detection with Combiner](outlier_combiner.md)
* [Stream Processing with KNative Eventing](knative_eventing_streaming.md)
* [Kafka CIFAR10](cifar10_kafka.md)
* [Kafka SpaCy SKlearn NLP](kafka_spacy_sklearn.md)
* [Kafka KEDA Autoscaling](kafka_keda.md)
* [CPP Wrapper Simple Single File](cpp_simple.md)
* [Advanced CPP Buildsystem Override](cpp_advanced.md)
* [Environment Variables](cpp_advanced.md/#environment-variables)

## Cloud-Specific Examples

* [AWS EKS Tensorflow Deep MNIST](aws_eks_deep_mnist.md)
* [Azure AKS Tensorflow Deep MNIST](azure_aks_deep_mnist.md)
* [GKE with GPU Tensorflow Deep MNIST](gpu_tensorflow_deep_mnist.md)
* [Alibaba Cloud Tensorflow Deep MNIST](alibaba_ack_deep_mnist.md)
* [Triton GPT2 Example Azure](triton_gpt2_example_azure.md)
* [Setup for Triton GPT2 Example Azure](triton_gpt2_example_azure_setup.md)

## Advanced Machine Learning Monitoring

* [Real Time Monitoring of Statistical Metrics](feedback_reward_custom_metrics.md)
* [Model Explainer Example](iris_explainer_poetry.md)
* [Model Explainer Open Inference Protocol Example](iris_anchor_tabular_explainer_v2.md)
* [Outlier Detection on CIFAR10](outlier_cifar10.md)
* [Training Outlier Detector for CIFAR10 with Poetry](cifar10_od_poetry.md)

## Batch Processing with Seldon Core

* [Batch Processing with Argo Workflows and S3 / Minio](argo_workflows_batch.md)
* [Batch Processing with Argo Workflows and HDFS](argo_workflows_hdfs_batch.md)
* [Batch Processing with Kubeflow Pipelines](kubeflow_pipelines_batch.md)

## MLOps: Scaling and Monitoring and Observability

* [Autoscaling Example](autoscaling_example.md)
* [KEDA Autoscaling example](keda.md)
* [Request Payload Logging with ELK](payload_logging.md)
* [Custom Metrics with Grafana & Prometheus](metrics.md)
* [Distributed Tracing with Jaeger](tracing.md)
* [CI / CD with Jenkins Classic](jenkins_classic.md)
* [CI / CD with Jenkins X](jenkins_x.md)
* [Replica control](scale.md)

## Production Configurations and Integrations

* [Example Helm Deployments](helm_examples.md)
* [Max gRPC Message Size](max_grpc_msg_size.md)
* [Deploy Multiple Seldon Core Operators](multiple_operators.md)
* [Protocol Examples](protocol_examples.md)
* [Configurable timeouts](timeouts.md)
* [Custom Protobuf Data Example](customdata_example.md) 
* [Disruption Budgets Example](pdbs_example.md)

## AB Tests and Progressive Rollouts

* [Istio AB Test](istio_canary.md)
* [Ambassador AB Test](ambassador_canary.md)
* [Seldon/Iter8 - Progressive AB Test with Single Seldon Deployment](iter8-single.md) 
* [Seldon/Iter8 - Progressive AB Test with Multiple Seldon Deployments](iter8-separate.md) 

## Complex Graph Examples

* [Chainer MNIST](chainer_mnist.md)
* [Custom pre-processors with the Open Inference Protocol](transformers-v2-protocol.md) 
* [Graph Examples](graph-examples.md)

## Ingress

* [Ambassador Canary](ambassador_canary.md)
* [Ambassador Shadow](ambassador_shadow.md)
* [Ambassador Headers](ambassador_headers.md)
* [Istio Examples](istio_examples.md)
* [Istio Canary](istio_canary.md)

## Infrastructure

* [Patch Volumes for Version 1.2.0 Upgrade](patch_1_2.md) 

## Benchmarking and Load Tests

* [Service Orchestrator Overhead](bench_svcOrch.md)
* [Tensorflow Benchmark](bench_tensorflow.md)
* [Argo Workflows Benchmarking](vegeta_bench_argo_workflows.md)
* [Python Serialization Cost Benchmark](python_serialization.md)
* [KMP_AFFINITY Benchmarking Example](python_kmp_affinity.md)
* [Kafka Payload Logging](kafka_logger.md)

## Upgrading Examples

* [RClone Storage Initializer - testing new secret format](rclone-upgrade.md)
* [RClone Storage Initializer - upgrading your cluster (AWS S3 / MinIO)](global-rclone-upgrade.md) 