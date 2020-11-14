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
   Deploy a XGBoost Model Binary <../servers/xgboost.md>
   Deploy Pre-packaged Model Server with Cluster's MinIO <minio-sklearn>

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

Specialised Framework Examples
------------------------------

.. toctree::
   :titlesonly:

   NVIDIA TensorRT MNIST <tensorrt>
   OpenVINO ImageNet <openvino>
   OpenVINO ImageNet Ensemble <openvino_ensemble>


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

Advanced Machine Learning Insights
----------------------------------

.. toctree::
   :titlesonly:

   Real time monitoring of statistical metrics <feedback_reward_custom_metrics>
   Tabular, Text and Image Model Explainers <explainer_examples>
   Outlier Detection on CIFAR10 <outlier_cifar10>
  
Batch Processing with Seldon Core
---------------------------------

.. toctree::
   :titlesonly:

   Batch Processing with Argo Workflows <argo_workflows_batch>
   Batch Processing with Kubeflow Pipelines <kubeflow_pipelines_batch>


MLOps: Scaling and Monitoring and Observability
-----------------------------------------------

.. toctree::
   :titlesonly:

   Autoscaling Example <autoscaling_example>    
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
   REST timeouts <rest_timeouts>
   Deploy Multiple Seldon Core Operators <multiple_operators>
   Protocol Examples <protocol_examples>
   Custom Protobuf Data Example <customdata_example>
   Disruption Budgets Example <disruption_budgets_example>

Complex Graph Examples
----------------------

.. toctree::
   :titlesonly:
  
   Chainer MNIST <chainer_mnist>

Ingress
-------

.. toctree::
   :titlesonly:
  
   Ambassador Canary <ambassador_canary>
   Ambassador Shadow <ambassador_shadow>
   Ambassador Headers <ambassador_headers>
   Ambassador Custom Config <ambassador_custom>
   Istio Canary <istio_canary>
   Istio Examples <istio_examples>   

Infrastructure
--------------

.. toctree::
   :titlesonly:
  
   Patch Volumes for Version 1.2.0 Upgrade <patch_1_2>
   

Benchmarking and Load Tests
---------------------------

.. toctree::
   :titlesonly:
  
   Service Orchestrator <bench_svcOrch>
   Tensorflow <bench_tensorflow>   
   Argo Workflows Benchmarking <vegeta_bench_argo_workflows>   

