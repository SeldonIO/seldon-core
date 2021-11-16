=========
Notebooks
=========

Seldon Core Setup
-----------------

.. toctree::
   :titlesonly:

   Install Seldon Core <seldon_core_setup>
   Install MinIO <minio_setup>



Prepackaged Inference Server Examples
-------------------------------------

.. toctree::
   :titlesonly:

   Deploy a Scikit-learn Model Binary <../servers/sklearn.md>
   Deploy a Tensorflow Exported Model <../servers/tensorflow.md>
   MLflow Pre-packaged Model Server A/B Test <mlflow_server_ab_test_ambassador>
   MLflow v2 Protocol End to End Workflow (Incubating) <mlflow_v2_protocol_end_to_end>
   Deploy a XGBoost Model Binary <../servers/xgboost.md>
   Deploy Pre-packaged Model Server with Cluster's MinIO <minio-sklearn>
   Custom Pre-packaged LightGBM Server <custom_server>

Python Language Wrapper Examples
--------------------------------

.. toctree::
   :titlesonly:

   SKLearn Spacy NLP <sklearn_spacy_text_classifier_example>
   SKLearn Iris Classifier <iris>
   Sagemaker SKLearn Example <sagemaker_sklearn>   
   TFserving MNIST <tfserving_mnist>
   Statsmodels Holt-Winter's time-series model <statsmodels>
   Runtime Metrics & Tags <runtime_metrics_tags>
   Triton GPT2 Example <triton_gpt2_example>

Specialised Framework Examples
------------------------------

.. toctree::
   :titlesonly:

   NVIDIA TensorRT MNIST <tensorrt>
   OpenVINO ImageNet <openvino>
   OpenVINO ImageNet Ensemble <openvino_ensemble>
   Triton Examples <triton_examples>

Incubating Projects Examples
----------------------------

.. toctree::
   :titlesonly:

   Kubeflow Seldon E2E Pipeline <kubeflow_seldon_e2e_pipeline>
   H2O Java MoJo <h2o_mojo>
   Outlier Detection with Combiner <outlier_combiner>
   Stream Processing with KNative Eventing <knative_eventing_streaming>
   Kafka CIFAR10 <cifar10_kafka>
   Kafka SpaCy SKlearn NLP <kafka_spacy_sklearn>
   CPP Wrapper Simple Single File <cpp_simple>
   CPP Wrapper Advanced Custom Build System <cpp_advanced>


Cloud-Specific Examples
-----------------------

.. toctree::
   :titlesonly:

   AWS EKS Tensorflow Deep MNIST <aws_eks_deep_mnist>
   Azure AKS Tensorflow Deep MNIST <azure_aks_deep_mnist>
   GKE with GPU Tensorflow Deep MNIST <gpu_tensorflow_deep_mnist>
   Alibaba Cloud Tensorflow Deep MNIST <alibaba_ack_deep_mnist>
   Triton GPT2 Example Azure <triton_gpt2_example_azure>
   Setup for Triton GPT2 Example Azure <triton_gpt2_example_azure_setup>

Advanced Machine Learning Insights
----------------------------------

.. toctree::
   :titlesonly:

   Real time monitoring of statistical metrics <feedback_reward_custom_metrics>
   Model Explainers <explainer_examples>
   Outlier Detection on CIFAR10 <outlier_cifar10>
  
Batch Processing with Seldon Core
---------------------------------

.. toctree::
   :titlesonly:

   Batch Processing with Argo Workflows and S3 / Minio <argo_workflows_batch>
   Batch Processing with Argo Workflows and HDFS <argo_workflows_hdfs_batch>
   Batch Processing with Kubeflow Pipelines <kubeflow_pipelines_batch>


MLOps: Scaling and Monitoring and Observability
-----------------------------------------------

.. toctree::
   :titlesonly:

   Autoscaling Example <autoscaling_example>
   KEDA Autoscaling example <keda>
   Request Payload Logging with ELK <payload_logging>
   Custom Metrics with Grafana & Prometheus <metrics>
   Distributed Tracing with Jaeger <tracing>
   CI / CD with Jenkins Classic <jenkins_classic>
   CI / CD with Jenkins X <jenkins_x>
   Replica control <scale>

Production Configurations and Integrations
------------------------------------------

.. toctree::
   :titlesonly:
  
   Example Helm Deployments <helm_examples>
   Max gRPC Message Size <max_grpc_msg_size>
   Configurable timeouts <timeouts>
   Deploy Multiple Seldon Core Operators <multiple_operators>
   Protocol Examples <protocol_examples>
   Custom Protobuf Data Example <customdata_example>
   Disruption Budgets Example <pdbs_example>

AB Tests and Progressive Rollouts
---------------------------------

.. toctree::
   :titlesonly:

   Istio AB Test <istio_canary>
   Ambassador AB Test <ambassador_canary>
   Seldon/Iter8 - Progressive AB Test with Single Seldon Deployment <iter8-single>
   Seldon/Iter8 - Progressive AB Test with Multiple Seldon Deployments <iter8-separate>



Complex Graph Examples
----------------------

.. toctree::
   :titlesonly:
  
   Chainer MNIST <chainer_mnist>
   Custom pre-processors with the V2 Protocol <transformers-v2-protocol>

Ingress
-------

.. toctree::
   :titlesonly:
  
   Ambassador Canary <ambassador_canary>
   Ambassador Shadow <ambassador_shadow>
   Ambassador Headers <ambassador_headers>
   Ambassador Custom Config <ambassador_custom>
   Istio Examples <istio_examples> 
   Istio Canary <istio_canary>

Infrastructure
--------------

.. toctree::
   :titlesonly:
  
   Patch Volumes for Version 1.2.0 Upgrade <patch_1_2>
   

Benchmarking and Load Tests
---------------------------

.. toctree::
   :titlesonly:
  
   Service Orchestrator Overhead <bench_svcOrch>
   Tensorflow Benchmark <bench_tensorflow>   
   Argo Workflows Benchmarking <vegeta_bench_argo_workflows>
   Python Serialization Cost Benchmark <python_serialization>
   KMP_AFFINITY Benchmarking Example <python_kmp_affinity>
   Kafka Payload Logging <kafka_logger>


Upgrading Examples
------------------

.. toctree::
   :titlesonly:

   Backwards Compatibility Tests <backwards_compatibility>
   RClone Storage Initializer - testing new secret format <rclone-upgrade>
   RClone Storage Initializer - upgrading your cluster (AWS S3 / MinIO) <global-rclone-upgrade>
