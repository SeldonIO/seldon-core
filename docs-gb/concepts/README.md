---
description: >-
  Learn how Seldon Core is different from a centralized orchestrator to a data
  flow architecture, enabling more efficient ML model deployment through stream
  processing and improved data handling.
---

# Concepts

In the context of machine learning and Seldon Core 2, concepts provide a framework for understanding key functionalities, architectures, and workflows within the system. Some of the key concepts in Seldon Core 2 are:

* [Data-Centric MLOps](./#data-centric-mlops)
* [Open Inference Protocol](./#open-inference-protocol-in-seldon-core-2)
* [Components](./#components-in-seldon-core-2)
* [Pipelines](./#pipelines)
* [Servers](./#servers)
* [Experiments](./#experiments)

### Data-Centric MLOps

Data-centricity is an approach that puts the management, integrity, and flow of data at the core of machine learning deployment. Rather than focusing solely on models, this approach ensures that data quality, consistency, and adaptability drive successful ML operations. In Seldon Core 2, data-centricity is embedded in every stage of the inference workflow.

![Seldon Core 2 Data-Centric Approach](../.gitbook/assets/data-centric-approach.png)

| **Key Principle**            | **Description**                                                                                                                                                                                                                        |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Flexible Workflows**       | Core 2 supports adaptable and scalable data pathways, accommodating various _use cases and experiments_. This ensures ML pipelines remain agile, allowing you to evolve the inference logic as requirements change..                   |
| **Real-Time Data Streaming** | Integrated data streaming capabilities allow you to view, store, manage, and process data in real time. This enhances responsiveness and decision-making, ensuring models work with the most up-to-date data for accurate predictions. |
| **Standardized Processing**  | Core 2 promotes reusable and consistent data transformation and routing mechanisms. Standardized processing ensures data integrity and uniformity across applications, reducing errors and inconsistencies.                            |
| **Comprehensive Monitoring** | Detailed metrics and logs provide real-time visibility into data integrity, transformations, and flow. This enables effective oversight and maintenance, allowing teams to detect anomalies, drifts, or inefficiencies early.          |

Why Data-Centricity Matters?

By adopting a data-centric approach, Seldon Core 2 enables:

* More reliable and high-quality predictions by ensuring clean, well-structured data.
* Scalable and future-proof ML deployments through standardized data management.
* Efficient monitoring and maintenance, reducing risks related to model drift and inconsistencies.

With data-centricity as a core principle, Seldon Core 2 ensures end-to-end control over ML workflows, enabling you to maximize model performance and reliability in production environments.

### Open Inference Protocol in Seldon Core 2

The **Open Inference Protocol (OIP)** defines a standard way for inference servers and clients to communicate in Seldon Core 2. Its goal is to enable **interoperability**, **flexibility**, and **consistency** across different model-serving runtimes. It exposes the Health API, Metadata API, and the Inference API.

Some of the features of Open Inference Protocol includes:

* **Transport Agnostic**: Servers can implement HTTP/REST or gRPC protocols.
* **Runtime Awareness**: Use the `protocolVersion` field in your runtime YAML or consult the supported runtimes table to verify compatibility.

By adopting OIP, Seldon Core 2 promotes a  **consistent** experience across a diverse set of model deployments.

### Components in Seldon Core 2

Components are the building blocks of an inference graph, processing data at various stages of the ML inference pipeline. They provide reusable, standardized interfaces, making it easier to maintain and update workflows without disrupting the entire system. Components include ML models, data processors, routers, and supplementary services.

Types of Components

| **Component Type**                       | **Description**                                                                                                                          |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **Data Processors**                      | Transform, filter, or aggregate data to ensure consistent, repeatable pre-processing.                                                    |
| **Data Routers**                         | Dynamically route data to different paths based on predefined rules for A/B testing, experimentation, or conditional logic.              |
| **Models**                               | Perform inference tasks, including classification, regression, and Large Language Models (LLMs), hosted internally or via external APIs. |
| **Supplementary Data Services**          | External services like vector databases that enable models to access embeddings and extended functionality.                              |
| **Drift/Outlier Detectors & Explainers** | Monitor model predictions for drift, anomalies, and explainability insights, ensuring transparency and performance tracking.             |

By modularizing inference workflows, components allow you to scale, experiment, and optimize ML deployments efficiently while ensuring data consistency and reliability.

### Pipelines

In a model serving platform like Seldon Core 2, a **pipeline** refers to an orchestrated sequence of models and components that work together to serve more complex AI applications. These pipelines allow you to connect models, routers, transformers, and other inference components, with **Kafka used to stream data** between them. Each component in the pipeline is **modular and independently managed**—meaning it can be scaled, configured, and updated separately based on its specific input and output requirements.

This use of "pipeline" is distinct from how the term is used in MLOps for **CI/CD pipelines**, which automate workflows for building, testing, and deploying models. In contrast, Core 2 pipelines operate at runtime and focus on the **live composition and orchestration** of inference systems in production.

### Servers

In Core 2, servers are responsible for hosting and serving machine learning models, handling inference requests, and ensuring scalability, efficiency, and observability in production. Core 2 supports multiple inference servers, including MLServer, and NVIDIA Triton Inference Server, enabling flexible and optimized model deployments.

**MLServer**: A lightweight, extensible inference server designed to work with multiple ML frameworks, including _scikit-learn, XGBoost, TensorFlow, and PyTorch_. It supports _custom Python models_ and integrates well with _MLOps workflows_. It is built for \*General-purpose model serving, custom model wrappers, multi-framework support.

### Experiments

In Seldon Core 2, experiments enable controlled A/B testing, model comparisons, and performance evaluations by defining an HTTP traffic split between different models or inference pipelines. This allows organizations to test multiple versions of a model in production while managing risk and ensuring continuous improvements.

| **Experiment Type**   | **Description**                                                                                                                                                                                                                              |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Traffic Splitting** | Distributes inference requests across different models or pipelines based on predefined percentage splits. This enables A/B testing and comparison of multiple model versions. For example, Canary deployment of the models.                 |
| **Mirror Testing**    | Sends a percentage of the traffic to a mirror model or pipeline without affecting the response returned to users. This allows evaluation of new models without impacting production workflows. For example, Shadow deployment of the models. |

Some of the advantages of using Experiments:

* A/B Testing & Model comparison: Compare different models under real-world conditions without full deployment.
* Risk-Free model validation: Test a new model or pipeline in parallel without affecting live predictions.
* Performance monitoring: Assess latency, accuracy, and reliability before a full rollout.
* Continuous improvement: Make data-driven deployment decisions based on real-time model performance.
