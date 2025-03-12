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

By modularizing inference workflows, components allow users to scale, experiment, and optimize ML deployments efficiently while ensuring data consistency and reliability. 