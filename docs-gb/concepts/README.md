
In the context of machine learning and Seldon Core 2, concepts provide a framework for understanding key functionalities, architectures, and workflows within the system. Some of the Key concepts in Seldon Core 2 are:

* [Data-Centric MLOps](#data-centric-mlops)
* [Open Inference Protocol](#open-inference-protocol-in-seldon-core-2)
* [Components](#components-in-seldon-core-2)
* [Pipelines](#pipelines)
* [Severs](#servers)
* [Experiments](#experiments)

## Data-Centric MLOps

Data-centricity is an approach that prioritizes the management, integrity, and flow of data at the core of machine learning deployment. Rather than focusing solely on models, this approach ensures that data quality, consistency, and adaptability drive successful ML operations. In Seldon Core 2, data-centricity is embedded in every stage of the inference workflow, enabling scalable, real-time, and standardized model serving.

| **Key Principle**          | **Description** |
|---------------------------|---------------|
| **Flexible Workflows**      | Seldon Core 2 supports *adaptable and scalable data pathways*, accommodating various **use cases and experiments**. This ensures ML pipelines remain *agile*, allowing you to evolve the **data processing strategies** as requirements change. |
| **Real-Time Data Streaming** | Integrated *data streaming capabilities* allow you to *view, store, manage, and process data in real time*. This enhances *responsiveness and decision-making*, ensuring models work with *the most up-to-date data* for accurate predictions. |
| **Standardized Processing**  | Seldon Core 2 promotes *reusable and consistent* data transformation and routing mechanisms. Standardized processing ensures *data integrity and uniformity* across applications, reducing errors and inconsistencies. |
| **Comprehensive Monitoring** | *Detailed metrics and logs* provide real-time visibility into *data integrity, transformations, and flow*. This enables *effective oversight and maintenance*, allowing teams to detect *anomalies, drifts, or inefficiencies* early. |

Why Data-Centricity Matters?

By adopting a data-centric approach, Seldon Core 2 enables:
* More reliable and high-quality predictions by ensuring clean, well-structured data.
* Scalable and future-proof ML deployments through standardized data management.
* Efficient monitoring and maintenance, reducing risks related to model drift and inconsistencies.

With data-centricity as a core principle, Seldon Core 2 ensures end-to-end control over ML workflows, enabling you to maximize model performance and reliability in production environments.

## Open Inference Protocol in Seldon Core 2
The Open Inference Protocol (OIP) ensures standardized communication between inference servers and clients, enabling interoperability and flexibility across different model-serving runtimes in Seldon Core 2.

To be compliant, an inference server must implement these three key APIs:
| **API**            | **Function** |
|-------------------|-------------|
| **Health API**    | Ensures the server is *operational and available* for inference requests. |
| **Metadata API**  | Provides *essential details* about deployed models, including *capabilities and configurations*. |
| **V2 Inference Protocol API** | Facilitates *standardized request and response handling* for model inference. |

**Protocol Compatibility**
* Flexible API Support: A compliant server can implement either HTTP/REST or gRPC APIs, or both.
* Runtime Compatibility:Users should refer to the model serving runtime table or the protocolVersion field in runtime YAML to confirm V2 Inference Protocol support for their serving runtime.

**Case Sensitivity and Extensions**
All strings are case-sensitive across API descriptions. V2 Inference Protocol includes an extension mechanism, though specific extensions are defined separately.

By adopting the V2 Inference Protocol, Seldon Core 2 ensures standardized, scalable, and flexible model serving across diverse deployment environments.

## Components in Seldon Core 2
Components are the building blocks of an inference graph, processing data at various stages of the ML inference pipeline. They provide reusable, standardized interfaces, making it easier to maintain and update workflows without disrupting the entire system. Components include ML models, data processors, routers, and supplementary services.

Types of Components

| **Component Type**                     | **Description** |
|-----------------------------------------|---------------|
| **Sources**                             | Starting points of an inference graph that receive and validate incoming data before processing. |
| **Data Processors**                     | Transform, filter, or aggregate data to ensure consistent, repeatable pre-processing. |
| **Data Routers**                         | Dynamically route data to different paths based on predefined rules for A/B testing, experimentation, or conditional logic. |
| **Models**                               | Perform inference tasks, including classification, regression, and Large Language Models (LLMs), hosted internally or via external APIs. |
| **Supplementary Data Services**         | External services like vector databases that enable models to access embeddings and extended functionality. |
| **Drift/Outlier Detectors & Explainers** | Monitor model predictions for drift, anomalies, and explainability insights, ensuring transparency and performance tracking. |
| **Sinks**                                | Endpoints of an inference graph that deliver results to external consumers while maintaining a stable interface. |

By modularizing inference workflows, components allow you to scale, experiment, and optimize ML deployments efficiently while ensuring data consistency and reliability. 


## Pipelines

In a model serving platform, a pipeline is an automated sequence of steps that manages the deployment, execution, and monitoring of machine learning models. Pipelines ensure that models are efficiently served, dynamically updated, and continuously monitored for performance and reliability in production environments.

| **Stage**                   | **Description** |
|-----------------------------|---------------|
| **Request Ingestion**        | Receives inference requests from applications, APIs, or streaming sources. |
| **Preprocessing**            | Transforms input data for example, tokenization, or normalization before passing it to the model. |
| **Model Selection & Routing**| Directs requests to the appropriate model based on rules, versions, or A/B testing. |
| **Inference Execution**      | Runs predictions using the deployed model. |
| **Postprocessing**           | Converts model outputs into a consumable format such as confidence scores, structured responses. |
| **Response Delivery**        | Returns inference results to the requesting application or system. |
| **Monitoring & Logging**     | Tracks model performance, latency, and accuracy; detects drift and triggers alerts if needed. |

## Servers
In Seldon Core 2, servers are responsible for hosting and serving machine learning models, handling inference requests, and ensuring scalability, efficiency, and observability in production. Seldon Core 2 supports multiple inference servers, including MLServer and NVIDIA Triton, enabling flexible and optimized model deployments.

| **Server**        | **Description** | **Best Suited For** |
|-------------------|---------------|---------------------|
| **MLServer**     | A lightweight, extensible inference server designed to work with multiple ML frameworks, including *scikit-learn, XGBoost, TensorFlow, and PyTorch*. It supports *custom Python models* and integrates well with *MLOps workflows*. | *General-purpose model serving*, custom model wrappers, multi-framework support. |
| **NVIDIA Triton** | A high-performance inference server optimized for *GPU and CPU acceleration*, supporting deep learning models across frameworks like *TensorFlow, PyTorch, and ONNX*. Triton enables *multi-model and ensemble model inference*, making it ideal for *scalable AI workloads*. | *High-throughput deep learning inference*, multi-model deployments, GPU-accelerated workloads. |

With MLServer and Triton, Seldon Core 2 provides a powerful, efficient, and flexible model-serving platform for production-scale AI applications. 

## Experiments

In Seldon Core 2, experiments enable controlled A/B testing, model comparisons, and performance evaluations by defining an HTTP traffic split between different models or inference pipelines. This allows organizations to test multiple versions of a model in production while managing risk and ensuring continuous improvements.

| **Experiment Type**  | **Description** |
|----------------------|----------------|
| **Traffic Splitting** | Distributes inference requests across different models or pipelines based on predefined percentage splits. This enables *A/B testing* and comparison of multiple model versions. |
| **Mirror Testing** | Sends a *percentage of the traffic* to a *mirror model or pipeline* without affecting the response returned to users. This allows evaluation of new models without impacting production workflows. |

Some of the advantages of using Experiments:
* A/B Testing & Model comparison: Compare different models under real-world conditions without full deployment.
* Risk-Free model validation: Test a new model or pipeline in parallel without affecting live predictions.
* Performance & Drift monitoring: Assess latency, accuracy, and reliability before a full rollout.
* Continuous improvement: Make data-driven deployment decisions based on real-time model performance.

